package subscription

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
)

func Init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)
	domainURL := os.Getenv("DOMAIN")
	if !strings.HasPrefix(domainURL, "http") {
		domainURL = "https://" + domainURL
	}
	priceID := os.Getenv("STRIPE_PRICE_ID")

	params := &stripe.CheckoutSessionParams{
		CustomerEmail:      stripe.String(userEmail),
		ClientReferenceID:  stripe.String(strconv.Itoa(int(userID))),
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:         stripe.String(domainURL + "/settings?success=true"),
		CancelURL:          stripe.String(domainURL + "/settings?canceled=true"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, err := checkoutsession.New(params)
	if err != nil {
		log.Printf("checkoutsession.New: %v", err)
		http.Error(w, "Error creating checkout session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	domainURL := os.Getenv("DOMAIN")
	if !strings.HasPrefix(domainURL, "http") {
		domainURL = "https://" + domainURL
	}

	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if user.StripeCustomerID == "" {
		http.Error(w, "No billing account found", http.StatusBadRequest)
		return
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String(domainURL + "/settings"),
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Error creating portal session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Verify the signature
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	
	// Use ConstructEventWithOptions to ignore API version mismatch
	// This is necessary because the Stripe Go SDK version might lag behind the API version used by the webhook
	event, err := webhook.ConstructEventWithOptions(payload, r.Header.Get("Stripe-Signature"), endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handleCheckoutSessionCompleted(&session)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handleSubscriptionUpdated(&subscription)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handleSubscriptionDeleted(&subscription)
	}

	w.WriteHeader(http.StatusOK)
}

func handleCheckoutSessionCompleted(session *stripe.CheckoutSession) {
	userIDStr := session.ClientReferenceID
	userID, _ := strconv.Atoi(userIDStr)
	customerID := session.Customer.ID
	subscriptionID := session.Subscription.ID

	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		log.Printf("User not found for ID %d: %v", userID, result.Error)
		return
	}

	user.StripeCustomerID = customerID
	user.SubscriptionID = subscriptionID
	user.SubscriptionTier = "pro"
	user.SubscriptionStatus = "active"

	database.DB.Save(&user)
	log.Printf("User %d upgraded to Pro", userID)
}

func handleSubscriptionUpdated(subscription *stripe.Subscription) {
	customerID := subscription.Customer.ID
	status := string(subscription.Status)

	var user domain.User
	if result := database.DB.Where("stripe_customer_id = ?", customerID).First(&user); result.Error != nil {
		log.Printf("User not found for customer ID %s: %v", customerID, result.Error)
		return
	}

	user.SubscriptionStatus = status
	if status == "active" || status == "trialing" {
		user.SubscriptionTier = "pro"
	} else {
		// past_due, canceled, unpaid
		// We might want to keep them as pro for a grace period, but for simplicity:
		if status == "canceled" || status == "unpaid" {
			user.SubscriptionTier = "free"
		}
	}
	database.DB.Save(&user)
}

func handleSubscriptionDeleted(subscription *stripe.Subscription) {
	customerID := subscription.Customer.ID

	var user domain.User
	if result := database.DB.Where("stripe_customer_id = ?", customerID).First(&user); result.Error != nil {
		log.Printf("User not found for customer ID %s: %v", customerID, result.Error)
		return
	}

	user.SubscriptionTier = "free"
	user.SubscriptionStatus = "canceled"
	user.SubscriptionID = ""
	database.DB.Save(&user)
	log.Printf("User %d downgraded to Free", user.ID)
}

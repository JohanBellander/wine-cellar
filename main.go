package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"wine-cellar/internal/features/auth"
	"wine-cellar/internal/features/reviews/add"
	"wine-cellar/internal/features/settings"
	"wine-cellar/internal/features/subscription"
	addWine "wine-cellar/internal/features/wines/add"
	deleteWine "wine-cellar/internal/features/wines/delete"
	"wine-cellar/internal/features/wines/details"
	"wine-cellar/internal/features/wines/edit"
	"wine-cellar/internal/features/wines/list"
	"wine-cellar/internal/features/wines/update"
	"wine-cellar/internal/shared/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Stripe
	subscription.Init()

	database.InitDB()
	database.Seed(database.DB)

	http.HandleFunc("/signup", auth.SignupHandler)
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/logout", auth.LogoutHandler)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/add", auth.Middleware(addWine.Handler))
	http.HandleFunc("/details/", auth.Middleware(details.Handler))
	http.HandleFunc("/edit/", auth.Middleware(edit.Handler))
	http.HandleFunc("/update-quantity", auth.Middleware(update.QuantityHandler))
	http.HandleFunc("/add-review", auth.Middleware(add.Handler))
	http.HandleFunc("/settings", auth.Middleware(settings.Handler))
	http.HandleFunc("/delete", auth.Middleware(deleteWine.Handler))
	http.HandleFunc("/create-checkout-session", auth.Middleware(subscription.CreateCheckoutSession))
	http.HandleFunc("/create-portal-session", auth.Middleware(subscription.CreatePortalSession))
	http.HandleFunc("/webhook/stripe", subscription.WebhookHandler)
	http.HandleFunc("/health", healthHandler)

	// Serve static files if we had any, but we are using CDNs mostly.
	// If we had local images, we would serve them here.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Define a FuncMap with the safeURL function
var funcMap = template.FuncMap{
	"safeURL": func(s string) template.URL {
		return template.URL(s)
	},
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	userID, email, authenticated := auth.GetSessionUser(r)
	if authenticated {
		// User is authenticated
		// We need to inject user_id into context like AuthMiddleware does
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "email", email)

		list.Handler(w, r.WithContext(ctx))
		return
	}

	// User is not authenticated
	landingHandler(w, r)
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("landing.html").Funcs(funcMap).ParseFiles("templates/landing.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		LoggedIn  bool
		UserEmail string
	}{
		LoggedIn:  false,
		UserEmail: "",
	}

	tmpl.Execute(w, data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}



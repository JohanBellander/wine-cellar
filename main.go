package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"wine-cellar/internal/features/auth"
	"wine-cellar/internal/features/reviews/add"
	deleteReview "wine-cellar/internal/features/reviews/delete"
	editReview "wine-cellar/internal/features/reviews/edit"
	"wine-cellar/internal/features/settings"
	"wine-cellar/internal/features/subscription"
	addTastingNote "wine-cellar/internal/features/tastingnotes/add"
	deleteTastingNote "wine-cellar/internal/features/tastingnotes/delete"
	editTastingNote "wine-cellar/internal/features/tastingnotes/edit"
	addWine "wine-cellar/internal/features/wines/add"
	deleteWine "wine-cellar/internal/features/wines/delete"
	"wine-cellar/internal/features/wines/details"
	"wine-cellar/internal/features/wines/edit"
	"wine-cellar/internal/features/wines/list"
	"wine-cellar/internal/features/wines/update"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/storage"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Stripe
	subscription.Init()

	// Initialize Auth (Session Store)
	auth.Init()

	// Initialize R2 Storage
	storage.Init()

	database.InitDB()
	database.Seed(database.DB)

	mux := http.NewServeMux()

	mux.HandleFunc("/signup", auth.SignupHandler)
	mux.HandleFunc("/login", auth.LoginHandler)
	mux.HandleFunc("/logout", auth.LogoutHandler)

	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/privacy.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, email, authenticated := auth.GetSessionUser(r)
		data := struct {
			LoggedIn  bool
			UserEmail string
		}{
			LoggedIn:  authenticated,
			UserEmail: email,
		}

		tmpl.Execute(w, data)
	})

	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/terms.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, email, authenticated := auth.GetSessionUser(r)
		data := struct {
			LoggedIn  bool
			UserEmail string
		}{
			LoggedIn:  authenticated,
			UserEmail: email,
		}

		tmpl.Execute(w, data)
	})

	mux.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		_, email, authenticated := auth.GetSessionUser(r)
		data := struct {
			LoggedIn  bool
			UserEmail string
		}{
			LoggedIn:  authenticated,
			UserEmail: email,
		}

		tmpl, err := template.ParseFiles("templates/contact.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, data)
	})

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/add", auth.Middleware(addWine.Handler))
	mux.HandleFunc("/details/", auth.Middleware(details.Handler))
	mux.HandleFunc("/edit/", auth.Middleware(edit.Handler))
	mux.HandleFunc("/update-quantity", auth.Middleware(update.QuantityHandler))
	mux.HandleFunc("/add-review", auth.Middleware(add.Handler))
	mux.HandleFunc("/delete-review", auth.Middleware(deleteReview.Handler))
	mux.HandleFunc("/edit-review", auth.Middleware(editReview.Handler))
	mux.HandleFunc("/add-tasting-note", auth.Middleware(addTastingNote.Handler))
	mux.HandleFunc("/delete-tasting-note", auth.Middleware(deleteTastingNote.Handler))
	mux.HandleFunc("/edit-tasting-note", auth.Middleware(editTastingNote.Handler))
	mux.HandleFunc("/settings", auth.Middleware(settings.Handler))
	mux.HandleFunc("/export", auth.Middleware(settings.ExportHandler))
	mux.HandleFunc("/delete-account", auth.Middleware(settings.DeleteAccountHandler))
	mux.HandleFunc("/delete", auth.Middleware(deleteWine.Handler))
	mux.HandleFunc("/delete-photo", auth.Middleware(edit.DeletePhotoHandler))
	mux.HandleFunc("/create-checkout-session", auth.Middleware(subscription.CreateCheckoutSession))
	mux.HandleFunc("/create-portal-session", auth.Middleware(subscription.CreatePortalSession))
	mux.HandleFunc("/webhook/stripe", subscription.WebhookHandler)
	mux.HandleFunc("/health", healthHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// CSRF Protection
	csrfKey := os.Getenv("CSRF_AUTH_KEY")
	if len(csrfKey) != 32 {
		if csrfKey == "" {
			log.Println("Warning: CSRF_AUTH_KEY is not set. Using default static key (NOT SAFE FOR PRODUCTION if multiple instances).")
			csrfKey = "01234567890123456789012345678901"
		} else {
			log.Printf("Warning: CSRF_AUTH_KEY is not 32 bytes long (got %d bytes). Padding/Truncating to 32 bytes.", len(csrfKey))
			if len(csrfKey) < 32 {
				csrfKey = csrfKey + "00000000000000000000000000000000"
			}
			csrfKey = csrfKey[:32]
		}
	}
	
	// Log the key fingerprint to verify consistency across restarts
	log.Printf("CSRF Key Fingerprint: %s...%s", csrfKey[:4], csrfKey[28:])

	domain := os.Getenv("DOMAIN")
	// Clean domain to ensure we build valid origins
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimRight(domain, "/")

	trustedOrigins := []string{
		"localhost:8080",
		"127.0.0.1:8080",
		domain,
	}
	if domain != "" {
		trustedOrigins = append(trustedOrigins, "www."+domain)
		trustedOrigins = append(trustedOrigins, "https://"+domain)
		trustedOrigins = append(trustedOrigins, "https://www."+domain)
		trustedOrigins = append(trustedOrigins, "http://"+domain)
		trustedOrigins = append(trustedOrigins, "http://www."+domain)
	}

	isProd := os.Getenv("GO_ENV") == "production"
	log.Printf("CSRF Config: Secure=%v, TrustedOrigins=%v", isProd, trustedOrigins)

	csrfMiddleware := csrf.Protect(
		[]byte(csrfKey),
		// Temporarily disable Secure flag to rule out proxy/protocol mismatch issues
		csrf.Secure(false), 
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.TrustedOrigins(trustedOrigins),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("CSRF Error: %v", csrf.FailureReason(r))
			log.Printf("Request Info: Protocol=%s, Host=%s, Method=%s, URL=%s", r.Proto, r.Host, r.Method, r.URL.String())
			// Check for CSRF cookie
			if cookie, err := r.Cookie("_gorilla_csrf"); err == nil {
				log.Printf("CSRF Cookie present: len=%d", len(cookie.Value))
			} else {
				log.Printf("CSRF Cookie missing: %v", err)
			}
			http.Error(w, "Forbidden - CSRF token invalid", http.StatusForbidden)
		})),
	)

	fmt.Printf("Server started at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, csrfMiddleware(mux)))
}

// Define a FuncMap with the safeURL function
var funcMap = template.FuncMap{
	"safeURL": func(s string) template.URL {
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "mailto:") || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "data:") {
			return template.URL(s)
		}
		return template.URL("")
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
	tmpl, err := template.New("landing.html").Funcs(funcMap).ParseFiles("templates/landing.html", "templates/header.html", "templates/footer.html")
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



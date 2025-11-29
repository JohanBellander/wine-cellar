package auth

import (
	"html/template"
	"net/http"
	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"

	"github.com/gorilla/csrf"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("internal/features/auth/signup.html", "templates/footer.html", "templates/analytics.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Check if tier parameter is passed (for Connoisseur signup flow)
		tier := r.URL.Query().Get("tier")
		
		data := map[string]interface{}{
			"CSRFField": csrf.TemplateField(r),
			"Tier":      tier,
		}
		
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		tier := r.FormValue("tier")

		hashedPassword, err := HashPassword(password)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		user := domain.User{Email: email, PasswordHash: hashedPassword}
		result := database.DB.Create(&user)
		if result.Error != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}

		// If signing up for pro tier, auto-login and redirect to Stripe checkout
		if tier == "pro" {
			// Auto-login the user
			session, _ := store.Get(r, "session-name")
			session.Values["authenticated"] = true
			session.Values["user_id"] = user.ID
			session.Values["email"] = user.Email
			session.Save(r, w)

			// Redirect to checkout
			http.Redirect(w, r, "/create-checkout-session", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("internal/features/auth/login.html", "templates/footer.html", "templates/analytics.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		data := map[string]interface{}{
			"CSRFField": csrf.TemplateField(r),
		}

		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user domain.User
		result := database.DB.Where("email = ?", email).First(&user)
		if result.Error != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !CheckPasswordHash(password, user.PasswordHash) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Check if password needs re-hashing (migration from cost 14 to 10)
		// This ensures existing users get faster logins next time
		if NeedsRehash(user.PasswordHash) {
			newHash, err := HashPassword(password)
			if err == nil {
				user.PasswordHash = newHash
				database.DB.Save(&user)
			}
		}

		session, _ := store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["user_id"] = user.ID
		session.Values["email"] = user.Email
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Values["email"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

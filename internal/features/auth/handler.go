package auth

import (
	"html/template"
	"net/http"
	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("internal/features/auth/signup.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

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

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("internal/features/auth/login.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
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

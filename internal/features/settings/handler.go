package settings

import (
	"html/template"
	"net/http"
	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	if r.Method == http.MethodGet {
		var user domain.User
		if result := database.DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		// Note: Path to templates is relative to the project root
		tmpl, err := template.New("settings.html").Funcs(ui.FuncMap).ParseFiles(
			"internal/features/settings/settings.html",
			"templates/header.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			User      domain.User
			LoggedIn  bool
			UserEmail string
		}{
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
		}

		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		currency := r.FormValue("currency")
		
		var user domain.User
		if result := database.DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		user.Currency = currency
		database.DB.Save(&user)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

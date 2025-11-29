package details

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}
	idStr := pathParts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var wine domain.Wine
	result := database.DB.Preload("Reviews").Preload("TastingNotes").Where("user_id = ?", userID).First(&wine, id)
	if result.Error != nil {
		http.NotFound(w, r)
		return
	}

	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("details.html").Funcs(ui.FuncMap).ParseFiles("internal/features/wines/details/details.html", "templates/header.html", "templates/footer.html", "templates/analytics.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Wine      domain.Wine
		User      domain.User
		LoggedIn  bool
		UserEmail string
		CSRFField template.HTML
	}{
		Wine:      wine,
		User:      user,
		LoggedIn:  true,
		UserEmail: userEmail,
		CSRFField: csrf.TemplateField(r),
	}

	tmpl.Execute(w, data)
}

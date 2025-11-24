package edit

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	if r.Method == http.MethodGet {
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
		if result := database.DB.First(&wine, id); result.Error != nil {
			http.NotFound(w, r)
			return
		}

		var user domain.User
		if result := database.DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("edit.html").Funcs(ui.FuncMap).ParseFiles("internal/features/wines/edit/edit.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			Wine      domain.Wine
			User      domain.User
			LoggedIn  bool
			UserEmail string
		}{
			Wine:      wine,
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
		}
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		// Parse multipart form, 10MB max
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var wine domain.Wine
		if result := database.DB.First(&wine, id); result.Error != nil {
			http.NotFound(w, r)
			return
		}

		vintage, _ := strconv.Atoi(r.FormValue("vintage"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

		wine.Name = r.FormValue("name")
		wine.Producer = r.FormValue("producer")
		wine.Vintage = vintage
		wine.IsNonVintage = r.FormValue("is_non_vintage") == "on"
		wine.Grape = r.FormValue("grape")
		wine.Country = r.FormValue("country")
		wine.Region = r.FormValue("region")
		wine.Category = r.FormValue("category")
		wine.SubCategory = r.FormValue("sub_category")
		wine.Quantity = quantity
		wine.Price = price
		wine.DrinkingWindow = r.FormValue("drinking_window")
		
		// Handle image upload
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			// Read file content
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				// Convert to base64
				base64Str := base64.StdEncoding.EncodeToString(fileBytes)
				// Determine mime type (simple check)
				mimeType := http.DetectContentType(fileBytes)
				wine.ImageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
			}
		}

		database.DB.Save(&wine)

		http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
	}
}

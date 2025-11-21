package add

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	var wineCount int64
	database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Count(&wineCount)

	if r.Method == http.MethodGet {
		tmpl, err := template.New("add.html").Funcs(ui.FuncMap).ParseFiles("internal/features/wines/add/add.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			Wine         domain.Wine
			User         domain.User
			LoggedIn     bool
			UserEmail    string
			WineCount    int64
			IsFreeTier   bool
			LimitReached bool
		}{
			Wine:         domain.Wine{},
			User:         user,
			LoggedIn:     true,
			UserEmail:    userEmail,
			WineCount:    wineCount,
			IsFreeTier:   user.SubscriptionTier == "free",
			LimitReached: user.SubscriptionTier == "free" && wineCount >= 10,
		}
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		if user.SubscriptionTier == "free" && wineCount >= 10 {
			http.Error(w, "Free tier limit reached. Please upgrade to add more wines.", http.StatusForbidden)
			return
		}

		// Parse multipart form, 10MB max
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		vintage, _ := strconv.Atoi(r.FormValue("vintage"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

		imageURL := "https://via.placeholder.com/150"
		
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
				imageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
			}
		}

		newWine := domain.Wine{
			Name:           r.FormValue("name"),
			Producer:       r.FormValue("producer"),
			Vintage:        vintage,
			Grape:          r.FormValue("grape"),
			Country:        r.FormValue("country"),
			Region:         r.FormValue("region"),
			Quantity:       quantity,
			Price:          price,
			DrinkingWindow: r.FormValue("drinking_window"),
			ImageURL:       imageURL,
			UserID:         userID,
		}

		database.DB.Create(&newWine)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

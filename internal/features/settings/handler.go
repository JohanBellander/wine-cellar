package settings

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)
	isDev := os.Getenv("APP_ENV") == "dev"

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
			"templates/footer.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			User      domain.User
			LoggedIn  bool
			UserEmail string
			IsDev     bool
		}{
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
			IsDev:     isDev,
		}

		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		// Debug: Set subscription tier
		if isDev {
			debugTier := r.FormValue("debug_tier")
			if debugTier != "" {
				var user domain.User
				if result := database.DB.First(&user, userID); result.Error != nil {
					http.Error(w, "User not found", http.StatusInternalServerError)
					return
				}
				user.SubscriptionTier = debugTier
				database.DB.Save(&user)
				http.Redirect(w, r, "/settings", http.StatusSeeOther)
				return
			}
		}

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

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	var wines []domain.Wine
	if result := database.DB.Where("user_id = ?", userID).Find(&wines); result.Error != nil {
		http.Error(w, "Error fetching wines", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=wines.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"Name", "Producer", "Vintage", "Grape", "Country", "Region", "Quantity", "Price", "Location", "Rating", "Notes"}
	if err := writer.Write(header); err != nil {
		http.Error(w, "Error writing CSV header", http.StatusInternalServerError)
		return
	}

	// Write data
	for _, wine := range wines {
		record := []string{
			wine.Name,
			wine.Producer,
			strconv.Itoa(wine.Vintage),
			wine.Grape,
			wine.Country,
			wine.Region,
			strconv.Itoa(wine.Quantity),
			fmt.Sprintf("%.2f", wine.Price),
			wine.Location,
			wine.Rating,
			wine.Notes,
		}
		if err := writer.Write(record); err != nil {
			http.Error(w, "Error writing CSV record", http.StatusInternalServerError)
			return
		}
	}
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// Start transaction
	tx := database.DB.Begin()

	// Delete all wines belonging to user
	if err := tx.Where("user_id = ?", userID).Delete(&domain.Wine{}).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Could not delete wines", http.StatusInternalServerError)
		return
	}

	// Delete user
	if err := tx.Delete(&domain.User{}, userID).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Could not delete user", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	// Clear session (redirect to logout handler or do it here)
	// For simplicity, we'll just redirect to logout which handles session clearing
	http.Redirect(w, r, "/logout", http.StatusSeeOther)
}

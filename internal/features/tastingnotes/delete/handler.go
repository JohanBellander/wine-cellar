package delete

import (
	"net/http"
	"strconv"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
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

	// Check subscription tier
	userID := r.Context().Value("user_id").(uint)
	var user domain.User
	if result := database.DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if user.SubscriptionTier != "pro" {
		http.Error(w, "Tasting notes are only available for Pro users", http.StatusForbidden)
		return
	}

	// Verify ownership through the wine
	var note domain.TastingNote
	if result := database.DB.First(&note, id); result.Error != nil {
		http.Error(w, "Tasting note not found", http.StatusNotFound)
		return
	}

	var wine domain.Wine
	if result := database.DB.First(&wine, note.WineID); result.Error != nil {
		http.Error(w, "Wine not found", http.StatusNotFound)
		return
	}

	if wine.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete the tasting note
	if result := database.DB.Delete(&note); result.Error != nil {
		http.Error(w, "Error deleting tasting note", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/details/"+strconv.Itoa(int(wine.ID)), http.StatusSeeOther)
}

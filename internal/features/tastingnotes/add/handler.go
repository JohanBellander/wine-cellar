package add

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
		http.Error(w, "Tasting notes are only available for Connoisseur users", http.StatusForbidden)
		return
	}

	note := r.FormValue("note")

	// Simple validation
	if note == "" {
		http.Error(w, "Note content is required", http.StatusBadRequest)
		return
	}

	newNote := domain.TastingNote{
		WineID: uint(id),
		Date:   time.Now().Format("2006-01-02"),
		Note:   note,
	}

	database.DB.Create(&newNote)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
}

package add

import (
	"fmt"
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

	reviewer := r.FormValue("reviewer")
	rating := r.FormValue("rating")
	content := r.FormValue("content")

	// Simple validation
	if reviewer == "" || content == "" {
		http.Error(w, "Reviewer and content are required", http.StatusBadRequest)
		return
	}

	newReview := domain.Review{
		WineID:   uint(id),
		Reviewer: reviewer,
		Date:     "Just now", // In a real app, use time.Now().Format(...)
		Rating:   rating,
		Content:  content,
	}

	database.DB.Create(&newReview)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
}

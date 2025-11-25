package edit

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	reviewIDStr := r.FormValue("id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		http.Error(w, "Invalid Review ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uint)

	var review domain.Review
	if result := database.DB.First(&review, reviewID); result.Error != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	// Check ownership of the wine associated with the review
	var wine domain.Wine
	if result := database.DB.First(&wine, review.WineID); result.Error != nil {
		http.Error(w, "Wine not found", http.StatusNotFound)
		return
	}

	if wine.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	reviewer := r.FormValue("reviewer")
	rating := r.FormValue("rating")
	content := r.FormValue("content")
	link := r.FormValue("link")

	if reviewer == "" || content == "" {
		http.Error(w, "Reviewer and content are required", http.StatusBadRequest)
		return
	}

	if link != "" && !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = "https://" + link
	}

	review.Reviewer = reviewer
	review.Rating = rating
	review.Content = content
	review.Link = link

	database.DB.Save(&review)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", wine.ID), http.StatusSeeOther)
}

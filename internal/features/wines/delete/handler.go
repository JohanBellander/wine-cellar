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

	// Parse multipart form explicitly to ensure we get the ID
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// If multipart parsing fails, try regular ParseForm just in case
		r.ParseForm()
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Perform the delete
	if result := database.DB.Delete(&domain.Wine{}, id); result.Error != nil {
		http.Error(w, "Database error: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

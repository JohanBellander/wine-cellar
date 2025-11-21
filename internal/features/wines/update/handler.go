package update

import (
	"fmt"
	"net/http"
	"strconv"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
)

func QuantityHandler(w http.ResponseWriter, r *http.Request) {
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

	action := r.FormValue("action")

	var wine domain.Wine
	result := database.DB.First(&wine, id)
	if result.Error != nil {
		http.NotFound(w, r)
		return
	}

	if action == "increment" {
		wine.Quantity++
	} else if action == "decrement" {
		if wine.Quantity > 0 {
			wine.Quantity--
		}
	}

	database.DB.Save(&wine)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
}

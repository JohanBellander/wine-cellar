package list

import (
	"html/template"
	"net/http"
	"strconv"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Note: We are using paths relative to the project root
	tmpl, err := template.New("list.html").Funcs(ui.FuncMap).ParseFiles("internal/features/wines/list/list.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	// Pagination logic
	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 5 // Items per page

	var totalWines int64
	database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Count(&totalWines)
	totalPages := int((totalWines + int64(limit) - 1) / int64(limit))

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	offset := (page - 1) * limit

	var paginatedWines []domain.Wine
	database.DB.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&paginatedWines)

	// Generate page numbers
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	data := struct {
		Wines       []domain.Wine
		CurrentPage int
		TotalPages  int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int
		Pages       []int
		LoggedIn    bool
		UserEmail   string
	}{
		Wines:       paginatedWines,
		CurrentPage: page,
		TotalPages:  totalPages,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		Pages:       pages,
		LoggedIn:    true,
		UserEmail:   userEmail,
	}

	tmpl.Execute(w, data)
}

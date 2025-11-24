package list

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"wine-cellar/internal/domain"
	"wine-cellar/internal/shared/database"
	"wine-cellar/internal/shared/ui"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Note: We are using paths relative to the project root
	tmpl, err := template.New("list.html").Funcs(ui.FuncMap).ParseFiles("internal/features/wines/list/list.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	// Fetch user to check subscription tier
	var user domain.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	isPro := user.SubscriptionTier == "pro"

	// Search logic
	searchQuery := r.FormValue("q")
	filterCategory := r.FormValue("category")
	filterCountry := r.FormValue("country")
	filterRegion := r.FormValue("region")
	filterProducer := r.FormValue("producer")
	filterVintage := r.FormValue("vintage")
	
	// Build base query
	query := database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID)

	if isPro {
		if searchQuery != "" {
			likeQuery := "%" + searchQuery + "%"
			query = query.Where("name LIKE ? OR producer LIKE ? OR region LIKE ? OR category LIKE ?", likeQuery, likeQuery, likeQuery, likeQuery)
		}
		if filterCategory != "" {
			query = query.Where("category = ?", filterCategory)
		}
		if filterCountry != "" {
			query = query.Where("country = ?", filterCountry)
		}
		if filterRegion != "" {
			query = query.Where("region = ?", filterRegion)
		}
		if filterProducer != "" {
			query = query.Where("producer = ?", filterProducer)
		}
		if filterVintage == "NV" {
			query = query.Where("is_non_vintage = ?", true)
		} else if filterVintage != "" {
			query = query.Where("vintage = ?", filterVintage)
		}
	}

	// Sorting logic
	sortField := r.FormValue("sort")
	sortDirection := r.FormValue("direction")

	if sortField == "" {
		sortField = "created_at" // Default sort
		sortDirection = "desc"
	}

	// Validate sort field to prevent SQL injection
	allowedSortFields := map[string]bool{
		"name":       true,
		"category":   true,
		"producer":   true,
		"region":     true,
		"vintage":    true,
		"quantity":   true,
		"created_at": true,
	}

	if !allowedSortFields[sortField] {
		sortField = "created_at"
	}

	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	query = query.Order(sortField + " " + sortDirection)

	// Pagination logic
	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 5 // Items per page

	var totalWines int64
	query.Count(&totalWines)
	totalPages := int((totalWines + int64(limit) - 1) / int64(limit))

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	offset := (page - 1) * limit

	var paginatedWines []domain.Wine
	query.Limit(limit).Offset(offset).Find(&paginatedWines)

	// Generate page numbers
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	// Build base query string for pagination
	v := r.URL.Query()
	v.Del("page")
	baseQueryString := v.Encode()

	// Fetch filter options if Pro
	var categories []string
	var countries []string
	var regions []string
	var producers []string
	var vintages []int
	var hasNV int64

	if isPro {
		database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Distinct("category").Pluck("category", &categories)
		database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Distinct("country").Pluck("country", &countries)
		database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Distinct("region").Pluck("region", &regions)
		database.DB.Model(&domain.Wine{}).Where("user_id = ?", userID).Distinct("producer").Pluck("producer", &producers)
		database.DB.Model(&domain.Wine{}).Where("user_id = ? AND vintage > 0", userID).Distinct("vintage").Order("vintage desc").Pluck("vintage", &vintages)
		database.DB.Model(&domain.Wine{}).Where("user_id = ? AND is_non_vintage = ?", userID, true).Count(&hasNV)
	}

	data := struct {
		Wines            []domain.Wine
		CurrentPage      int
		TotalPages       int
		HasPrev          bool
		HasNext          bool
		PrevPage         int
		NextPage         int
		Pages            []int
		LoggedIn         bool
		UserEmail        string
		IsPro            bool
		SearchQuery      string
		FilterCategory   string
		FilterCountry    string
		FilterRegion     string
		FilterProducer   string
		FilterVintage    string
		FilterCategories []string
		FilterCountries  []string
		FilterRegions    []string
		FilterProducers  []string
		FilterVintages   []int
		HasNV            bool
		BaseQueryString  string
		Sort             string
		Direction        string
		QueryParams      url.Values
	}{
		Wines:            paginatedWines,
		CurrentPage:      page,
		TotalPages:       totalPages,
		HasPrev:          page > 1,
		HasNext:          page < totalPages,
		PrevPage:         page - 1,
		NextPage:         page + 1,
		Pages:            pages,
		LoggedIn:         true,
		UserEmail:        userEmail,
		IsPro:            isPro,
		SearchQuery:      searchQuery,
		FilterCategory:   filterCategory,
		FilterCountry:    filterCountry,
		FilterRegion:     filterRegion,
		FilterProducer:   filterProducer,
		FilterVintage:    filterVintage,
		FilterCategories: categories,
		FilterCountries:  countries,
		FilterRegions:    regions,
		FilterProducers:  producers,
		FilterVintages:   vintages,
		HasNV:            hasNV > 0,
		BaseQueryString:  baseQueryString,
		Sort:             sortField,
		Direction:        sortDirection,
		QueryParams:      r.URL.Query(),
	}

	tmpl.Execute(w, data)
}

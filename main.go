package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	InitDB()
	Seed(DB)

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/details/", detailsHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/update-quantity", updateQuantityHandler)
	http.HandleFunc("/add-review", addReviewHandler)

	// Serve static files if we had any, but we are using CDNs mostly.
	// If we had local images, we would serve them here.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/list.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pagination logic
	pageStr := r.FormValue("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 5 // Items per page

	var totalWines int64
	DB.Model(&Wine{}).Count(&totalWines)
	totalPages := int((totalWines + int64(limit) - 1) / int64(limit))

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	offset := (page - 1) * limit

	var paginatedWines []Wine
	DB.Limit(limit).Offset(offset).Find(&paginatedWines)

	// Generate page numbers
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	data := struct {
		Wines       []Wine
		CurrentPage int
		TotalPages  int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int
		Pages       []int
	}{
		Wines:       paginatedWines,
		CurrentPage: page,
		TotalPages:  totalPages,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		Pages:       pages,
	}

	tmpl.Execute(w, data)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/wine_form.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, Wine{})
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		vintage, _ := strconv.Atoi(r.FormValue("vintage"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

		newWine := Wine{
			Name:           r.FormValue("name"),
			Producer:       r.FormValue("producer"),
			Vintage:        vintage,
			Grape:          r.FormValue("grape"),
			Country:        r.FormValue("country"),
			Region:         r.FormValue("region"),
			Quantity:       quantity,
			Price:          price,
			DrinkingWindow: r.FormValue("drinking_window"),
			Type:           "Red", // Defaulting for now, or could infer
			ImageURL:       "https://via.placeholder.com/150", // Placeholder
		}
		
		DB.Create(&newWine)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}
	idStr := pathParts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var wine Wine
	result := DB.Preload("Reviews").First(&wine, id)
	if result.Error != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/details.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, wine)
}

func updateQuantityHandler(w http.ResponseWriter, r *http.Request) {
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

	var wine Wine
	result := DB.First(&wine, id)
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

	DB.Save(&wine)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
}

func addReviewHandler(w http.ResponseWriter, r *http.Request) {
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

	newReview := Review{
		WineID:   uint(id),
		Reviewer: reviewer,
		Date:     "Just now", // In a real app, use time.Now().Format(...)
		Rating:   rating,
		Content:  content,
	}

	DB.Create(&newReview)

	http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.NotFound(w, r)
			return
		}
		idStr := pathParts[2]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var wine Wine
		if result := DB.First(&wine, id); result.Error != nil {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFiles("templates/wine_form.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, wine)
		return
	}

	if r.Method == http.MethodPost {
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

		var wine Wine
		if result := DB.First(&wine, id); result.Error != nil {
			http.NotFound(w, r)
			return
		}

		vintage, _ := strconv.Atoi(r.FormValue("vintage"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

		wine.Name = r.FormValue("name")
		wine.Producer = r.FormValue("producer")
		wine.Vintage = vintage
		wine.Grape = r.FormValue("grape")
		wine.Country = r.FormValue("country")
		wine.Region = r.FormValue("region")
		wine.Quantity = quantity
		wine.Price = price
		wine.DrinkingWindow = r.FormValue("drinking_window")
		
		DB.Save(&wine)

		http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
	}
}

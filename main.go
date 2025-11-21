package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	InitDB()
	Seed(DB)

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/add", AuthMiddleware(addHandler))
	http.HandleFunc("/details/", AuthMiddleware(detailsHandler))
	http.HandleFunc("/edit/", AuthMiddleware(editHandler))
	http.HandleFunc("/update-quantity", AuthMiddleware(updateQuantityHandler))
	http.HandleFunc("/add-review", AuthMiddleware(addReviewHandler))
	http.HandleFunc("/settings", AuthMiddleware(settingsHandler))
	http.HandleFunc("/delete", AuthMiddleware(deleteHandler))
	http.HandleFunc("/health", healthHandler)

	// Serve static files if we had any, but we are using CDNs mostly.
	// If we had local images, we would serve them here.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Define a FuncMap with the safeURL function
var funcMap = template.FuncMap{
	"safeURL": func(s string) template.URL {
		return template.URL(s)
	},
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		// User is authenticated
		// We need to inject user_id into context like AuthMiddleware does
		userID := session.Values["user_id"].(uint)
		email := session.Values["email"].(string)
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "email", email)

		listHandler(w, r.WithContext(ctx))
		return
	}

	// User is not authenticated
	landingHandler(w, r)
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("landing.html").Funcs(funcMap).ParseFiles("templates/landing.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		LoggedIn  bool
		UserEmail string
	}{
		LoggedIn:  false,
		UserEmail: "",
	}

	tmpl.Execute(w, data)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("list.html").Funcs(funcMap).ParseFiles("templates/list.html", "templates/header.html")
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
	DB.Model(&Wine{}).Where("user_id = ?", userID).Count(&totalWines)
	totalPages := int((totalWines + int64(limit) - 1) / int64(limit))

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	offset := (page - 1) * limit

	var paginatedWines []Wine
	DB.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&paginatedWines)

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

func addHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	if r.Method == http.MethodGet {
		var user User
		if result := DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("wine_form.html").Funcs(funcMap).ParseFiles("templates/wine_form.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			Wine      Wine
			User      User
			LoggedIn  bool
			UserEmail string
		}{
			Wine:      Wine{},
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
		}
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		// Parse multipart form, 10MB max
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		vintage, _ := strconv.Atoi(r.FormValue("vintage"))
		quantity, _ := strconv.Atoi(r.FormValue("quantity"))
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

		imageURL := "https://via.placeholder.com/150"
		
		// Handle image upload
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			// Read file content
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				// Convert to base64
				base64Str := base64.StdEncoding.EncodeToString(fileBytes)
				// Determine mime type (simple check)
				mimeType := http.DetectContentType(fileBytes)
				imageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
			}
		}

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
			ImageURL:       imageURL,
			UserID:         userID,
		}

		DB.Create(&newWine)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

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

	var user User
	if result := DB.First(&user, userID); result.Error != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("details.html").Funcs(funcMap).ParseFiles("templates/details.html", "templates/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Wine      Wine
		User      User
		LoggedIn  bool
		UserEmail string
	}{
		Wine:      wine,
		User:      user,
		LoggedIn:  true,
		UserEmail: userEmail,
	}

	tmpl.Execute(w, data)
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
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

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

		var user User
		if result := DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("wine_form.html").Funcs(funcMap).ParseFiles("templates/wine_form.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			Wine      Wine
			User      User
			LoggedIn  bool
			UserEmail string
		}{
			Wine:      wine,
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
		}
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		// Parse multipart form, 10MB max
		err := r.ParseMultipartForm(10 << 20)
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
		
		// Handle image upload
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			// Read file content
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				// Convert to base64
				base64Str := base64.StdEncoding.EncodeToString(fileBytes)
				// Determine mime type (simple check)
				mimeType := http.DetectContentType(fileBytes)
				wine.ImageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
			}
		}

		DB.Save(&wine)

		http.Redirect(w, r, fmt.Sprintf("/details/%d", id), http.StatusSeeOther)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
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
	if result := DB.Delete(&Wine{}, id); result.Error != nil {
		http.Error(w, "Database error: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/signup.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		hashedPassword, err := HashPassword(password)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		user := User{Email: email, PasswordHash: hashedPassword}
		result := DB.Create(&user)
		if result.Error != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user User
		result := DB.Where("email = ?", email).First(&user)
		if result.Error != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !CheckPasswordHash(password, user.PasswordHash) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["user_id"] = user.ID
		session.Values["email"] = user.Email
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Values["email"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	userEmail := r.Context().Value("email").(string)

	if r.Method == http.MethodGet {
		var user User
		if result := DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("settings.html").Funcs(funcMap).ParseFiles("templates/settings.html", "templates/header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			User      User
			LoggedIn  bool
			UserEmail string
		}{
			User:      user,
			LoggedIn:  true,
			UserEmail: userEmail,
		}

		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		currency := r.FormValue("currency")
		
		var user User
		if result := DB.First(&user, userID); result.Error != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		user.Currency = currency
		DB.Save(&user)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

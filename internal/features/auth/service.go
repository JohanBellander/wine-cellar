package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte(getEnv("SESSION_SECRET", "super-secret-key")))

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a hashed password with a plain text password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Middleware checks if the user is logged in
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Add user ID to context for easy access in handlers
		userID := session.Values["user_id"].(uint)
		email := session.Values["email"].(string)
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "email", email)
		next(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) uint {
	if userID, ok := r.Context().Value("userID").(uint); ok {
		return userID
	}
	return 0
}

// GetUserIDFromSession retrieves the user ID directly from the session (useful if not using middleware)
func GetUserIDFromSession(r *http.Request) uint {
	session, _ := store.Get(r, "session-name")
	if userID, ok := session.Values["user_id"].(uint); ok {
		return userID
	}
	return 0
}

// GetSessionUser returns the user ID and email if authenticated
func GetSessionUser(r *http.Request) (uint, string, bool) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		userID := session.Values["user_id"].(uint)
		email := session.Values["email"].(string)
		return userID, email, true
	}
	return 0, "", false
}

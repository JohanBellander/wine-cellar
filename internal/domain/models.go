package domain

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email              string `gorm:"uniqueIndex;not null"`
	PasswordHash       string `gorm:"not null"`
	Currency           string `gorm:"default:'USD'"`
	SubscriptionTier   string `gorm:"default:'free'"` // "free" or "pro"
	StripeCustomerID   string
	SubscriptionStatus string // "active", "past_due", "canceled", etc.
	SubscriptionID     string
}

type Wine struct {
	gorm.Model
	UserID         uint    `gorm:"index"` // Foreign key to User
	Name           string
	Producer       string
	Vintage        int
	IsNonVintage   bool    `gorm:"default:false"`
	Grape          string  `json:"grape"`
	Country        string
	Region         string
	Quantity       int
	Price          float64
	ABV            float64
	Location       string
	Rating         string
	DrinkingWindow string
	Notes          string
	ImageURL       string
	Type           string  `json:"type"`
	Category       string  `json:"category"`
	SubCategory    string  `json:"sub_category"`
	BottleSize     string  `gorm:"default:'75cl'"`
	Reviews        []Review
	TastingNotes   []TastingNote
}

type Review struct {
	gorm.Model
	WineID   uint
	Reviewer string
	Date     string
	Rating   string
	Content  string
	Link     string
}

type TastingNote struct {
	gorm.Model
	WineID uint
	Date   string
	Note   string
}

package main

import (
	"gorm.io/gorm"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"os"
	"log"
)

var DB *gorm.DB

func InitDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		// Use PostgreSQL
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database: ", err)
		}
	} else {
		// Use SQLite
		DB, err = gorm.Open(sqlite.Open("wines.db"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to database: ", err)
		}
	}

	// Auto Migrate the schema
	DB.AutoMigrate(&User{}, &Wine{}, &Review{})
}

type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type Wine struct {
	gorm.Model
	UserID         uint `gorm:"index"` // Foreign key to User
	Name           string
	Producer       string
	Vintage        int
	Grape          string
	Country        string
	Region         string
	Quantity       int
	Price          float64
	ABV            float64
	Location       string
	Rating         string
	DrinkingWindow string
	Notes          string
	Type           string // Red, White, etc.
	ImageURL       string
	Reviews        []Review
}

type Review struct {
	gorm.Model
	WineID   uint
	Reviewer string
	Date     string
	Rating   string
	Content  string
}

func Seed(db *gorm.DB) {
	var count int64
	db.Model(&Wine{}).Count(&count)
	if count > 0 {
		return
	}

	wines := []Wine{
		{
			Name:           "Sangre de Toro",
			Producer:       "Torres",
			Vintage:        2019,
			Grape:          "Garnacha",
			Country:        "Spain",
			Region:         "Catalunya",
			Quantity:       2,
			Price:          15.00,
			ABV:            13.5,
			Location:       "Rack A",
			Rating:         "88p",
			DrinkingWindow: "2020-2025",
			Notes:          "Classic Garnacha with red fruit notes.",
			Type:           "Red",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCJZBBeDWwP7V0Ul-1_r1JRlubGA6xngAJ-QRlQNMcY-GiwZmc-ltzftzu5Oda2fRtqixLB9PUho7Gu2p-R5dzrxuLl6xlsuNd_rPnYPozfHCIbciHnDrrerbcVf_X6q2bnJAJ4K6b2Hd7jM5iXz7f9bpauuhpctucH20652we_7_8n4It4-PinO2kFeZeHtjybIsPwkLy9cfSZvce55qCbO6e9x98Yib_Wl7bK7bBjzxRpPDCRJSuPbGyeRl-uCQLPR1weRNmp6nFB",
			Reviews: []Review{
				{
					Reviewer: "Alex Johnson",
					Date:     "2 days ago",
					Rating:   "99p",
					Content:  "Simply breathtaking. The complexity is mind-boggling. Worth every penny for a special occasion. A true masterpiece.",
				},
				{
					Reviewer: "Samantha Bee",
					Date:     "1 week ago",
					Rating:   "93p",
					Content:  "An incredible wine, though perhaps still a bit young. The tannins are powerful but well-integrated. Will be even better in 5-10 years.",
				},
			},
		},
		{
			Name:           "Viña Sol",
			Producer:       "Torres",
			Vintage:        2022,
			Grape:          "Parellada",
			Country:        "Spain",
			Region:         "Penedès",
			Quantity:       4,
			Price:          12.00,
			ABV:            11.5,
			Location:       "Fridge",
			Rating:         "90p",
			DrinkingWindow: "2023-2024",
			Notes:          "Fresh and fruity white wine.",
			Type:           "White",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCjb9k2iqRivRNC_RFqF9OTfltf0Uf-JmEjBSk6f1712OMefedNYcghKDlRWvsHhMbIs6ue-LYhZPqU2NbHqFsurG2eyieBo1rcdX7CzV5rrH_ixWvtUQcudQTbZ1BpgxZv5D-9hYmZDssU4CtpQfc905C7NPR5MEaJsdrguiKQfC3C-2xyywuH4hfDFxNrdWr8tEFR-N0IfK01jMMJg0g1FXaqOarls2ML8S0Uh6vmT0_szFUWYcfftmZ1ltlyQCHK0-eEBa6dHaCV",
		},
		{
			Name:           "Celeste Crianza",
			Producer:       "Torres",
			Vintage:        2018,
			Grape:          "Tempranillo",
			Country:        "Spain",
			Region:         "Ribera del Duero",
			Quantity:       1,
			Price:          25.00,
			ABV:            14.0,
			Location:       "Rack B",
			Rating:         "95p",
			DrinkingWindow: "2020-2030",
			Notes:          "Intense blackberry color. Spicy and intense nose.",
			Type:           "Red",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCvxAKKR0zWJQ2-tMZ8nX73NYU0y4rwMCGIzeqBNU7brIrM_nXOxNaXBDkSg0yPzTw_opw3AIUFj0r-wslLcmd6d9lO8Z_IxvGIdkQtvwgiCAyH0Aq5RNnaE_clWTBIVi0JPHwYp795owlRWiSnuJyqf_VS3l3HRZ9PyOTQWVhpfIgLJ_SJoLL-f1Rcaykm7hY9RDqJ7P3XjwnQJ_A0X7YPItiEY4GAU9RXUpj18qGvNtxVycgXDTyLgm_QBl2qkq_guPe_Cd69y6d3",
		},
		{
			Name:           "Gran Coronas",
			Producer:       "Torres",
			Vintage:        2016,
			Grape:          "Cabernet Sauvignon",
			Country:        "Spain",
			Region:         "Penedès",
			Quantity:       3,
			Price:          20.00,
			ABV:            14.0,
			Location:       "Rack A",
			Rating:         "92p-94p",
			DrinkingWindow: "2018-2028",
			Notes:          "Notes of blueberry and cherry jam.",
			Type:           "Red",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuBwrG6hfCIh9_S3X3PF8sJPgSmXMew0hJEO75SkKnCrCCIlSF0qznSBrcnX2lS_-HEoZ248s1nt_trvCjYHARn8TY4TDSKjJ9IFF1k5q8PYwr_vaGi40ltvXMcWSeybAnfvA0D3LDtR9VsNsdGOVoMp7f3kvbtbmMQ9YUeiMJmJSolQDRWMT3hUxWSd6hsMp-UnBGwm_TdNZGmh6lHP0m_X1RvSJGiFNuTwzWWKriYW5MECsyTJ1tVHvQLkeIv39uF3pI7QM60nag6F",
		},
		{
			Name:           "Château Margaux",
			Producer:       "Château Margaux",
			Vintage:        2015,
			Grape:          "Cabernet Sauvignon",
			Country:        "France",
			Region:         "Bordeaux",
			Quantity:       1,
			Price:          700.00,
			ABV:            13.5,
			Location:       "Cellar Special",
			Rating:         "99p",
			DrinkingWindow: "2025-2050",
			Notes:          "Legendary vintage. Violet, black currant, and truffle.",
			Type:           "Red",
			ImageURL:       "https://via.placeholder.com/150?text=Margaux",
		},
		{
			Name:           "Tignanello",
			Producer:       "Antinori",
			Vintage:        2018,
			Grape:          "Sangiovese",
			Country:        "Italy",
			Region:         "Tuscany",
			Quantity:       2,
			Price:          140.00,
			ABV:            14.0,
			Location:       "Rack B",
			Rating:         "96p",
			DrinkingWindow: "2023-2040",
			Notes:          "Super Tuscan. Cherry, spice, and earth.",
			Type:           "Red",
			ImageURL:       "https://via.placeholder.com/150?text=Tignanello",
		},
		{
			Name:           "Opus One",
			Producer:       "Opus One Winery",
			Vintage:        2017,
			Grape:          "Cabernet Sauvignon",
			Country:        "USA",
			Region:         "Napa Valley",
			Quantity:       2,
			Price:          350.00,
			ABV:            14.5,
			Location:       "Rack C",
			Rating:         "97p",
			DrinkingWindow: "2022-2045",
			Notes:          "Iconic Napa blend. Black fruit, cocoa, and velvet tannins.",
			Type:           "Red",
			ImageURL:       "https://via.placeholder.com/150?text=Opus+One",
		},
		{
			Name:           "Cloudy Bay Sauvignon Blanc",
			Producer:       "Cloudy Bay",
			Vintage:        2023,
			Grape:          "Sauvignon Blanc",
			Country:        "New Zealand",
			Region:         "Marlborough",
			Quantity:       6,
			Price:          30.00,
			ABV:            13.0,
			Location:       "Fridge",
			Rating:         "91p",
			DrinkingWindow: "2023-2026",
			Notes:          "Zesty lime, passionfruit, and fresh herbs.",
			Type:           "White",
			ImageURL:       "https://via.placeholder.com/150?text=Cloudy+Bay",
		},
		{
			Name:           "Barolo Lazzarito",
			Producer:       "Vietti",
			Vintage:        2016,
			Grape:          "Nebbiolo",
			Country:        "Italy",
			Region:         "Piedmont",
			Quantity:       1,
			Price:          180.00,
			ABV:            14.5,
			Location:       "Rack B",
			Rating:         "95p",
			DrinkingWindow: "2024-2045",
			Notes:          "Powerful tannins, tar, and roses.",
			Type:           "Red",
			ImageURL:       "https://via.placeholder.com/150?text=Barolo",
		},
		{
			Name:           "Whispering Angel",
			Producer:       "Château d'Esclans",
			Vintage:        2022,
			Grape:          "Grenache",
			Country:        "France",
			Region:         "Provence",
			Quantity:       3,
			Price:          25.00,
			ABV:            13.0,
			Location:       "Fridge",
			Rating:         "89p",
			DrinkingWindow: "2023-2024",
			Notes:          "Pale pink, refreshing strawberry and citrus.",
			Type:           "Rosé",
			ImageURL:       "https://via.placeholder.com/150?text=Whispering+Angel",
		},
		{
			Name:           "Penfolds Grange",
			Producer:       "Penfolds",
			Vintage:        2014,
			Grape:          "Shiraz",
			Country:        "Australia",
			Region:         "South Australia",
			Quantity:       1,
			Price:          600.00,
			ABV:            14.5,
			Location:       "Cellar Special",
			Rating:         "98p",
			DrinkingWindow: "2025-2055",
			Notes:          "Rich, intense, dark chocolate and blackberry.",
			Type:           "Red",
			ImageURL:       "https://via.placeholder.com/150?text=Grange",
		},
		{
			Name:           "Dom Pérignon",
			Producer:       "Moët & Chandon",
			Vintage:        2012,
			Grape:          "Chardonnay",
			Country:        "France",
			Region:         "Champagne",
			Quantity:       2,
			Price:          200.00,
			ABV:            12.5,
			Location:       "Fridge",
			Rating:         "96p",
			DrinkingWindow: "2022-2040",
			Notes:          "Toasty, brioche, and citrus zest.",
			Type:           "Sparkling",
			ImageURL:       "https://via.placeholder.com/150?text=Dom+Perignon",
		},
	}

	db.Create(&wines)
}

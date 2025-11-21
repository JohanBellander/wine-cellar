package database

import (
	"log"
	"os"
	"wine-cellar/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	DB.AutoMigrate(&domain.User{}, &domain.Wine{}, &domain.Review{})
}

func Seed(db *gorm.DB) {
	var count int64
	db.Model(&domain.Wine{}).Count(&count)
	if count > 0 {
		return
	}

	wines := []domain.Wine{
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
			Reviews: []domain.Review{
				{
					Reviewer: "Alex Johnson",
					Date:     "2 days ago",
					Rating:   "99p",
					Content:  "Simply breathtaking. The complexity is mind-boggling. Worth every penny for a special occasion. A true masterpiece.",
				},
			},
		},
		{
			Name:           "Chablis Grand Cru",
			Producer:       "Domaine Laroche",
			Vintage:        2018,
			Grape:          "Chardonnay",
			Country:        "France",
			Region:         "Burgundy",
			Quantity:       1,
			Price:          85.00,
			ABV:            13.0,
			Location:       "Rack B",
			Rating:         "94p",
			DrinkingWindow: "2022-2030",
			Notes:          "Crisp acidity with mineral undertones.",
			Type:           "White",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCJZBBeDWwP7V0Ul-1_r1JRlubGA6xngAJ-QRlQNMcY-GiwZmc-ltzftzu5Oda2fRtqixLB9PUho7Gu2p-R5dzrxuLl6xlsuNd_rPnYPozfHCIbciHnDrrerbcVf_X6q2bnJAJ4K6b2Hd7jM5iXz7f9bpauuhpctucH20652we_7_8n4It4-PinO2kFeZeHtjybIsPwkLy9cfSZvce55qCbO6e9x98Yib_Wl7bK7bBjzxRpPDCRJSuPbGyeRl-uCQLPR1weRNmp6nFB",
			Reviews: []domain.Review{
				{
					Reviewer: "Maria Garcia",
					Date:     "1 week ago",
					Rating:   "95p",
					Content:  "An absolute delight! The balance of flavors is exquisite. Highly recommended for anyone who appreciates fine wine.",
				},
			},
		},
		{
			Name:           "Barolo",
			Producer:       "Pio Cesare",
			Vintage:        2016,
			Grape:          "Nebbiolo",
			Country:        "Italy",
			Region:         "Piedmont",
			Quantity:       3,
			Price:          60.00,
			ABV:            14.5,
			Location:       "Rack C",
			Rating:         "92p",
			DrinkingWindow: "2024-2035",
			Notes:          "Robust tannins with cherry and tar aromas.",
			Type:           "Red",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCJZBBeDWwP7V0Ul-1_r1JRlubGA6xngAJ-QRlQNMcY-GiwZmc-ltzftzu5Oda2fRtqixLB9PUho7Gu2p-R5dzrxuLl6xlsuNd_rPnYPozfHCIbciHnDrrerbcVf_X6q2bnJAJ4K6b2Hd7jM5iXz7f9bpauuhpctucH20652we_7_8n4It4-PinO2kFeZeHtjybIsPwkLy9cfSZvce55qCbO6e9x98Yib_Wl7bK7bBjzxRpPDCRJSuPbGyeRl-uCQLPR1weRNmp6nFB",
			Reviews: []domain.Review{
				{
					Reviewer: "John Smith",
					Date:     "3 days ago",
					Rating:   "88p",
					Content:  "A solid choice, but I expected a bit more depth. Still, a very enjoyable experience overall.",
				},
			},
		},
		{
			Name:           "Riesling Kabinett",
			Producer:       "Dr. Loosen",
			Vintage:        2020,
			Grape:          "Riesling",
			Country:        "Germany",
			Region:         "Mosel",
			Quantity:       6,
			Price:          22.00,
			ABV:            8.5,
			Location:       "Fridge",
			Rating:         "90p",
			DrinkingWindow: "2021-2028",
			Notes:          "Off-dry with high acidity and slate notes.",
			Type:           "White",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCJZBBeDWwP7V0Ul-1_r1JRlubGA6xngAJ-QRlQNMcY-GiwZmc-ltzftzu5Oda2fRtqixLB9PUho7Gu2p-R5dzrxuLl6xlsuNd_rPnYPozfHCIbciHnDrrerbcVf_X6q2bnJAJ4K6b2Hd7jM5iXz7f9bpauuhpctucH20652we_7_8n4It4-PinO2kFeZeHtjybIsPwkLy9cfSZvce55qCbO6e9x98Yib_Wl7bK7bBjzxRpPDCRJSuPbGyeRl-uCQLPR1weRNmp6nFB",
			Reviews: []domain.Review{
				{
					Reviewer: "Emily Davis",
					Date:     "5 days ago",
					Rating:   "92p",
					Content:  "Refreshing and crisp! Perfect for a summer evening. Will definitely buy again.",
				},
			},
		},
		{
			Name:           "Malbec Reserva",
			Producer:       "Catena Zapata",
			Vintage:        2018,
			Grape:          "Malbec",
			Country:        "Argentina",
			Region:         "Mendoza",
			Quantity:       4,
			Price:          25.00,
			ABV:            14.0,
			Location:       "Rack A",
			Rating:         "91p",
			DrinkingWindow: "2020-2026",
			Notes:          "Rich plum flavors with a hint of vanilla.",
			Type:           "Red",
			ImageURL:       "https://lh3.googleusercontent.com/aida-public/AB6AXuCJZBBeDWwP7V0Ul-1_r1JRlubGA6xngAJ-QRlQNMcY-GiwZmc-ltzftzu5Oda2fRtqixLB9PUho7Gu2p-R5dzrxuLl6xlsuNd_rPnYPozfHCIbciHnDrrerbcVf_X6q2bnJAJ4K6b2Hd7jM5iXz7f9bpauuhpctucH20652we_7_8n4It4-PinO2kFeZeHtjybIsPwkLy9cfSZvce55qCbO6e9x98Yib_Wl7bK7bBjzxRpPDCRJSuPbGyeRl-uCQLPR1weRNmp6nFB",
			Reviews: []domain.Review{
				{
					Reviewer: "Michael Brown",
					Date:     "1 day ago",
					Rating:   "90p",
					Content:  "Great value for money. Smooth finish and lovely aroma. A crowd pleaser for sure.",
				},
			},
		},
	}

	for _, wine := range wines {
		db.Create(&wine)
	}
}

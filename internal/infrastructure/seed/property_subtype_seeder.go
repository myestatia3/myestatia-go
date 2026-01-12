package seed

import (
	"log"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

func SeedPropertySubtypes(db *gorm.DB) error {
	var count int64
	if err := db.Model(&entity.PropertySubtype{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	subtypes := []entity.PropertySubtype{
		// APARTMENT
		{Name: "ground_floor", DisplayName: "Ground Floor", Type: entity.TypeApartment},
		{Name: "mid_floor", DisplayName: "Mid Floor", Type: entity.TypeApartment},
		{Name: "top_floor", DisplayName: "Top Floor", Type: entity.TypeApartment},
		{Name: "penthouse", DisplayName: "Penthouse", Type: entity.TypeApartment},
		{Name: "duplex_penthouse", DisplayName: "Duplex Penthouse", Type: entity.TypeApartment},
		{Name: "duplex", DisplayName: "Duplex", Type: entity.TypeApartment},
		{Name: "studio_gf", DisplayName: "Studio (Ground Floor)", Type: entity.TypeApartment},
		{Name: "studio_mf", DisplayName: "Studio (Mid Floor)", Type: entity.TypeApartment},
		{Name: "studio_tf", DisplayName: "Studio (Top Floor)", Type: entity.TypeApartment},

		// HOUSE
		{Name: "villa", DisplayName: "Villa", Type: entity.TypeHouse},
		{Name: "semi_detached", DisplayName: "Semi-Detached", Type: entity.TypeHouse},
		{Name: "terraced", DisplayName: "Terraced", Type: entity.TypeHouse},
		{Name: "finca", DisplayName: "Finca / Country House", Type: entity.TypeHouse},
		{Name: "bungalow", DisplayName: "Bungalow", Type: entity.TypeHouse},
		{Name: "quad", DisplayName: "Quad", Type: entity.TypeHouse},
		{Name: "castle", DisplayName: "Castle", Type: entity.TypeHouse},
		{Name: "city_mansion", DisplayName: "City Mansion", Type: entity.TypeHouse},
		{Name: "wooden_house", DisplayName: "Wooden House", Type: entity.TypeHouse},
		{Name: "wooden_cabin", DisplayName: "Wooden Cabin", Type: entity.TypeHouse},
		{Name: "motorhome", DisplayName: "Motorhome / Mobile Home", Type: entity.TypeHouse},
		{Name: "cave_house", DisplayName: "Cave House", Type: entity.TypeHouse},

		// LAND
		{Name: "urban", DisplayName: "Urban Land", Type: entity.TypeLand},
		{Name: "commercial", DisplayName: "Commercial Land", Type: entity.TypeLand},
		{Name: "rustic", DisplayName: "Rustic Land", Type: entity.TypeLand},
		{Name: "with_ruin", DisplayName: "Land with Ruin", Type: entity.TypeLand},

		// COMMERCIAL
		{Name: "bar", DisplayName: "Bar", Type: entity.TypeCommercial},
		{Name: "restaurant", DisplayName: "Restaurant", Type: entity.TypeCommercial},
		{Name: "hotel", DisplayName: "Hotel", Type: entity.TypeCommercial},
		{Name: "hostel", DisplayName: "Hostel", Type: entity.TypeCommercial},
		{Name: "shop", DisplayName: "Shop / Retail", Type: entity.TypeCommercial},
		{Name: "office", DisplayName: "Office", Type: entity.TypeCommercial},
		{Name: "warehouse", DisplayName: "Warehouse", Type: entity.TypeCommercial},
		{Name: "garage", DisplayName: "Garage", Type: entity.TypeCommercial},
		{Name: "camping", DisplayName: "Camping", Type: entity.TypeCommercial},
		{Name: "aparthotel", DisplayName: "Aparthotel", Type: entity.TypeCommercial},
		{Name: "other", DisplayName: "Other", Type: entity.TypeCommercial},
	}

	log.Println("Seeding property subtypes...")
	return db.Create(&subtypes).Error
}

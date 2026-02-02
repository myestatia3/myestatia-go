package resales

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/datatypes"
)

func MapToDomain(resalesProp Property, companyID string) (*entity.Property, error) {
	price := toFloat(resalesProp.Price)
	rooms := toInt(resalesProp.Bedrooms)
	baths := toInt(resalesProp.Bathrooms)
	area := toFloat(resalesProp.Built)
	propType := mapType(resalesProp.Type)

	photos := []string{}
	if resalesProp.MainImage != "" {
		photos = append(photos, resalesProp.MainImage)
	}
	for _, pic := range resalesProp.Pictures.Picture {
		if pic.URL != "" && pic.URL != resalesProp.MainImage { // Avoid duplicates if possible
			photos = append(photos, pic.URL)
		}
	}
	photosJSON, _ := json.Marshal(photos)

	featuresMap := make(map[string][]string)
	for _, cat := range resalesProp.PropertyFeatures.Category {
		var values []string
		switch v := cat.Value.(type) {
		case string:
			values = []string{v}
		case []interface{}: // JSON decoding often gives []interface{}
			for _, item := range v {
				if str, ok := item.(string); ok {
					values = append(values, str)
				}
			}
		case []string:
			values = v
		}
		featuresMap[cat.Type] = values
	}
	featuresJSON, _ := json.Marshal(featuresMap)

	now := time.Now()

	prop := &entity.Property{
		Reference:   resalesProp.Reference,
		CompanyID:   companyID,
		Origin:      entity.OriginImport,
		Status:      "AVAILABLE",
		Title:       fmt.Sprintf("%s in %s", resalesProp.Type, resalesProp.Location),
		Description: resalesProp.Description,
		Type:        propType,
		Country:     "Spain",
		City:        resalesProp.Location,
		Address:     resalesProp.Location,
		AreaM2:      area,
		Rooms:       rooms,
		Bathrooms:   baths,
		Price:       price,
		Currency:    resalesProp.Currency,
		Image:       resalesProp.MainImage,
		Photos:      datatypes.JSON(photosJSON),
		Features:    datatypes.JSON(featuresJSON),
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: &now,
	}

	return prop, nil
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	}
	return 0
}

func mapType(resalesType string) entity.PropertyType {
	lowerType := strings.ToLower(resalesType)
	if strings.Contains(lowerType, "apartment") || strings.Contains(lowerType, "flat") || strings.Contains(lowerType, "penthouse") {
		return entity.TypeApartment
	}
	if strings.Contains(lowerType, "villa") || strings.Contains(lowerType, "house") || strings.Contains(lowerType, "chalet") || strings.Contains(lowerType, "townhouse") {
		return entity.TypeHouse
	}
	if strings.Contains(lowerType, "land") || strings.Contains(lowerType, "plot") {
		return entity.TypeLand
	}
	if strings.Contains(lowerType, "commercial") || strings.Contains(lowerType, "office") || strings.Contains(lowerType, "shop") {
		return entity.TypeCommercial
	}
	return entity.TypeOther
}

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

func MapToDomain(resalesProp Property, companyID string, processedCount int) (*entity.Property, error) {
	// CRITICAL: Use AgencyRef as our Reference (unique identifier)
	// If AgencyRef is empty, fall back to Resales Reference
	reference := resalesProp.AgencyRef
	if reference == "" {
		reference = resalesProp.Reference
	}
	if reference == "" {
		return nil, fmt.Errorf("property has no reference")
	}

	// Parse numeric fields
	price := toFloat(resalesProp.Price)
	originalPrice := toFloat(resalesProp.OriginalPrice)
	rooms := toInt(resalesProp.Bedrooms)
	baths := toInt(resalesProp.Bathrooms)
	area := toFloat(resalesProp.Built)
	plotM2 := toFloat(resalesProp.GardenPlot)
	terraceM2 := toFloat(resalesProp.Terrace)

	// Map property type using nested PropertyType structure
	propType := mapPropertyType(resalesProp.PropertyType)

	// Map status
	status := mapStatus(resalesProp.Status)

	// Use CDN URLs directly - no need to download images
	// This saves disk space and the CDN images display correctly
	var mainImage string
	var photos []string

	// Add main image first
	if resalesProp.MainImage != "" {
		mainImage = resalesProp.MainImage
		photos = append(photos, resalesProp.MainImage)
	}

	// Add all other photos from Pictures
	for _, pic := range resalesProp.Pictures.Picture {
		if pic.URL != "" && pic.URL != resalesProp.MainImage {
			photos = append(photos, pic.URL)
		}
	}

	// If no main image but we have photos, use the first one as main
	if mainImage == "" && len(photos) > 0 {
		mainImage = photos[0]
	}

	photosJSON, _ := json.Marshal(photos)

	// Extract and flatten features
	featuresMap := make(map[string]interface{})
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
		if len(values) > 0 {
			featuresMap[cat.Type] = values
		}
	}

	// Add boolean flags for common features
	featuresMap["has_pool"] = resalesProp.Pool == 1 || hasFeatureValue(featuresMap, "Pool")
	featuresMap["has_parking"] = resalesProp.Parking == 1 || hasFeatureValue(featuresMap, "Parking")
	featuresMap["has_garden"] = resalesProp.Garden == 1 || hasFeatureValue(featuresMap, "Garden")

	featuresJSON, _ := json.Marshal(featuresMap)

	// Build metadata with Resales-specific fields
	metadata := map[string]interface{}{
		"resales_reference":    resalesProp.Reference, // Store original Resales reference
		"resales_type_id":      resalesProp.PropertyType.TypeId,
		"resales_type":         resalesProp.PropertyType.Type,
		"resales_name_type":    resalesProp.PropertyType.NameType,
		"resales_subtype1":     resalesProp.PropertyType.Subtype1,
		"resales_subtype1_id":  resalesProp.PropertyType.SubtypeId1,
		"resales_subtype2":     resalesProp.PropertyType.Subtype2,
		"resales_subtype2_id":  resalesProp.PropertyType.SubtypeId2,
		"resales_own_property": resalesProp.OwnProperty,
		"resales_co2_rated":    resalesProp.CO2Rated,
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Build full address from location hierarchy
	address := buildAddress(resalesProp)

	// Generate title
	title := generateTitle(resalesProp)

	now := time.Now()

	prop := &entity.Property{
		Reference:         reference, // AgencyRef (or Reference as fallback)
		CompanyID:         companyID,
		Origin:            "RESALES", // Portal-specific origin for filtering
		Status:            status,
		Title:             title,
		Description:       resalesProp.Description,
		Type:              propType,
		Country:           resalesProp.Country,
		Province:          resalesProp.Province,
		City:              resalesProp.Location,
		Zone:              resalesProp.Area, // Costa del Sol, etc.
		Address:           address,
		AreaM2:            area,
		Rooms:             rooms,
		Bathrooms:         baths,
		Price:             price,
		OriginalPrice:     originalPrice,
		PlotM2:            plotM2,
		TerraceM2:         terraceM2,
		Currency:          resalesProp.Currency,
		EnergyCertificate: resalesProp.EnergyRated,
		Image:             mainImage,
		Photos:            datatypes.JSON(photosJSON),
		Features:          datatypes.JSON(featuresJSON),
		Metadata:          datatypes.JSON(metadataJSON),
		CreatedAt:         now,
		UpdatedAt:         now,
		PublishedAt:       &now,
	}

	return prop, nil
}

// mapPropertyType maps the nested PropertyType structure to our entity.PropertyType enum
func mapPropertyType(apiType PropertyTypeInfo) entity.PropertyType {
	mainType := strings.ToLower(apiType.Type)
	nameType := strings.ToLower(apiType.NameType)

	// Check for commercial properties first
	if strings.Contains(mainType, "commercial") {
		// Check specific subtypes
		if strings.Contains(nameType, "garage") ||
			strings.Contains(nameType, "parking") ||
			strings.Contains(nameType, "storage") ||
			strings.Contains(nameType, "shop") ||
			strings.Contains(nameType, "office") {
			return entity.TypeCommercial
		}
		return entity.TypeCommercial
	}

	// Apartments and flats
	if strings.Contains(nameType, "apartment") ||
		strings.Contains(nameType, "flat") ||
		strings.Contains(nameType, "penthouse") ||
		strings.Contains(nameType, "studio") {
		return entity.TypeApartment
	}

	// Houses and villas
	if strings.Contains(nameType, "villa") ||
		strings.Contains(nameType, "house") ||
		strings.Contains(nameType, "chalet") ||
		strings.Contains(nameType, "townhouse") ||
		strings.Contains(nameType, "bungalow") {
		return entity.TypeHouse
	}

	// Land and plots
	if strings.Contains(nameType, "land") ||
		strings.Contains(nameType, "plot") {
		return entity.TypeLand
	}

	return entity.TypeOther
}

// mapStatus maps the nested Status structure to our status string
func mapStatus(apiStatus PropertyStatus) string {
	switch strings.ToUpper(apiStatus.System) {
	case "AVAILABLE":
		return "available"
	case "UNDER OFFER":
		return "reserved"
	case "SALE AGREED":
		return "sold"
	default:
		return "available"
	}
}

// buildAddress creates a full address from the location fields
func buildAddress(prop Property) string {
	parts := []string{}

	if prop.SubLocation != "" {
		parts = append(parts, prop.SubLocation)
	}
	if prop.Location != "" {
		parts = append(parts, prop.Location)
	}
	if prop.Area != "" {
		parts = append(parts, prop.Area)
	}
	if prop.Province != "" && prop.Province != prop.Area {
		parts = append(parts, prop.Province)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return prop.Location
}

// generateTitle creates a descriptive title for the property
func generateTitle(prop Property) string {
	typeName := prop.PropertyType.NameType
	if typeName == "" {
		typeName = prop.PropertyType.Type
	}

	location := prop.Location
	if location == "" {
		location = prop.Area
	}

	if typeName != "" && location != "" {
		return fmt.Sprintf("%s in %s", typeName, location)
	}
	if typeName != "" {
		return typeName
	}
	return "Property"
}

// hasFeatureValue checks if a feature category has any values
func hasFeatureValue(features map[string]interface{}, key string) bool {
	if val, ok := features[key]; ok {
		switch v := val.(type) {
		case []string:
			return len(v) > 0
		case []interface{}:
			return len(v) > 0
		}
	}
	return false
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

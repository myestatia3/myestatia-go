package resales

type APIResponse struct {
	Transaction Transaction `json:"transaction"`
	QueryInfo   *QueryInfo  `json:"QueryInfo,omitempty"`
	Property    []Property  `json:"Property,omitempty"`
}

type Transaction struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type QueryInfo struct {
	QueryId       string `json:"QueryId"`
	PropertyCount int    `json:"PropertyCount"`
	CurrentPage   int    `json:"CurrentPage"`
}

type Property struct {
	Reference        string           `json:"Reference"`
	AgencyRef        string           `json:"AgencyRef"` // Agency's internal reference - THIS is our unique Reference
	Country          string           `json:"Country"`
	Province         string           `json:"Province"`
	Area             string           `json:"Area"`
	Location         string           `json:"Location"`
	SubLocation      string           `json:"SubLocation"`
	PropertyType     PropertyTypeInfo `json:"PropertyType"`
	Status           PropertyStatus   `json:"Status"`
	Price            interface{}      `json:"Price"` // Can be string or number
	OriginalPrice    interface{}      `json:"OriginalPrice"`
	Currency         string           `json:"Currency"`
	Bedrooms         interface{}      `json:"Bedrooms"`
	Bathrooms        interface{}      `json:"Bathrooms"`
	Built            interface{}      `json:"Built"`
	Plot             interface{}      `json:"Plot"`
	GardenPlot       interface{}      `json:"GardenPlot"`
	Terrace          interface{}      `json:"Terrace"`
	Pool             int              `json:"Pool"`
	Parking          int              `json:"Parking"`
	Garden           int              `json:"Garden"`
	EnergyRated      string           `json:"EnergyRated"`
	CO2Rated         string           `json:"CO2Rated"`
	Dimensions       string           `json:"Dimensions"`
	OwnProperty      string           `json:"OwnProperty"`
	PropertyFeatures PropertyFeatures `json:"PropertyFeatures"`
	Description      string           `json:"Description"`
	MainImage        string           `json:"MainImage"`
	Pictures         Pictures         `json:"Pictures"`
}

type PropertyTypeInfo struct {
	NameType   string `json:"NameType"`
	Type       string `json:"Type"`
	TypeId     string `json:"TypeId"`
	Subtype1   string `json:"Subtype1,omitempty"`
	SubtypeId1 string `json:"SubtypeId1,omitempty"`
	Subtype2   string `json:"Subtype2,omitempty"`
	SubtypeId2 string `json:"SubtypeId2,omitempty"`
}

type PropertyStatus struct {
	System string `json:"system"`
	En     string `json:"en"`
}

type PropertyFeatures struct {
	Category []FeatureCategory `json:"Category"`
}

type FeatureCategory struct {
	Type  string      `json:"Type"`
	Value interface{} `json:"Value"`
}

type Pictures struct {
	Picture []Picture `json:"Picture"`
}

type Picture struct {
	URL string `json:"PictureURL"`
}

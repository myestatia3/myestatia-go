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
	Price            interface{}      `json:"Price"` // Can be string or number in some APIs, best to handle carefully
	Currency         string           `json:"Currency"`
	PropertyFeatures PropertyFeatures `json:"PropertyFeatures"`
	Description      string           `json:"Description"`
	MainImage        string           `json:"MainImage"`
	Pictures         Pictures         `json:"Pictures"`
	Location         string           `json:"Location"`
	Type             string           `json:"Type"`
	Bedrooms         interface{}      `json:"Bedrooms"`
	Bathrooms        interface{}      `json:"Bathrooms"`
	Plot             interface{}      `json:"Plot"`
	Built            interface{}      `json:"Built"`
	Terrace          interface{}      `json:"Terrace"`
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

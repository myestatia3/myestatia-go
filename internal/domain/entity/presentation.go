package entity

type PresentationToken struct {
	LeadID      string   `json:"leadId"`
	PropertyIDs []string `json:"propertyIds"`
	ExpiresAt   int64    `json:"exp"`
}

type Presentation struct {
	Lead         *Lead      `json:"lead"`
	Properties   []Property `json:"properties"`
	ContactPhone string     `json:"contactPhone"`
}

type PropertyMatch struct {
	Property     *Property `json:"property"`
	MatchPercent int       `json:"matchPercent"`
	IsInquired   bool      `json:"isInquired"`
	IsDismissed  bool      `json:"isDismissed"`
}

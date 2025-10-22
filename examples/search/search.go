package search

// SearchComponent represents the data for a search component.
type SearchComponent struct {
	Query       string `form:"q"`
	Limit       int    `form:"limit"`
	IsBoosted   bool   `json:"-"` // Set by SetHxBoosted
	IsRequest   bool   `json:"-"` // Set by SetHxRequest
	CurrentURL  string `json:"-"` // Set by SetHxCurrentURL
	TriggerName string `json:"-"` // Set by SetHxTriggerName
}

// Implement request header interfaces

func (c *SearchComponent) SetHxBoosted(v bool) {
	c.IsBoosted = v
}

func (c *SearchComponent) SetHxRequest(v bool) {
	c.IsRequest = v
}

func (c *SearchComponent) SetHxCurrentURL(v string) {
	c.CurrentURL = v
}

func (c *SearchComponent) SetHxTriggerName(v string) {
	c.TriggerName = v
}

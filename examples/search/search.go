package search

// SearchForm represents the data for a search component.
type SearchForm struct {
	Query       string `form:"q"`
	Limit       int    `form:"limit"`
	IsBoosted   bool   `json:"-"` // Set by SetHxBoosted
	IsRequest   bool   `json:"-"` // Set by SetHxRequest
	CurrentURL  string `json:"-"` // Set by SetHxCurrentURL
	TriggerName string `json:"-"` // Set by SetHxTriggerName
}

// Implement request header interfaces

func (s *SearchForm) SetHxBoosted(v bool) {
	s.IsBoosted = v
}

func (s *SearchForm) SetHxRequest(v bool) {
	s.IsRequest = v
}

func (s *SearchForm) SetHxCurrentURL(v string) {
	s.CurrentURL = v
}

func (s *SearchForm) SetHxTriggerName(v string) {
	s.TriggerName = v
}

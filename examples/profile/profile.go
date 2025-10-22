package profile

// ProfileComponent represents the data for a user profile component.
type ProfileComponent struct {
	Name        string   `form:"name"`
	Email       string   `form:"email"`
	Tags        []string `form:"tags"`
	LocationURL string   `json:"-"` // Response header
	Success     bool     `json:"-"`
}

// Implement response header interface

func (c *ProfileComponent) GetHxLocation() string {
	return c.LocationURL
}

// Process implements the Processor interface to handle profile update logic.
// This is called automatically by the registry after form decoding
// and before rendering the component.
func (c *ProfileComponent) Process() error {
	// Simple validation
	if c.Name == "" || c.Email == "" {
		return nil
	}

	c.Success = true
	// In a real app, you might redirect to the profile view page
	// c.LocationURL = "/profile/view"
	return nil
}

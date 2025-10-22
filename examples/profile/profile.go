package profile

// UserProfile represents the data for a user profile component.
type UserProfile struct {
	Name        string   `form:"name"`
	Email       string   `form:"email"`
	Tags        []string `form:"tags"`
	LocationURL string   `json:"-"` // Response header
	Success     bool     `json:"-"`
}

// Implement response header interface

func (u *UserProfile) GetHxLocation() string {
	return u.LocationURL
}

// ProcessUpdate simulates profile update logic.
func (u *UserProfile) ProcessUpdate() error {
	// Simple validation
	if u.Name == "" || u.Email == "" {
		return nil
	}

	u.Success = true
	// In a real app, you might redirect to the profile view page
	// u.LocationURL = "/profile/view"
	return nil
}

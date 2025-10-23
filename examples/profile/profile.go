package profile

import (
	"context"
	"io"
)

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
func (c *ProfileComponent) Process(ctx context.Context) error {
	// Simple validation
	if c.Name == "" || c.Email == "" {
		return nil
	}

	c.Success = true
	// In a real app, you might redirect to the profile view page
	// c.LocationURL = "/profile/view"
	return nil
}

// Render implements templ.Component interface.
// This allows the component to be used both as an HTMX component
// and as a regular templ component in templates.
func (c *ProfileComponent) Render(ctx context.Context, w io.Writer) error {
	return Profile(*c).Render(ctx, w)
}

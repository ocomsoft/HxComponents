package card

import (
	"context"
	"fmt"
	"io"
	"time"
)

// CardComponent demonstrates the Constructor pattern with Init interface.
// This component can be used both in templ templates via the constructor
// and as an HTMX component via form submissions.
type CardComponent struct {
	Title       string `form:"title"`
	Count       int    `form:"count"`
	Description string `json:"-"` // Computed field
	Timestamp   string `json:"-"` // Computed field
	InitCalled  bool   `json:"-"` // For demonstration
}

// NewCard is a constructor function for use in templ templates.
// It creates a new CardComponent and calls Init to set up defaults.
//
// Usage in templ:
//
//	@card.NewCard(ctx, "My Card", 42)
func NewCard(ctx context.Context, title string, count int) *CardComponent {
	c := &CardComponent{
		Title: title,
		Count: count,
	}
	// Call Init to set up defaults and computed fields
	// Errors are ignored in template context - component will still render
	_ = c.Init(ctx)
	return c
}

// Init implements the Initializer interface.
// This is called:
// - Explicitly by the constructor for template usage
// - Automatically by the registry for HTMX form submissions
func (c *CardComponent) Init(ctx context.Context) error {
	// Set defaults
	if c.Title == "" {
		c.Title = "Untitled Card"
	}

	// Compute derived fields
	c.Description = fmt.Sprintf("This card contains %d item(s)", c.Count)
	c.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	c.InitCalled = true

	// You could also load data from database here using ctx
	// user, err := db.GetUser(ctx, userID)
	// if err != nil {
	//     return err
	// }

	return nil
}

// Render implements templ.Component interface.
// This allows the component to be used both as an HTMX component
// and as a regular templ component in templates.
func (c *CardComponent) Render(ctx context.Context, w io.Writer) error {
	return Card(*c).Render(ctx, w)
}

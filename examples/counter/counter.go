package counter

import (
	"context"
	"io"
)

// CounterComponent represents a simple counter that can increment/decrement.
type CounterComponent struct {
	Count  int    `form:"count"`
	Action string `form:"action"` // "increment" or "decrement"
}

// Process implements the Processor interface to handle counter logic.
func (c *CounterComponent) Process() error {
	switch c.Action {
	case "increment":
		c.Count++
	case "decrement":
		c.Count--
	}
	return nil
}

// Render implements templ.Component interface.
// This allows the component to be used both as an HTMX component
// and as a regular templ component in templates.
func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {
	return Counter(*c).Render(ctx, w)
}

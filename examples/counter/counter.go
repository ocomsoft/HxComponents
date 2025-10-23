package counter

import (
	"context"
	"io"
)

// CounterComponent represents a simple counter that can increment/decrement.
// It demonstrates the event-driven pattern using hxc-event parameter.
type CounterComponent struct {
	Count int `form:"count"`
	// Note: No Action field needed - we use hxc-event param and event handler methods
}

// OnIncrement is an event handler that increments the counter.
// This method is called automatically when hxc-event=increment is received.
func (c *CounterComponent) OnIncrement(ctx context.Context) error {
	c.Count++
	return nil
}

// OnDecrement is an event handler that decrements the counter.
// This method is called automatically when hxc-event=decrement is received.
func (c *CounterComponent) OnDecrement(ctx context.Context) error {
	c.Count--
	return nil
}

// Render implements templ.Component interface.
// This allows the component to be used both as an HTMX component
// and as a regular templ component in templates.
func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {
	return Counter(*c).Render(ctx, w)
}

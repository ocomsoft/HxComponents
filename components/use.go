package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/a-h/templ"
)

// Use is a helper function for initializing components in templ templates.
// It takes a component and a context, checks if the component implements the
// Initializer interface, and calls Init if supported. If Init returns an error,
// it logs the error and returns the component anyway to prevent template rendering
// from failing.
//
// This function is particularly useful when you want to use components directly
// in templ templates without creating explicit constructor functions.
//
// Example usage in templ:
//
//	templ MyPage() {
//	    <div class="container">
//	        // Use a component with automatic initialization
//	        @components.Use(ctx, &card.CardComponent{
//	            Title: "My Card",
//	            Count: 42,
//	        })
//	    </div>
//	}
//
// Compared to constructor pattern:
//
//	// Instead of creating a constructor:
//	func NewCard(ctx context.Context, title string, count int) *CardComponent {
//	    c := &CardComponent{Title: title, Count: count}
//	    _ = c.Init(ctx)
//	    return c
//	}
//
//	// You can use Use() directly:
//	@components.Use(ctx, &CardComponent{Title: "My Card", Count: 42})
//
// The function will:
// 1. Check if the component implements Initializer interface
// 2. Call Init(ctx) if the interface is implemented
// 3. Log any errors that occur during initialization (using slog.Error)
// 4. Return the component for rendering (even if Init failed)
//
// Note: This function always returns the component to ensure template rendering
// doesn't fail due to initialization errors. Check your logs for any errors.
func Use[T templ.Component](ctx context.Context, component T) T {
	// Check if component implements Initializer interface
	if initializer, ok := any(component).(Initializer); ok {
		// Call Init and log any errors
		if err := initializer.Init(ctx); err != nil {
			slog.Error("component initialization error in Use()",
				"error", err,
				"component_type", typeNameOf(component))
		}
	}

	// Always return the component for rendering
	return component
}

// typeNameOf returns a human-readable type name for logging purposes
func typeNameOf(v any) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T", v)
}

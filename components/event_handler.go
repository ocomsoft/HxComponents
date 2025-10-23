package components

import "context"

// BeforeEventHandler is an optional interface that components can implement to perform
// logic before any event handler is called. This is useful for loading data, authentication,
// validation, or setting up state that all events need.
//
// BeforeEvent is called after form data is decoded and request headers are applied,
// but before the specific event handler (On{EventName}) is called.
//
// The context parameter provides request-scoped values and cancellation signals,
// which is useful for database queries, API calls, and other operations.
//
// The eventName parameter is the name of the event that will be called (e.g., "increment").
//
// Example:
//
//	func (c *MyComponent) BeforeEvent(ctx context.Context, eventName string) error {
//	    log.Printf("Event %s starting", eventName)
//
//	    // Load user from database using context
//	    user, err := db.GetUser(ctx, c.UserID)
//	    if err != nil {
//	        return fmt.Errorf("failed to load user: %w", err)
//	    }
//	    if user == nil {
//	        return fmt.Errorf("user not authenticated")
//	    }
//	    c.User = user
//	    return nil
//	}
//
// If BeforeEvent returns an error, the event handler and all subsequent processing
// (AfterEvent, Process, rendering) will be skipped and an error will be returned.
type BeforeEventHandler interface {
	BeforeEvent(ctx context.Context, eventName string) error
}

// AfterEventHandler is an optional interface that components can implement to perform
// logic after an event handler succeeds. This is useful for saving data, logging,
// cleanup, updating statistics, or triggering side effects.
//
// AfterEvent is called after the specific event handler (On{EventName}) succeeds,
// but before Process() is called.
//
// The context parameter provides request-scoped values and cancellation signals,
// which is useful for database operations, API calls, and other I/O operations.
//
// The eventName parameter is the name of the event that was just called (e.g., "increment").
//
// Example:
//
//	func (c *MyComponent) AfterEvent(ctx context.Context, eventName string) error {
//	    log.Printf("Event %s completed", eventName)
//	    c.LastAction = eventName
//	    c.UpdatedAt = time.Now()
//
//	    // Save to database using context
//	    if err := db.SaveComponent(ctx, c); err != nil {
//	        return fmt.Errorf("failed to save: %w", err)
//	    }
//	    return nil
//	}
//
// If AfterEvent returns an error, Process() and rendering will be skipped and an
// error will be returned.
type AfterEventHandler interface {
	AfterEvent(ctx context.Context, eventName string) error
}

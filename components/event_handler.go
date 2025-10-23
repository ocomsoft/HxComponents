package components

// BeforeEventHandler is an optional interface that components can implement to perform
// logic before any event handler is called. This is useful for logging, validation,
// or setting up state that all events need.
//
// BeforeEvent is called after form data is decoded and request headers are applied,
// but before the specific event handler (On{EventName}) is called.
//
// The eventName parameter is the name of the event that will be called (e.g., "increment").
//
// Example:
//
//	func (c *MyComponent) BeforeEvent(eventName string) error {
//	    log.Printf("Event %s starting", eventName)
//	    if c.UserID == "" {
//	        return fmt.Errorf("user not authenticated")
//	    }
//	    return nil
//	}
//
// If BeforeEvent returns an error, the event handler and all subsequent processing
// (AfterEvent, Process, rendering) will be skipped and an error will be returned.
type BeforeEventHandler interface {
	BeforeEvent(eventName string) error
}

// AfterEventHandler is an optional interface that components can implement to perform
// logic after an event handler succeeds. This is useful for logging, cleanup,
// updating statistics, or triggering side effects.
//
// AfterEvent is called after the specific event handler (On{EventName}) succeeds,
// but before Process() is called.
//
// The eventName parameter is the name of the event that was just called (e.g., "increment").
//
// Example:
//
//	func (c *MyComponent) AfterEvent(eventName string) error {
//	    log.Printf("Event %s completed", eventName)
//	    c.LastAction = eventName
//	    c.UpdatedAt = time.Now()
//	    return nil
//	}
//
// If AfterEvent returns an error, Process() and rendering will be skipped and an
// error will be returned.
type AfterEventHandler interface {
	AfterEvent(eventName string) error
}

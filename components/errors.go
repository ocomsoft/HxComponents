package components

import "fmt"

// ComponentError represents an error that occurred during component processing.
type ComponentError struct {
	ComponentName string
	Operation     string // "parse", "decode", "process", "render", "event"
	Err           error
	StatusCode    int
}

func (e *ComponentError) Error() string {
	return fmt.Sprintf("[%s] %s error: %v", e.ComponentName, e.Operation, e.Err)
}

func (e *ComponentError) Unwrap() error {
	return e.Err
}

// ErrComponentNotFound represents a component not found error.
type ErrComponentNotFound struct {
	ComponentName string
}

func (e *ErrComponentNotFound) Error() string {
	return fmt.Sprintf("component '%s' not found", e.ComponentName)
}

// ErrEventNotFound represents an event handler not found error.
type ErrEventNotFound struct {
	ComponentName string
	EventName     string
}

func (e *ErrEventNotFound) Error() string {
	return fmt.Sprintf("event handler '%s' not found on component '%s'", e.EventName, e.ComponentName)
}

// ErrInvalidComponentName represents an invalid component name error.
type ErrInvalidComponentName struct {
	ComponentName string
	Reason        string
}

func (e *ErrInvalidComponentName) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid component name '%s': %s", e.ComponentName, e.Reason)
	}
	return fmt.Sprintf("invalid component name '%s'", e.ComponentName)
}

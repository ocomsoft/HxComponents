package components

import (
	"context"
	"fmt"
	"reflect"
)

// SimulateEvent is a helper function for testing that simulates the complete
// component lifecycle when handling an event. This simulates what happens during
// a POST request with an hxc-event parameter.
//
// The function executes the following lifecycle steps in order:
//  1. Init - if component implements Initializer
//  2. BeforeEvent - if component implements BeforeEventHandler
//  3. On{EventName} - the event handler method
//  4. AfterEvent - if component implements AfterEventHandler
//  5. Process - if component implements Processor
//
// Parameters:
//   - ctx: The context to pass to all lifecycle methods
//   - component: The component instance to test (must be a pointer to a struct)
//   - eventName: The name of the event to trigger (e.g., "increment", "submit")
//
// Returns an error if any step in the lifecycle fails.
//
// Example usage:
//
//	func TestCounterIncrement(t *testing.T) {
//	    counter := &CounterComponent{Count: 5}
//	    ctx := context.Background()
//
//	    err := components.SimulateEvent(ctx, counter, "increment")
//	    require.NoError(t, err)
//
//	    assert.Equal(t, 6, counter.Count)
//	}
//
// Example with lifecycle tracking:
//
//	func TestEventLifecycle(t *testing.T) {
//	    component := &TestComponent{
//	        Value: 10,
//	        Log:   []string{},
//	    }
//	    ctx := context.Background()
//
//	    err := components.SimulateEvent(ctx, component, "process")
//	    require.NoError(t, err)
//
//	    // Verify lifecycle was executed in correct order
//	    expected := []string{
//	        "Init",
//	        "BeforeEvent:process",
//	        "OnProcess",
//	        "AfterEvent:process",
//	        "Process",
//	    }
//	    assert.Equal(t, expected, component.Log)
//	}
func SimulateEvent(ctx context.Context, component interface{}, eventName string) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	// Verify component is a pointer to a struct
	v := reflect.ValueOf(component)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("component must be a pointer to a struct, got %T", component)
	}
	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("component must be a pointer to a struct, got %T", component)
	}

	// Step 1: Call Init if component implements Initializer
	if initializer, ok := component.(Initializer); ok {
		if err := initializer.Init(ctx); err != nil {
			return fmt.Errorf("Init failed: %w", err)
		}
	}

	// Step 2: Call BeforeEvent if component implements BeforeEventHandler
	if beforeHandler, ok := component.(BeforeEventHandler); ok {
		if err := beforeHandler.BeforeEvent(ctx, eventName); err != nil {
			return fmt.Errorf("BeforeEvent failed: %w", err)
		}
	}

	// Step 3: Call the event handler method On{EventName}
	methodName := "On" + capitalize(eventName)
	method := v.MethodByName(methodName)

	if !method.IsValid() {
		return fmt.Errorf("event handler method '%s' not found on component %T", methodName, component)
	}

	// Validate event handler signature: On{Event}(ctx context.Context) error
	methodType := method.Type()
	if methodType.NumIn() != 1 {
		return fmt.Errorf("event handler '%s' must have signature %s(ctx context.Context) error", methodName, methodName)
	}

	// Check that first parameter is context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(0).Implements(ctxType) {
		return fmt.Errorf("event handler '%s' first parameter must be context.Context", methodName)
	}

	// Call the event handler method with context
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx)})

	// Check if method returns an error
	if len(results) > 0 {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			return fmt.Errorf("event handler failed: %w", err)
		}
	}

	// Step 4: Call AfterEvent if component implements AfterEventHandler
	if afterHandler, ok := component.(AfterEventHandler); ok {
		if err := afterHandler.AfterEvent(ctx, eventName); err != nil {
			return fmt.Errorf("AfterEvent failed: %w", err)
		}
	}

	// Step 5: Call Process if component implements Processor
	if processor, ok := component.(Processor); ok {
		if err := processor.Process(ctx); err != nil {
			return fmt.Errorf("Process failed: %w", err)
		}
	}

	return nil
}

// SimulateProcess is a helper function for testing that simulates the component
// lifecycle for a non-event request (e.g., a simple GET or POST without an event).
//
// The function executes the following lifecycle steps in order:
//  1. Init - if component implements Initializer
//  2. Process - if component implements Processor
//
// Parameters:
//   - ctx: The context to pass to all lifecycle methods
//   - component: The component instance to test (must be a pointer to a struct)
//
// Returns an error if any step in the lifecycle fails.
//
// Example usage:
//
//	func TestFormProcessing(t *testing.T) {
//	    form := &LoginForm{
//	        Username: "testuser",
//	        Password: "password123",
//	    }
//	    ctx := context.Background()
//
//	    err := components.SimulateProcess(ctx, form)
//	    require.NoError(t, err)
//
//	    assert.Equal(t, "/dashboard", form.RedirectTo)
//	}
func SimulateProcess(ctx context.Context, component interface{}) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	// Verify component is a pointer to a struct
	v := reflect.ValueOf(component)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("component must be a pointer to a struct, got %T", component)
	}
	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("component must be a pointer to a struct, got %T", component)
	}

	// Step 1: Call Init if component implements Initializer
	if initializer, ok := component.(Initializer); ok {
		if err := initializer.Init(ctx); err != nil {
			return fmt.Errorf("Init failed: %w", err)
		}
	}

	// Step 2: Call Process if component implements Processor
	if processor, ok := component.(Processor); ok {
		if err := processor.Process(ctx); err != nil {
			return fmt.Errorf("Process failed: %w", err)
		}
	}

	return nil
}

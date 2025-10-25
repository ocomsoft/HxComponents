package components

import (
	"context"
	"errors"
	"io"
	"testing"
)

// TestComponentWithInit implements both templ.Component and Initializer
type TestComponentWithInit struct {
	Value      string
	InitCalled bool
	InitError  error
}

func (c *TestComponentWithInit) Init(ctx context.Context) error {
	c.InitCalled = true
	if c.InitError != nil {
		return c.InitError
	}
	// Set a default value during init
	if c.Value == "" {
		c.Value = "initialized"
	}
	return nil
}

func (c *TestComponentWithInit) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(c.Value))
	return err
}

// TestComponentWithoutInit only implements templ.Component
type TestComponentWithoutInit struct {
	Value string
}

func (c *TestComponentWithoutInit) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(c.Value))
	return err
}

func TestUse(t *testing.T) {
	ctx := context.Background()

	t.Run("calls Init when component implements Initializer", func(t *testing.T) {
		component := &TestComponentWithInit{}

		result := Use(ctx, component)

		// Check that Init was called
		if !result.InitCalled {
			t.Error("expected Init to be called, but it wasn't")
		}

		// Check that the component is returned
		if result != component {
			t.Error("expected Use to return the same component instance")
		}

		// Check that default value was set
		if result.Value != "initialized" {
			t.Errorf("expected Value to be 'initialized', got '%s'", result.Value)
		}
	})

	t.Run("preserves existing values when Init is called", func(t *testing.T) {
		component := &TestComponentWithInit{
			Value: "custom",
		}

		result := Use(ctx, component)

		// Check that Init was called
		if !result.InitCalled {
			t.Error("expected Init to be called")
		}

		// Check that existing value is preserved
		if result.Value != "custom" {
			t.Errorf("expected Value to be 'custom', got '%s'", result.Value)
		}
	})

	t.Run("returns component even when Init returns error", func(t *testing.T) {
		expectedError := errors.New("init failed")
		component := &TestComponentWithInit{
			InitError: expectedError,
		}

		// Use should not panic even if Init returns an error
		result := Use(ctx, component)

		// Check that Init was called
		if !result.InitCalled {
			t.Error("expected Init to be called despite error")
		}

		// Check that the component is still returned
		if result != component {
			t.Error("expected Use to return the component even when Init fails")
		}

		// The error should be logged but not returned
		// (We can't easily test logging without additional infrastructure)
	})

	t.Run("does not fail when component does not implement Initializer", func(t *testing.T) {
		component := &TestComponentWithoutInit{
			Value: "no init",
		}

		// Should not panic
		result := Use(ctx, component)

		// Check that the component is returned unchanged
		if result != component {
			t.Error("expected Use to return the same component instance")
		}

		if result.Value != "no init" {
			t.Errorf("expected Value to be 'no init', got '%s'", result.Value)
		}
	})

	t.Run("works with nil context", func(t *testing.T) {
		component := &TestComponentWithInit{}

		// Should not panic with nil context
		result := Use(nil, component)

		// Init should still be called
		if !result.InitCalled {
			t.Error("expected Init to be called even with nil context")
		}

		if result != component {
			t.Error("expected Use to return the component")
		}
	})

	t.Run("preserves component type through generic return", func(t *testing.T) {
		component := &TestComponentWithInit{Value: "test"}

		// The return type should be *TestComponentWithInit, not templ.Component
		result := Use(ctx, component)

		// This should compile and work without type assertion
		if result.InitCalled != true {
			t.Error("expected to access InitCalled field directly")
		}

		// Verify the Value can be accessed directly
		if result.Value != "test" {
			t.Errorf("expected Value to be accessible, got '%s'", result.Value)
		}
	})
}

func TestTypeNameOf(t *testing.T) {
	t.Run("returns type name for struct pointer", func(t *testing.T) {
		component := &TestComponentWithInit{}
		name := typeNameOf(component)

		if name != "*components.TestComponentWithInit" {
			t.Errorf("expected type name '*components.TestComponentWithInit', got '%s'", name)
		}
	})

	t.Run("returns type name for struct", func(t *testing.T) {
		component := TestComponentWithInit{}
		name := typeNameOf(component)

		if name != "components.TestComponentWithInit" {
			t.Errorf("expected type name 'components.TestComponentWithInit', got '%s'", name)
		}
	})

	t.Run("returns <nil> for nil value", func(t *testing.T) {
		name := typeNameOf(nil)

		if name != "<nil>" {
			t.Errorf("expected '<nil>', got '%s'", name)
		}
	})
}

// Benchmark to ensure Use function has minimal overhead
func BenchmarkUse(b *testing.B) {
	ctx := context.Background()

	b.Run("with Initializer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			component := &TestComponentWithInit{Value: "test"}
			Use(ctx, component)
		}
	})

	b.Run("without Initializer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			component := &TestComponentWithoutInit{Value: "test"}
			Use(ctx, component)
		}
	})
}

package components_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/ocomsoft/HxComponents/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLifecycleComponent is a test component that tracks all lifecycle methods
type TestLifecycleComponent struct {
	Value   int      `form:"value"`
	Log     []string `json:"-"`
	InitRan bool     `json:"-"`
}

func (t *TestLifecycleComponent) Init(ctx context.Context) error {
	t.Log = append(t.Log, "Init")
	t.InitRan = true
	if t.Value == 0 {
		t.Value = 10 // Set default
	}
	return nil
}

func (t *TestLifecycleComponent) BeforeEvent(ctx context.Context, eventName string) error {
	t.Log = append(t.Log, fmt.Sprintf("BeforeEvent:%s", eventName))
	return nil
}

func (t *TestLifecycleComponent) OnIncrement(ctx context.Context) error {
	t.Log = append(t.Log, "OnIncrement")
	t.Value++
	return nil
}

func (t *TestLifecycleComponent) OnDecrement(ctx context.Context) error {
	t.Log = append(t.Log, "OnDecrement")
	t.Value--
	return nil
}

func (t *TestLifecycleComponent) OnError(ctx context.Context) error {
	t.Log = append(t.Log, "OnError")
	return fmt.Errorf("intentional error from event handler")
}

func (t *TestLifecycleComponent) AfterEvent(ctx context.Context, eventName string) error {
	t.Log = append(t.Log, fmt.Sprintf("AfterEvent:%s", eventName))
	return nil
}

func (t *TestLifecycleComponent) Process(ctx context.Context) error {
	t.Log = append(t.Log, "Process")
	return nil
}

func (t *TestLifecycleComponent) Render(ctx context.Context, w io.Writer) error {
	t.Log = append(t.Log, "Render")
	fmt.Fprintf(w, "<div>%d</div>", t.Value)
	return nil
}

// TestSimpleCounter is a minimal component for basic event testing
type TestSimpleCounter struct {
	Count int `form:"count"`
}

func (t *TestSimpleCounter) OnIncrement(ctx context.Context) error {
	t.Count++
	return nil
}

func (t *TestSimpleCounter) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "<div>%d</div>", t.Count)
	return nil
}

// TestErrorComponent tests error handling in different lifecycle phases
type TestErrorComponent struct {
	FailPhase string `json:"-"`
}

func (t *TestErrorComponent) Init(ctx context.Context) error {
	if t.FailPhase == "init" {
		return fmt.Errorf("init error")
	}
	return nil
}

func (t *TestErrorComponent) BeforeEvent(ctx context.Context, eventName string) error {
	if t.FailPhase == "before" {
		return fmt.Errorf("before event error")
	}
	return nil
}

func (t *TestErrorComponent) OnTest(ctx context.Context) error {
	if t.FailPhase == "event" {
		return fmt.Errorf("event handler error")
	}
	return nil
}

func (t *TestErrorComponent) AfterEvent(ctx context.Context, eventName string) error {
	if t.FailPhase == "after" {
		return fmt.Errorf("after event error")
	}
	return nil
}

func (t *TestErrorComponent) Process(ctx context.Context) error {
	if t.FailPhase == "process" {
		return fmt.Errorf("process error")
	}
	return nil
}

func (t *TestErrorComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprint(w, "<div>Test</div>")
	return nil
}

func TestSimulateEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("executes full lifecycle in correct order", func(t *testing.T) {
		component := &TestLifecycleComponent{Value: 5}

		err := components.SimulateEvent(ctx, component, "increment")
		require.NoError(t, err)

		// Verify value was incremented
		assert.Equal(t, 6, component.Value)

		// Verify lifecycle executed in correct order
		expected := []string{
			"Init",
			"BeforeEvent:increment",
			"OnIncrement",
			"AfterEvent:increment",
			"Process",
		}
		assert.Equal(t, expected, component.Log)
	})

	t.Run("works with simple component without lifecycle hooks", func(t *testing.T) {
		counter := &TestSimpleCounter{Count: 10}

		err := components.SimulateEvent(ctx, counter, "increment")
		require.NoError(t, err)

		assert.Equal(t, 11, counter.Count)
	})

	t.Run("handles Init with default values", func(t *testing.T) {
		component := &TestLifecycleComponent{} // No value set

		err := components.SimulateEvent(ctx, component, "increment")
		require.NoError(t, err)

		// Init should set default to 10, then increment to 11
		assert.Equal(t, 11, component.Value)
		assert.True(t, component.InitRan)
	})

	t.Run("handles decrement event", func(t *testing.T) {
		component := &TestLifecycleComponent{Value: 20}

		err := components.SimulateEvent(ctx, component, "decrement")
		require.NoError(t, err)

		assert.Equal(t, 19, component.Value)
		assert.Contains(t, component.Log, "OnDecrement")
	})
}

func TestSimulateEventErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when component is nil", func(t *testing.T) {
		err := components.SimulateEvent(ctx, nil, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "component cannot be nil")
	})

	t.Run("returns error when component is not a pointer", func(t *testing.T) {
		component := TestSimpleCounter{Count: 5}
		err := components.SimulateEvent(ctx, component, "increment")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer to a struct")
	})

	t.Run("returns error when event handler does not exist", func(t *testing.T) {
		component := &TestSimpleCounter{Count: 5}
		err := components.SimulateEvent(ctx, component, "nonExistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "OnNonExistent")
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when Init fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "init"}
		err := components.SimulateEvent(ctx, component, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Init failed")
	})

	t.Run("returns error when BeforeEvent fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "before"}
		err := components.SimulateEvent(ctx, component, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BeforeEvent failed")
	})

	t.Run("returns error when event handler fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "event"}
		err := components.SimulateEvent(ctx, component, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event handler failed")
	})

	t.Run("returns error when AfterEvent fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "after"}
		err := components.SimulateEvent(ctx, component, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AfterEvent failed")
	})

	t.Run("returns error when Process fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "process"}
		err := components.SimulateEvent(ctx, component, "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Process failed")
	})

	t.Run("stops at event handler error", func(t *testing.T) {
		component := &TestLifecycleComponent{}
		err := components.SimulateEvent(ctx, component, "error")
		require.Error(t, err)

		// Process should not have been called
		assert.NotContains(t, component.Log, "Process")
		// But BeforeEvent should have been called
		assert.Contains(t, component.Log, "BeforeEvent:error")
	})
}

func TestSimulateProcess(t *testing.T) {
	ctx := context.Background()

	t.Run("executes Init and Process", func(t *testing.T) {
		component := &TestLifecycleComponent{Value: 0}

		err := components.SimulateProcess(ctx, component)
		require.NoError(t, err)

		// Init should have set default value
		assert.Equal(t, 10, component.Value)
		assert.True(t, component.InitRan)

		// Should have called Init and Process but not event handlers
		expected := []string{"Init", "Process"}
		assert.Equal(t, expected, component.Log)
	})

	t.Run("works with component without Init or Process", func(t *testing.T) {
		counter := &TestSimpleCounter{Count: 5}

		err := components.SimulateProcess(ctx, counter)
		require.NoError(t, err)

		// Should not change anything
		assert.Equal(t, 5, counter.Count)
	})

	t.Run("returns error when component is nil", func(t *testing.T) {
		err := components.SimulateProcess(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "component cannot be nil")
	})

	t.Run("returns error when Init fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "init"}
		err := components.SimulateProcess(ctx, component)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Init failed")
	})

	t.Run("returns error when Process fails", func(t *testing.T) {
		component := &TestErrorComponent{FailPhase: "process"}
		err := components.SimulateProcess(ctx, component)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Process failed")
	})
}

// Example test showing real-world usage pattern
func TestCounterComponent(t *testing.T) {
	ctx := context.Background()

	t.Run("increment increases count", func(t *testing.T) {
		counter := &TestSimpleCounter{Count: 0}

		// Simulate user clicking increment button
		err := components.SimulateEvent(ctx, counter, "increment")
		require.NoError(t, err)

		assert.Equal(t, 1, counter.Count)
	})

	t.Run("multiple increments", func(t *testing.T) {
		counter := &TestSimpleCounter{Count: 0}

		// Simulate multiple clicks
		for i := 0; i < 5; i++ {
			err := components.SimulateEvent(ctx, counter, "increment")
			require.NoError(t, err)
		}

		assert.Equal(t, 5, counter.Count)
	})
}

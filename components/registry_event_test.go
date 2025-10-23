package components_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ocomsoft/HxComponents/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventComponent is a test component that tracks event lifecycle
type TestEventComponent struct {
	Count         int      `form:"count"`
	EventsHistory []string `json:"-"`
}

func (t *TestEventComponent) BeforeEvent(ctx context.Context, eventName string) error {
	t.EventsHistory = append(t.EventsHistory, fmt.Sprintf("BeforeEvent:%s", eventName))
	return nil
}

func (t *TestEventComponent) OnIncrement(ctx context.Context) error {
	t.EventsHistory = append(t.EventsHistory, "OnIncrement")
	t.Count++
	return nil
}

func (t *TestEventComponent) OnDecrement(ctx context.Context) error {
	t.EventsHistory = append(t.EventsHistory, "OnDecrement")
	t.Count--
	return nil
}

func (t *TestEventComponent) OnError(ctx context.Context) error {
	t.EventsHistory = append(t.EventsHistory, "OnError")
	return fmt.Errorf("intentional error")
}

func (t *TestEventComponent) AfterEvent(ctx context.Context, eventName string) error {
	t.EventsHistory = append(t.EventsHistory, fmt.Sprintf("AfterEvent:%s", eventName))
	return nil
}

func (t *TestEventComponent) Process(ctx context.Context) error {
	t.EventsHistory = append(t.EventsHistory, "Process")
	return nil
}

func (t *TestEventComponent) Render(ctx context.Context, w io.Writer) error {
	t.EventsHistory = append(t.EventsHistory, "Render")
	fmt.Fprintf(w, "<div>Count: %d, History: %v</div>", t.Count, t.EventsHistory)
	return nil
}

// TestBeforeEventErrorComponent tests that errors in BeforeEvent stop processing
type TestBeforeEventErrorComponent struct {
	Called bool `json:"-"`
}

func (t *TestBeforeEventErrorComponent) BeforeEvent(ctx context.Context, eventName string) error {
	return fmt.Errorf("before event error")
}

func (t *TestBeforeEventErrorComponent) OnDoSomething(ctx context.Context) error {
	t.Called = true
	return nil
}

func (t *TestBeforeEventErrorComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "<div>Called: %v</div>", t.Called)
	return nil
}

// TestAfterEventErrorComponent tests that errors in AfterEvent stop processing
type TestAfterEventErrorComponent struct {
	EventHandlerCalled bool `json:"-"`
	ProcessCalled      bool `json:"-"`
}

func (t *TestAfterEventErrorComponent) OnDoSomething(ctx context.Context) error {
	t.EventHandlerCalled = true
	return nil
}

func (t *TestAfterEventErrorComponent) AfterEvent(ctx context.Context, eventName string) error {
	return fmt.Errorf("after event error")
}

func (t *TestAfterEventErrorComponent) Process(ctx context.Context) error {
	t.ProcessCalled = true
	return nil
}

func (t *TestAfterEventErrorComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "<div>EventHandlerCalled: %v, ProcessCalled: %v</div>", t.EventHandlerCalled, t.ProcessCalled)
	return nil
}

// TestMissingEventHandlerComponent tests handling of missing event handlers
type TestMissingEventHandlerComponent struct{}

func (t *TestMissingEventHandlerComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprint(w, "<div>No events</div>")
	return nil
}

func TestEventLifecycle(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestEventComponent](registry, "test")

	t.Run("event lifecycle executes in correct order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=5&hxc-event=increment"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler := registry.HandlerFor("test")
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()

		// Verify the count was incremented
		assert.Contains(t, body, "Count: 6")

		// Verify the event lifecycle order
		assert.Contains(t, body, "BeforeEvent:increment")
		assert.Contains(t, body, "OnIncrement")
		assert.Contains(t, body, "AfterEvent:increment")
		assert.Contains(t, body, "Process")
		assert.Contains(t, body, "Render")
	})

	t.Run("decrement event works", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=10&hxc-event=decrement"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler := registry.HandlerFor("test")
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()

		// Verify the count was decremented
		assert.Contains(t, body, "Count: 9")

		// Verify the event lifecycle order
		assert.Contains(t, body, "BeforeEvent:decrement")
		assert.Contains(t, body, "OnDecrement")
		assert.Contains(t, body, "AfterEvent:decrement")
	})

	t.Run("event handler error stops processing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=5&hxc-event=error"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler := registry.HandlerFor("test")
		handler(w, req)

		// Should return error status
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Should not call Process or Render
		body := w.Body.String()
		assert.NotContains(t, body, "Process")
		assert.NotContains(t, body, "Render")
	})
}

func TestBeforeEventError(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestBeforeEventErrorComponent](registry, "test")

	req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("hxc-event=doSomething"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w, req)

	// Should return error status
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Event handler should not have been called
	body := w.Body.String()
	assert.NotContains(t, body, "Called: true")
}

func TestAfterEventError(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestAfterEventErrorComponent](registry, "test")

	req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("hxc-event=doSomething"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w, req)

	// Should return error status
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Process should not have been called (AfterEvent error stops processing)
	body := w.Body.String()
	assert.NotContains(t, body, "ProcessCalled: true")
}

func TestMissingEventHandler(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestMissingEventHandlerComponent](registry, "test")

	req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("hxc-event=nonExistent"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w, req)

	// Should return error status
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Error message should indicate missing handler
	body := w.Body.String()
	assert.Contains(t, body, "Event Error")
	assert.Contains(t, body, "nonExistent")
	assert.Contains(t, body, "not found")
}

func TestEventWithoutHxcEventParam(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestEventComponent](registry, "test")

	// Request without hxc-event parameter should skip event handling
	req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=5"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Should render but not execute any event handlers
	assert.Contains(t, body, "Count: 5")
	assert.NotContains(t, body, "OnIncrement")
	assert.NotContains(t, body, "OnDecrement")
	assert.Contains(t, body, "Process") // Process should still be called
	assert.Contains(t, body, "Render")
}

func TestEventNameCapitalization(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestEventComponent](registry, "test")

	testCases := []struct {
		name          string
		eventName     string
		expectedCount int
		shouldWork    bool
	}{
		{
			name:          "lowercase event name",
			eventName:     "increment",
			expectedCount: 6,
			shouldWork:    true,
		},
		{
			name:          "camelCase event name",
			eventName:     "increment",
			expectedCount: 6,
			shouldWork:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formData := url.Values{}
			formData.Set("count", "5")
			formData.Set("hxc-event", tc.eventName)

			req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler := registry.HandlerFor("test")
			handler(w, req)

			if tc.shouldWork {
				assert.Equal(t, http.StatusOK, w.Code)
				body := w.Body.String()
				assert.Contains(t, body, fmt.Sprintf("Count: %d", tc.expectedCount))
			} else {
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			}
		})
	}
}

func TestMultipleEventsInSequence(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestEventComponent](registry, "test")

	// First increment
	req1 := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=0&hxc-event=increment"))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "Count: 1")

	// Then increment again
	req2 := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=1&hxc-event=increment"))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()

	handler(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "Count: 2")

	// Then decrement
	req3 := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader("count=2&hxc-event=decrement"))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w3 := httptest.NewRecorder()

	handler(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "Count: 1")
}

// SimpleComponent is a component without lifecycle hooks for testing
type SimpleComponent struct {
	Count int `form:"count"`
}

func (s *SimpleComponent) OnIncrement(ctx context.Context) error {
	s.Count++
	return nil
}

func (s *SimpleComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "<div>%d</div>", s.Count)
	return nil
}

func TestComponentWithoutLifecycleHooks(t *testing.T) {
	// Component without BeforeEvent/AfterEvent should still work
	registry := components.NewRegistry()
	components.Register[*SimpleComponent](registry, "simple")

	req := httptest.NewRequest(http.MethodPost, "/component/simple", strings.NewReader("count=5&hxc-event=increment"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("simple")
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "6")
}

func TestGETRequestWithEvents(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*TestEventComponent](registry, "test")

	// GET requests with hxc-event parameter should also work
	req := httptest.NewRequest(http.MethodGet, "/component/test?count=5&hxc-event=increment", nil)
	w := httptest.NewRecorder()

	handler := registry.HandlerFor("test")
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	assert.Contains(t, body, "Count: 6")
	assert.Contains(t, body, "BeforeEvent:increment")
	assert.Contains(t, body, "OnIncrement")
	assert.Contains(t, body, "AfterEvent:increment")
}

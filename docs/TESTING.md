# Testing HxComponents

This guide covers strategies for testing HxComponents applications at multiple levels.

## Test Helpers

HxComponents provides built-in test helpers to simulate the component lifecycle without needing HTTP requests. These helpers are especially useful for unit testing components in isolation.

### SimulateEvent

The `SimulateEvent` helper simulates a complete event lifecycle, calling Init, BeforeEvent, the event handler, AfterEvent, and Process in the correct order. This is perfect for testing event handlers without setting up HTTP requests.

```go
package counter_test

import (
	"context"
	"testing"

	"github.com/ocomsoft/HxComponents/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterIncrement(t *testing.T) {
	// Create component
	counter := &counter.CounterComponent{Count: 5}
	ctx := context.Background()

	// Simulate increment event
	err := components.SimulateEvent(ctx, counter, "increment")
	require.NoError(t, err)

	// Assert the count was incremented
	assert.Equal(t, 6, counter.Count)
}

func TestMultipleEvents(t *testing.T) {
	counter := &counter.CounterComponent{Count: 0}
	ctx := context.Background()

	// Simulate multiple clicks
	for i := 0; i < 5; i++ {
		err := components.SimulateEvent(ctx, counter, "increment")
		require.NoError(t, err)
	}

	assert.Equal(t, 5, counter.Count)
}
```

**Lifecycle executed by SimulateEvent:**
1. `Init(ctx)` - if component implements `Initializer`
2. `BeforeEvent(ctx, eventName)` - if component implements `BeforeEventHandler`
3. `On{EventName}(ctx)` - the event handler method
4. `AfterEvent(ctx, eventName)` - if component implements `AfterEventHandler`
5. `Process(ctx)` - if component implements `Processor`

### SimulateProcess

The `SimulateProcess` helper simulates a non-event request (e.g., a simple GET or POST without an event). It calls Init and Process only.

```go
func TestFormProcessing(t *testing.T) {
	form := &login.LoginComponent{
		Username: "testuser",
		Password: "password123",
	}
	ctx := context.Background()

	err := components.SimulateProcess(ctx, form)
	require.NoError(t, err)

	// Assert redirect was set
	assert.Equal(t, "/dashboard", form.RedirectTo)
}
```

**Lifecycle executed by SimulateProcess:**
1. `Init(ctx)` - if component implements `Initializer`
2. `Process(ctx)` - if component implements `Processor`

### Testing Lifecycle Hooks with Helpers

The test helpers make it easy to verify that lifecycle hooks are called in the correct order:

```go
type TestComponent struct {
	Value int
	Log   []string
}

func (t *TestComponent) Init(ctx context.Context) error {
	t.Log = append(t.Log, "Init")
	if t.Value == 0 {
		t.Value = 10
	}
	return nil
}

func (t *TestComponent) BeforeEvent(ctx context.Context, eventName string) error {
	t.Log = append(t.Log, fmt.Sprintf("BeforeEvent:%s", eventName))
	return nil
}

func (t *TestComponent) OnProcess(ctx context.Context) error {
	t.Log = append(t.Log, "OnProcess")
	t.Value++
	return nil
}

func (t *TestComponent) AfterEvent(ctx context.Context, eventName string) error {
	t.Log = append(t.Log, fmt.Sprintf("AfterEvent:%s", eventName))
	return nil
}

func (t *TestComponent) Process(ctx context.Context) error {
	t.Log = append(t.Log, "Process")
	return nil
}

func (t *TestComponent) Render(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, "<div>%d</div>", t.Value)
	return nil
}

func TestLifecycleOrder(t *testing.T) {
	component := &TestComponent{Value: 0}
	ctx := context.Background()

	err := components.SimulateEvent(ctx, component, "process")
	require.NoError(t, err)

	// Verify lifecycle was executed in correct order
	expected := []string{
		"Init",
		"BeforeEvent:process",
		"OnProcess",
		"AfterEvent:process",
		"Process",
	}
	assert.Equal(t, expected, component.Log)

	// Verify Init set default value, then OnProcess incremented it
	assert.Equal(t, 11, component.Value)
}
```

### Error Handling in Test Helpers

The test helpers properly handle errors at each lifecycle stage:

```go
func TestErrorInBeforeEvent(t *testing.T) {
	component := &MyComponent{
		FailPhase: "before",
	}
	ctx := context.Background()

	err := components.SimulateEvent(ctx, component, "submit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BeforeEvent failed")
}

func TestErrorInEventHandler(t *testing.T) {
	component := &MyComponent{
		FailPhase: "event",
	}
	ctx := context.Background()

	err := components.SimulateEvent(ctx, component, "submit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event handler failed")

	// Process should not have been called
	assert.False(t, component.ProcessCalled)
}
```

### When to Use Test Helpers vs HTTP Tests

**Use SimulateEvent/SimulateProcess when:**
- Testing component logic in isolation
- Verifying lifecycle hook order
- Running fast unit tests
- Testing error conditions
- You don't need to test HTTP-specific behavior

**Use HTTP tests (httptest) when:**
- Testing form parsing and decoding
- Verifying HTTP headers (HTMX headers, redirects)
- Testing the full HTTP request/response cycle
- Integration testing with the registry

```go
// Fast unit test - uses SimulateEvent
func TestCounterLogic(t *testing.T) {
	counter := &counter.CounterComponent{Count: 5}
	err := components.SimulateEvent(context.Background(), counter, "increment")
	require.NoError(t, err)
	assert.Equal(t, 6, counter.Count)
}

// Integration test - uses httptest
func TestCounterHTTP(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*counter.CounterComponent](registry, "counter")

	form := url.Values{}
	form.Add("count", "5")
	form.Add("hxc-event", "increment")

	req := httptest.NewRequest(http.MethodPost, "/component/counter",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()
	registry.Handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "6")
}
```

## Unit Tests for Component Methods

Test component logic in isolation:

```go
package counter_test

import (
	"testing"

	"myproject/components/counter"
	"github.com/stretchr/testify/assert"
)

func TestCounterIncrement(t *testing.T) {
	// Create component
	c := &counter.CounterComponent{Count: 0}

	// Call event handler
	err := c.OnIncrement()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, c.Count)
}

func TestCounterDecrement(t *testing.T) {
	c := &counter.CounterComponent{Count: 5}
	err := c.OnDecrement()

	assert.NoError(t, err)
	assert.Equal(t, 4, c.Count)
}

func TestCounterDoubled(t *testing.T) {
	c := &counter.CounterComponent{Count: 7}

	assert.Equal(t, 14, c.Doubled())
}
```

## Integration Tests with httptest

Test HTTP handlers and component registration:

```go
package counter_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ocomsoft/HxComponents/components"
	"myproject/components/counter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterHTTPHandler(t *testing.T) {
	// Setup registry
	registry := components.NewRegistry()
	components.Register[*counter.CounterComponent](registry, "counter")

	// Test POST request
	form := url.Values{}
	form.Add("count", "5")
	form.Add("hxc-event", "increment")

	req := httptest.NewRequest(http.MethodPost, "/component/counter",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	registry.Handler(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "6") // Count should be 6
}

func TestCounterGETRequest(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*counter.CounterComponent](registry, "counter")

	// Test GET request with query parameters
	req := httptest.NewRequest(http.MethodGet,
		"/component/counter?count=10", nil)

	w := httptest.NewRecorder()
	registry.Handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "10")
}
```

## End-to-End Tests with Playwright

Test complete user workflows in a real browser:

```go
package counter_test

import (
	"testing"

	"myproject/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterE2E(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("counter increments on button click", func(t *testing.T) {
		// Navigate to page
		pt.Goto(server.URL)

		// Find counter
		counter := pt.Page.Locator(".counter")
		span := counter.Locator("span")

		// Verify initial value
		text, err := span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "0", text)

		// Click increment button
		incrementBtn := counter.Locator("button:has-text('+')")
		err = incrementBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify updated value
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "1", text)
	})

	t.Run("counter handles rapid clicks", func(t *testing.T) {
		pt.Goto(server.URL)

		counter := pt.Page.Locator(".counter")
		incrementBtn := counter.Locator("button:has-text('+')")

		// Click multiple times rapidly
		for i := 0; i < 5; i++ {
			incrementBtn.Click()
			pt.WaitForHTMX()
		}

		span := counter.Locator("span")
		text, _ := span.TextContent()
		assert.Equal(t, "5", text)
	})
}
```

## Testing Lifecycle Hooks

```go
func TestComponentLifecycle(t *testing.T) {
	c := &todolist.TodoListComponent{
		Items: []todolist.TodoItem{
			{ID: 1, Text: "Test", Completed: false},
		},
	}

	// Test BeforeEvent
	ctx := context.Background()
	err := c.BeforeEvent(ctx, "addItem")
	assert.NoError(t, err)

	// Test event handler
	c.NewItemText = "New item"
	err = c.OnAddItem()
	assert.NoError(t, err)
	assert.Len(t, c.Items, 2)
	assert.Equal(t, "", c.NewItemText) // Should be cleared

	// Test AfterEvent
	err = c.AfterEvent(ctx, "addItem")
	assert.NoError(t, err)
	assert.Equal(t, "addItem", c.LastEvent)
	assert.Equal(t, 1, c.EventCount)
}
```

## Testing Validation

```go
func TestValidationErrors(t *testing.T) {
	form := &userform.UserFormComponent{
		Email:    "",
		Password: "short",
	}

	// Trigger validation
	ctx := context.Background()
	err := form.BeforeEvent(ctx, "submit")
	assert.NoError(t, err) // BeforeEvent shouldn't error

	// Check validation errors were set
	assert.NotEmpty(t, form.EmailError)
	assert.NotEmpty(t, form.PasswordError)
	assert.False(t, form.IsValid())

	// OnSubmit should fail
	err = form.OnSubmit()
	assert.Error(t, err)

	// Fix validation
	form.Email = "test@example.com"
	form.Password = "password123"
	err = form.BeforeEvent(ctx, "submit")
	assert.NoError(t, err)
	assert.Empty(t, form.EmailError)
	assert.Empty(t, form.PasswordError)
	assert.True(t, form.IsValid())
}
```

## Testing Error Handling

```go
func TestErrorHandling(t *testing.T) {
	c := &todolist.TodoListComponent{}

	// Test adding empty item
	c.NewItemText = ""
	err := c.OnAddItem()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	// Test deleting non-existent item
	c.ItemID = 999
	err = c.OnDeleteItem()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
```

## Testing HTMX Headers

```go
func TestHTMXHeaders(t *testing.T) {
	registry := components.NewRegistry()
	components.Register[*login.LoginComponent](registry, "login")

	form := url.Values{}
	form.Add("username", "demo")
	form.Add("password", "password")
	form.Add("hxc-event", "submit")

	req := httptest.NewRequest(http.MethodPost, "/component/login",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()
	registry.Handler(w, req)

	// Check response headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/dashboard", w.Header().Get("HX-Redirect"))
}
```

## Testing Response Headers

```go
func TestResponseHeaders(t *testing.T) {
	c := &mycomponent.MyComponent{}

	// Trigger event that sets response headers
	err := c.OnSubmit()
	assert.NoError(t, err)

	// Check response headers were set
	headers := c.GetHTMXResponseHeaders()
	assert.Equal(t, "/success", headers["HX-Redirect"])
	assert.Equal(t, "itemUpdated", headers["HX-Trigger"])
}
```

## Testing Component Rendering

```go
func TestComponentRendering(t *testing.T) {
	c := &counter.CounterComponent{Count: 42}

	// Render to buffer
	var buf bytes.Buffer
	err := c.Render(context.Background(), &buf)

	assert.NoError(t, err)
	html := buf.String()

	// Check rendered output
	assert.Contains(t, html, "42")
	assert.Contains(t, html, "hx-post=\"/component/counter\"")
	assert.Contains(t, html, "hxc-event")
}
```

## Testing with Table-Driven Tests

```go
func TestCounter TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initial       int
		operation     string
		expected      int
		shouldError   bool
	}{
		{"increment from zero", 0, "increment", 1, false},
		{"decrement from zero", 0, "decrement", -1, false},
		{"increment from positive", 5, "increment", 6, false},
		{"decrement from positive", 5, "decrement", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter.CounterComponent{Count: tt.initial}

			var err error
			if tt.operation == "increment" {
				err = c.OnIncrement()
			} else {
				err = c.OnDecrement()
			}

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, c.Count)
			}
		})
	}
}
```

## Test Utilities

Create helper functions for common test scenarios:

```go
package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ocomsoft/HxComponents/components"
	"github.com/playwright-community/playwright-go"
)

type TestServer struct {
	*httptest.Server
	Registry *components.Registry
}

func NewTestServer(t *testing.T) *TestServer {
	registry := components.NewRegistry()

	// Register all components
	// components.Register[*counter.CounterComponent](registry, "counter")
	// ...

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/component/*", registry.Handler)
	router.Post("/component/*", registry.Handler)

	// Add page routes
	// ...

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &TestServer{
		Server:   server,
		Registry: registry,
	}
}

type PlaywrightTest struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	Page    playwright.Page
}

func NewPlaywrightTest(t *testing.T) *PlaywrightTest {
	pw, err := playwright.Run()
	if err != nil {
		t.Fatal(err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatal(err)
	}

	page, err := browser.NewPage()
	if err != nil {
		t.Fatal(err)
	}

	pt := &PlaywrightTest{
		pw:      pw,
		browser: browser,
		Page:    page,
	}

	t.Cleanup(func() {
		page.Close()
		browser.Close()
		pw.Stop()
	})

	return pt
}

func (pt *PlaywrightTest) Goto(url string) {
	if _, err := pt.Page.Goto(url); err != nil {
		panic(err)
	}
}

func (pt *PlaywrightTest) WaitForHTMX() {
	// Wait for htmx requests to complete
	pt.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	time.Sleep(100 * time.Millisecond) // Small buffer for HTMX swap
}

func (pt *PlaywrightTest) Close() {
	pt.Page.Close()
	pt.browser.Close()
	pt.pw.Stop()
}
```

## Best Practices

1. **Test at Multiple Levels**
   - Unit tests for component logic
   - Integration tests for HTTP handlers
   - E2E tests for critical user workflows

2. **Use Table-Driven Tests**
   - Test multiple scenarios efficiently
   - Easy to add new test cases

3. **Test Error Paths**
   - Test validation failures
   - Test invalid inputs
   - Test error recovery

4. **Mock External Dependencies**
   - Mock database calls
   - Mock external APIs
   - Use test doubles for services

5. **Test Lifecycle Hooks**
   - Ensure BeforeEvent/AfterEvent work correctly
   - Test state persistence
   - Test side effects

6. **Test HTMX Behavior**
   - Verify HTMX headers
   - Test response headers
   - Test swap behavior

7. **Use Test Fixtures**
   - Create reusable test data
   - Use factories for complex objects

8. **Keep Tests Fast**
   - Use unit tests for business logic
   - Reserve E2E tests for critical paths
   - Run tests in parallel when possible

9. **Test Accessibility**
   - Use Playwright's accessibility testing
   - Check ARIA attributes
   - Test keyboard navigation

10. **Continuous Integration**
    - Run tests on every commit
    - Use GitHub Actions or similar
    - Generate coverage reports

# HxComponents Troubleshooting Guide

This guide covers common issues you might encounter when working with HxComponents and how to resolve them.

## Table of Contents

- [Component Not Rendering](#component-not-rendering)
- [Form Fields Empty](#form-fields-empty)
- [Events Not Firing](#events-not-firing)
- [Context Errors](#context-errors)
- [Registration Panics](#registration-panics)
- [HTMX Headers Not Working](#htmx-headers-not-working)
- [Performance Issues](#performance-issues)
- [Debugging Tips](#debugging-tips)

---

## Component Not Rendering

### Symptoms
- Empty response or blank page
- 404 Not Found error
- Component doesn't appear on page

### Solutions

#### 1. Check Component Registration
```go
// Ensure component is registered before starting server
registry := components.NewRegistry()
components.Register[*MyComponent](registry, "mycomponent")
```

#### 2. Verify URL Matches Component Name
```go
// Registration name must match URL
components.Register[*MyComponent](registry, "mycomponent")

// URL must be: /component/mycomponent
// NOT: /component/MyComponent (case sensitive!)
```

#### 3. Run `templ generate`
```bash
# Templates must be generated before running
templ generate

# Or watch for changes
templ generate --watch
```

#### 4. Check Component Implements templ.Component
```go
// Component MUST implement templ.Component interface
func (c *MyComponent) Render(ctx context.Context, w io.Writer) error {
    return MyTemplate(*c).Render(ctx, w)
}
```

#### 5. Enable Debug Mode
```go
registry := components.NewRegistry()
registry.EnableDebugMode() // Add debugging headers

// Check browser dev tools Network tab for:
// - X-HxComponent-Name
// - X-HxComponent-FormFields
// - X-HxComponent-HasEvent
```

---

## Form Fields Empty

### Symptoms
- Struct fields are zero values after form submission
- Form data not being decoded into component

### Solutions

#### 1. Check Form Tags Match HTML Names
```go
type MyForm struct {
    Email string `form:"email"` // Must match <input name="email">
    Name  string `form:"name"`  // Must match <input name="name">
}
```

```html
<!-- HTML input names must match form tags exactly -->
<input name="email" type="email" />
<input name="name" type="text" />
```

#### 2. Verify Content-Type Header
```html
<!-- For POST requests, ensure proper content type -->
<form hx-post="/component/myform"
      hx-headers='{"Content-Type": "application/x-www-form-urlencoded"}'>
    <input name="email" />
</form>
```

#### 3. Check Form Method (POST vs GET)
```go
// For POST: data comes from request body
// For GET: data comes from query parameters

// Both work, but ensure your HTML uses the right method:
<form hx-post="/component/myform">  <!-- POST -->
<div hx-get="/component/myform?email=test@example.com">  <!-- GET -->
```

#### 4. Enable Debug Logging
```go
import "log/slog"

// Set log level to debug
slog.SetLogLoggerLevel(slog.LevelDebug)

// Check logs for "form decode error" messages
```

#### 5. Inspect Form Data
```go
// In your Process method, log the received data:
func (c *MyComponent) Process(ctx context.Context) error {
    slog.Debug("received form data",
        "email", c.Email,
        "name", c.Name)
    return nil
}
```

---

## Events Not Firing

### Symptoms
- Event handler method not being called
- No errors in logs
- Button clicks don't do anything

### Solutions

#### 1. Check Method Name Matches Event
```go
// Event name in HTML: "increment"
// Method name MUST be: "OnIncrement" (capitalize first letter)
func (c *CounterComponent) OnIncrement() error {
    c.Count++
    return nil
}
```

```html
<!-- Event name must match method (case-insensitive) -->
<button hx-vals='{"hxc-event": "increment"}'>+</button>
```

#### 2. Verify hxc-event Parameter
```html
<!-- MUST include hxc-event in hx-vals -->
<button hx-post="/component/counter"
        hx-vals='{"count": 5, "hxc-event": "increment"}'>
    +
</button>

<!-- Or as hidden field -->
<form hx-post="/component/counter">
    <input type="hidden" name="hxc-event" value="increment" />
    <button type="submit">Submit</button>
</form>
```

#### 3. Check Method Signature
```go
// Correct signature:
func (c *Component) OnMyEvent() error {
    return nil
}

// WRONG - no parameters allowed (except receiver):
func (c *Component) OnMyEvent(ctx context.Context) error { // ❌
    return nil
}
```

#### 4. Ensure Event Method is Exported
```go
// Correct - starts with capital letter:
func (c *Component) OnIncrement() error { // ✅

// Wrong - lowercase means unexported:
func (c *Component) onIncrement() error { // ❌
```

#### 5. Use Debug Mode to See Events
```go
registry.EnableDebugMode()

// Check response headers:
// X-HxComponent-HasEvent: true/false
```

---

## Context Errors

### Symptoms
- "context deadline exceeded" error
- "context canceled" error
- Database timeouts

### Solutions

#### 1. Check Database Query Timeouts
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
    // Add timeout to context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    items, err := db.GetItems(ctx, c.UserID)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("database query timeout")
        }
        return err
    }
    c.Items = items
    return nil
}
```

#### 2. Respect Context Cancellation
```go
func (c *Component) Process(ctx context.Context) error {
    for i := 0; i < 1000; i++ {
        // Check if context was cancelled
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue processing
        }

        // Do work...
    }
    return nil
}
```

#### 3. Don't Create New Context
```go
// WRONG - don't create background context:
func (c *Component) Process(ctx context.Context) error {
    newCtx := context.Background() // ❌ Loses cancellation
    db.Query(newCtx, "...")
    return nil
}

// Correct - use provided context:
func (c *Component) Process(ctx context.Context) error {
    db.Query(ctx, "...") // ✅
    return nil
}
```

---

## Registration Panics

### Symptoms
- Panic on startup: "component already registered"
- Panic: "component type must be a pointer type"
- Panic: "component must point to a struct"

### Solutions

#### 1. Don't Register Same Component Twice
```go
// WRONG - duplicate registration:
components.Register[*MyComponent](registry, "mycomponent")
components.Register[*MyComponent](registry, "mycomponent") // ❌ Panic!

// Correct - register once:
components.Register[*MyComponent](registry, "mycomponent")
```

#### 2. Use Pointer Type
```go
// WRONG - not a pointer:
components.Register[MyComponent](registry, "mycomponent") // ❌

// Correct - must be pointer:
components.Register[*MyComponent](registry, "mycomponent") // ✅
```

#### 3. Ensure Type is a Struct
```go
// WRONG - can't register primitive types:
components.Register[*string](registry, "test") // ❌

// Correct - must be a struct:
type MyComponent struct {
    Data string
}
components.Register[*MyComponent](registry, "test") // ✅
```

---

## HTMX Headers Not Working

### Symptoms
- HX-Redirect not redirecting
- HX-Trigger not firing events
- Request headers not being captured

### Solutions

#### 1. Implement Response Header Interfaces
```go
type LoginComponent struct {
    RedirectTo string `json:"-"`
}

// Must implement interface to set header:
func (c *LoginComponent) GetHxRedirect() string {
    return c.RedirectTo
}
```

#### 2. Implement Request Header Interfaces
```go
type MyComponent struct {
    IsBoosted bool `json:"-"`
}

// Must implement setter interface:
func (c *MyComponent) SetHxBoosted(v bool) {
    c.IsBoosted = v
}
```

#### 3. Set Values in Process Method
```go
func (c *LoginComponent) Process(ctx context.Context) error {
    if c.Username == "demo" && c.Password == "password" {
        c.RedirectTo = "/dashboard" // This will set HX-Redirect header
        return nil
    }
    return nil
}
```

#### 4. Check Header Names
```go
// Response headers must have exact getter names:
GetHxRedirect()       // Sets: HX-Redirect
GetHxTrigger()        // Sets: HX-Trigger
GetHxPushUrl()        // Sets: HX-Push-Url

// Request headers must have exact setter names:
SetHxBoosted(bool)    // Reads: HX-Boosted
SetHxRequest(bool)    // Reads: HX-Request
SetHxCurrentURL(string) // Reads: HX-Current-URL
```

---

## Performance Issues

### Symptoms
- Slow component rendering
- High memory usage
- Database connection pool exhaustion

### Solutions

#### 1. Use Context for Timeouts
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // Query with timeout
    data, err := db.Query(ctx, "...")
    return err
}
```

#### 2. Limit Data Loaded
```go
// BAD - loading all data:
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
    c.Items = db.GetAllItems(ctx) // Could be millions of rows!
    return nil
}

// GOOD - pagination:
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
    c.Items = db.GetItems(ctx, c.Page, 20) // Only 20 items
    return nil
}
```

#### 3. Use Hidden Fields for Small State
```go
// Instead of reloading from DB every time:
type Component struct {
    Count int `form:"count"` // Passed via hidden field
}
```

```html
<input type="hidden" name="count" value="{ fmt.Sprint(data.Count) }" />
```

#### 4. Cache Database Queries
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
    // Check cache first
    if cached, ok := cache.Get(c.UserID); ok {
        c.Data = cached
        return nil
    }

    // Load from database
    data, err := db.LoadData(ctx, c.UserID)
    if err != nil {
        return err
    }

    // Store in cache
    cache.Set(c.UserID, data, 5*time.Minute)
    c.Data = data
    return nil
}
```

---

## Debugging Tips

### Enable Debug Mode
```go
registry := components.NewRegistry()
registry.EnableDebugMode()

// Check browser Network tab for debug headers:
// - X-HxComponent-Name: component name
// - X-HxComponent-FormFields: number of fields
// - X-HxComponent-HasEvent: true/false
```

### Enable Debug Logging
```go
import "log/slog"

// Enable debug level logging
slog.SetLogLoggerLevel(slog.LevelDebug)

// Or use JSON handler for structured logging:
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
slog.SetDefault(slog.New(handler))
```

### Inspect Network Requests
1. Open browser DevTools (F12)
2. Go to Network tab
3. Trigger your component
4. Click on the request
5. Check:
   - Request Headers (HX-*)
   - Form Data
   - Response Headers (HX-*)
   - Response Preview

### Use Playwright for E2E Testing
```go
func TestMyComponent(t *testing.T) {
    // Full browser testing
    page := setupPlaywright(t)
    page.Goto("http://localhost:8080")

    // Interact and inspect
    page.Click("#my-button")
    content := page.Locator("#result").TextContent()

    require.Contains(t, content, "Expected text")
}
```

### Check Server Logs
```go
// Add strategic logging in your components:
func (c *Component) Process(ctx context.Context) error {
    slog.Info("processing component",
        "user_id", c.UserID,
        "action", c.Action)

    // Your logic here

    slog.Info("processing complete",
        "success", true)
    return nil
}
```

---

## Common Error Messages

### "Component not found"
**Cause**: Component name in URL doesn't match registered name
**Fix**: Check that URL path matches the registered name exactly (case-sensitive)

### "Method not allowed"
**Cause**: Using wrong HTTP method (e.g., PUT, DELETE)
**Fix**: Only GET and POST are supported

### "Failed to decode form data"
**Cause**: Form tag doesn't match HTML input name, or invalid data type
**Fix**: Ensure form tags match input names exactly

### "Component does not implement templ.Component"
**Cause**: Missing Render method or wrong signature
**Fix**: Implement `Render(ctx context.Context, w io.Writer) error`

### "event handler method 'OnXxx' not found"
**Cause**: Method name doesn't match event name
**Fix**: Ensure method is named `On` + capitalized event name

### "BeforeEvent/AfterEvent failed"
**Cause**: Error returned from lifecycle hook
**Fix**: Check your hook implementation for bugs

---

## Getting More Help

If you're still stuck:

1. **Check the Examples** - Look at `/examples` directory for working code
2. **Read the Docs** - See [docs/README.md](README.md) for comprehensive guides
3. **Enable Debug Mode** - Use `registry.EnableDebugMode()` and check headers
4. **Check Logs** - Enable debug logging: `slog.SetLogLoggerLevel(slog.LevelDebug)`
5. **GitHub Issues** - Search or create an issue at https://github.com/ocomsoft/HxComponents/issues

---

## Best Practices to Avoid Issues

1. ✅ **Always run `templ generate`** before testing
2. ✅ **Use debug mode** during development
3. ✅ **Enable debug logging** to see what's happening
4. ✅ **Test with browser DevTools** Network tab open
5. ✅ **Use exact naming** for form tags and input names
6. ✅ **Implement Render method** correctly
7. ✅ **Use pointer types** for component registration
8. ✅ **Check context errors** in database queries
9. ✅ **Respect context cancellation** in long-running operations
10. ✅ **Write tests** - unit, integration, and E2E

---

**Last Updated**: 2025-10-23

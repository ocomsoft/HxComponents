# HTMX Generic Component Registry

A type-safe, reusable Go library pattern for building dynamic HTMX components with minimal boilerplate.

## Overview

The HTMX Generic Component Registry eliminates repetitive handler code by providing a clean, type-safe abstraction for:

- **Form parsing** into typed structs using generics
- **HTMX request headers** (HX-Boosted, HX-Request, HX-Current-URL, etc.)
- **HTMX response headers** (HX-Redirect, HX-Trigger, HX-Reswap, etc.)
- **Dynamic component routing** via centralized registry
- **Type-safe rendering** with templ templates

## Features

- ✅ **Type-safe** - Uses Go 1.23+ generics for compile-time verification
- ✅ **Zero boilerplate** - Automatic form parsing and header handling
- ✅ **Interface-based** - Optional HTMX headers via clean interfaces
- ✅ **Framework agnostic** - Works with chi, gorilla/mux, net/http
- ✅ **Composable** - Mix and match header interfaces as needed
- ✅ **Battle-tested** - Minimal dependencies, uses reflection efficiently

## Quick Start

### Installation

```bash
go get github.com/ocomsoft/HxComponents
```

### Basic Example

**1. Define your component data struct:**

```go
package mycomponent

type SearchComponent struct {
    Query string `form:"q"`
    Limit int    `form:"limit"`
}
```

**2. Create a templ component:**

```templ
package mycomponent

templ Search(data SearchComponent) {
    <div>
        <p>Query: { data.Query }</p>
        <p>Limit: { fmt.Sprint(data.Limit) }</p>
    </div>
}
```

**3. Register and serve:**

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.SearchComponent)

    router := chi.NewRouter()
    registry.Mount(router)

    http.ListenAndServe(":8080", router)
}
```

**4. Use in HTML:**

```html
<form hx-post="/component/search" hx-target="#results">
    <input type="text" name="q" />
    <input type="number" name="limit" value="10" />
    <button>Search</button>
</form>
<div id="results"></div>
```

**5. Or use with GET for initial state:**

```html
<!-- Load component with query parameters -->
<div hx-get="/component/search?q=golang&limit=5" hx-trigger="load" hx-target="this"></div>

<!-- Or via a link/button -->
<button hx-get="/component/search?q=htmx&limit=10" hx-target="#results">
    Load Search Results
</button>
```

## HTMX Request Headers

Capture HTMX request headers and HTTP method by implementing optional interfaces:

```go
type SearchComponent struct {
    Query      string `form:"q"`
    IsBoosted  bool   `json:"-"`
    CurrentURL string `json:"-"`
    Method     string `json:"-"` // "GET" or "POST"
}

func (s *SearchComponent) SetHxBoosted(v bool)      { s.IsBoosted = v }
func (s *SearchComponent) SetHxCurrentURL(v string) { s.CurrentURL = v }
func (s *SearchComponent) SetHttpMethod(v string)   { s.Method = v }
```

The `HttpMethod` interface is useful for varying component behavior based on GET vs POST:

```go
func (s *SearchComponent) Process() error {
    if s.Method == "GET" {
        // Load default search results
        s.Query = s.Query // Keep query from URL params
    } else {
        // POST - user submitted form
        // Validate input, log search, etc.
    }
    return nil
}
```

**Available Request Interfaces:**

| Interface | Header/Source | Type |
|-----------|--------|------|
| `HxBoosted` | HX-Boosted | bool |
| `HxRequest` | HX-Request | bool |
| `HxCurrentURL` | HX-Current-URL | string |
| `HxPrompt` | HX-Prompt | string |
| `HxTarget` | HX-Target | string |
| `HxTrigger` | HX-Trigger | string |
| `HxTriggerName` | HX-Trigger-Name | string |
| `HttpMethod` | HTTP Method (GET/POST) | string |

## HTMX Response Headers

Set HTMX response headers by implementing getter interfaces:

```go
type LoginComponent struct {
    Username   string `form:"username"`
    Password   string `form:"password"`
    RedirectTo string `json:"-"`
}

func (f *LoginComponent) GetHxRedirect() string {
    return f.RedirectTo
}
```

**Available Response Interfaces:**

| Interface | Header | Type |
|-----------|--------|------|
| `HxLocationResponse` | HX-Location | string |
| `HxPushUrlResponse` | HX-Push-Url | string |
| `HxRedirectResponse` | HX-Redirect | string |
| `HxRefreshResponse` | HX-Refresh | bool |
| `HxReplaceUrlResponse` | HX-Replace-Url | string |
| `HxReswapResponse` | HX-Reswap | string |
| `HxRetargetResponse` | HX-Retarget | string |
| `HxReselectResponse` | HX-Reselect | string |
| `HxTriggerResponse` | HX-Trigger | string |
| `HxTriggerAfterSettleResponse` | HX-Trigger-After-Settle | string |
| `HxTriggerAfterSwapResponse` | HX-Trigger-After-Swap | string |

## GET vs POST Requests

The registry supports both GET and POST requests for maximum flexibility:

### POST Requests (Standard Pattern)

POST is the standard HTMX pattern for form submissions:

```html
<form hx-post="/component/search" hx-target="#results">
    <input type="text" name="q" value="htmx" />
    <button>Search</button>
</form>
```

Form data is sent in the request body and parsed into the component struct.

### GET Requests (Initial State)

GET requests are useful for loading components with initial state or query parameters:

```html
<!-- Load on page load -->
<div hx-get="/component/search?q=golang&limit=5"
     hx-trigger="load"
     hx-target="this">
</div>

<!-- Load on click -->
<button hx-get="/component/search?q=htmx&limit=10"
        hx-target="#results">
    Load Popular Searches
</button>

<!-- Preload with hx-boost -->
<a href="/component/search?q=go&limit=20"
   hx-boost="true"
   hx-target="#results">
    Go Results
</a>
```

Query parameters are parsed into the component struct, just like POST form data.

**Use Cases for GET:**
- Loading components with default/initial values
- Deep-linking to specific component states
- Shareable URLs with query parameters
- Server-side rendering of initial state
- Progressive enhancement patterns

## Component Processing

Components can implement the `Processor` interface to perform business logic, validation, or data transformation after form decoding but before rendering.

### The Processor Interface

```go
type Processor interface {
    Process() error
}
```

The registry automatically calls `Process()` if your component implements this interface:

```go
type LoginComponent struct {
    Username   string `form:"username"`
    Password   string `form:"password"`
    RedirectTo string `json:"-"`
    Error      string `json:"-"`
}

// Implement Processor interface
func (f *LoginComponent) Process() error {
    if f.Username == "demo" && f.Password == "password" {
        f.RedirectTo = "/dashboard"  // This will trigger HX-Redirect
        return nil
    }
    f.Error = "Invalid credentials"
    return nil
}

// Implement response header interface
func (f *LoginComponent) GetHxRedirect() string {
    return f.RedirectTo
}
```

**Processing Flow:**
1. Form data decoded into struct
2. Request headers applied (HX-Boosted, HX-Request, etc.)
3. **`Process()` called** (if interface implemented)
4. Response headers applied (HX-Redirect, HX-Trigger, etc.)
5. Component rendered

**When to Use:**
- Form validation
- Authentication/authorization
- Database operations
- Setting conditional response headers
- Business logic that affects rendering

**Error Handling:**
- Return `error` only for unexpected system failures
- Store validation errors in struct fields for rendering
- Example: `f.Error = "Invalid input"` instead of `return err`

## Advanced Examples

### Login Component with Redirect

This example is now simplified with the `Processor` interface:

```go
type LoginComponent struct {
    Username   string `form:"username"`
    Password   string `form:"password"`
    RedirectTo string `json:"-"`
    Error      string `json:"-"`
}

// Processor interface - called automatically by registry
func (f *LoginComponent) Process() error {
    if f.Username == "demo" && f.Password == "password" {
        f.RedirectTo = "/dashboard"
        return nil
    }
    f.Error = "Invalid credentials"
    return nil
}

// Response header interface
func (f *LoginComponent) GetHxRedirect() string {
    return f.RedirectTo
}
```

**Register (simplified):**

```go
components.Register(registry, "login", Login)
```

The registry automatically calls `Process()` before rendering!

### Profile Update with Array Support

```go
type ProfileComponent struct {
    Name  string   `form:"name"`
    Email string   `form:"email"`
    Tags  []string `form:"tags"`
}
```

**HTML:**

```html
<form hx-post="/component/profile" hx-target="#result">
    <input name="name" value="John" />
    <input name="email" value="john@example.com" />
    <input name="tags" value="developer" />
    <input name="tags" value="golang" />
    <button>Update</button>
</form>
```

## Running the Example

The `examples/` directory contains a complete demo application:

```bash
cd examples
go run main.go
```

Then open http://localhost:8080 in your browser.

The demo includes:
- **Search Component** - Demonstrates request header capture
- **Login Component** - Demonstrates response headers (redirects)
- **Profile Component** - Demonstrates complex form data with arrays

## Architecture

### Registry Flow

```
1. HTTP POST /component/{name}
2. Registry finds component entry
3. Parse form data into typed struct
4. Apply HTMX request headers (if interfaces implemented)
5. Execute component logic (optional)
6. Apply HTMX response headers (if interfaces implemented)
7. Render templ component
8. Return HTML response
```

### Type Safety

The registry uses Go generics to maintain type safety:

```go
func Register[T any](r *Registry, name string, render func(T) templ.Component)
```

This ensures:
- ✅ Compile-time type checking
- ✅ No runtime type assertion errors
- ✅ IDE autocomplete and refactoring support

## API Reference

### Registry Methods

#### `NewRegistry() *Registry`
Creates a new component registry.

#### `Register[T any](r *Registry, name string, render func(T) templ.Component)`
Registers a component with type-safe rendering function.

**Parameters:**
- `name` - Component name used in URL path
- `render` - Function that takes typed data and returns templ.Component

#### `Mount(router *chi.Mux)`
Mounts the registry handler to chi router at `/component/{component_name}`.

#### `Handler(w http.ResponseWriter, req *http.Request)`
HTTP handler for component rendering. Can be used with any router.

## Best Practices

### 1. Use Descriptive Component Names

```go
// Good
components.Register(registry, "user-search", SearchComponent)
components.Register(registry, "profile-edit", ProfileComponent)

// Avoid
components.Register(registry, "comp1", SearchComponent)
```

### 2. Keep Component Logic Separate

```go
// Good - logic in method
func (f *LoginComponent) ProcessLogin() error { ... }

components.Register(registry, "login", func(data LoginComponent) templ.Component {
    data.ProcessLogin()
    return LoginComponent(data)
})

// Avoid - logic in template
```

### 3. Use JSON Tags to Hide Internal Fields

```go
type MyForm struct {
    UserInput  string `form:"input"`
    RedirectTo string `json:"-"` // Won't be serialized
    IsBoosted  bool   `json:"-"` // Internal state
}
```

### 4. Implement Only Needed Interfaces

```go
// Good - only implement what you need
type SimpleForm struct {
    Query string `form:"q"`
}

// Avoid - implementing unused interfaces
```

## Troubleshooting

### Form Fields Not Parsing

**Problem:** Struct fields remain empty after form submission.

**Solution:** Ensure form tags match HTML input names exactly:

```go
type Form struct {
    Email string `form:"email"` // Must match <input name="email">
}
```

### Headers Not Applied

**Problem:** HTMX headers not being captured or set.

**Solution:** Verify interface implementation:

```go
// Implement pointer receiver
func (f *MyForm) SetHxBoosted(v bool) {
    f.IsBoosted = v
}
```

### Component Not Found

**Problem:** 404 error when posting to component.

**Solution:** Check component name matches registration:

```go
components.Register(registry, "my-component", ...)
// POST to: /component/my-component
```

## Dependencies

- [templ](https://github.com/a-h/templ) - Type-safe Go templating
- [chi](https://github.com/go-chi/chi) - Lightweight router (optional)
- [form](https://github.com/go-playground/form) - Form decoding

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details.

## Learn More

- [HTMX Documentation](https://htmx.org/)
- [templ Documentation](https://templ.guide/)
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)

---

**Built with ❤️ for the Go + HTMX community**

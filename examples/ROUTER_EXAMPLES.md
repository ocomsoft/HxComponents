# Router Integration Examples

This document demonstrates how to use the HxComponents registry with different Go HTTP routers.

## Table of Contents

- [chi Router](#chi-router)
  - [Wildcard Pattern (Recommended)](#wildcard-pattern-recommended)
  - [Specific URLs](#specific-urls)
- [gorilla/mux Router](#gorillamux-router)
- [net/http (Standard Library)](#nethttp-standard-library)
- [Mixing Components with Custom Handlers](#mixing-components-with-custom-handlers)

## chi Router

### Wildcard Pattern (Recommended)

The recommended approach uses wildcard routing with the `Handler` method. The registry automatically extracts the component name from the URL path:

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    // Create registry and register components
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    // Setup chi router
    router := chi.NewRouter()
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)

    // Mount all components with wildcard pattern
    router.Get("/component/*", registry.Handler)
    router.Post("/component/*", registry.Handler)

    // Serve
    http.ListenAndServe(":8080", router)
}
```

**URLs:**
- `/component/search` → renders `search` component
- `/component/login` → renders `login` component
- `/component/profile` → renders `profile` component

### Specific URLs

For explicit control over each component URL, use `HandlerFor()`:

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    router := chi.NewRouter()

    // Mount components at specific URLs
    router.Get("/search", registry.HandlerFor("search"))
    router.Post("/search", registry.HandlerFor("search"))

    router.Get("/auth/login", registry.HandlerFor("login"))
    router.Post("/auth/login", registry.HandlerFor("login"))

    router.Get("/user/profile", registry.HandlerFor("profile"))
    router.Post("/user/profile", registry.HandlerFor("profile"))

    http.ListenAndServe(":8080", router)
}
```

**URLs:**
- `/search` → renders `search` component
- `/auth/login` → renders `login` component
- `/user/profile` → renders `profile` component

## gorilla/mux Router

### Wildcard Pattern

```go
package main

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    router := mux.NewRouter()

    // Mount all components with path prefix
    router.PathPrefix("/component/").HandlerFunc(registry.Handler).Methods("GET", "POST")

    http.ListenAndServe(":8080", router)
}
```

**URLs:**
- `/component/search` → renders `search` component
- `/component/login` → renders `login` component
- `/component/profile` → renders `profile` component

### Specific URLs

```go
package main

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    router := mux.NewRouter()

    // Mount components at specific URLs
    router.HandleFunc("/search", registry.HandlerFor("search")).Methods("GET", "POST")
    router.HandleFunc("/auth/login", registry.HandlerFor("login")).Methods("GET", "POST")
    router.HandleFunc("/user/profile", registry.HandlerFor("profile")).Methods("GET", "POST")

    http.ListenAndServe(":8080", router)
}
```

**URLs:**
- `/search` → renders `search` component
- `/auth/login` → renders `login` component
- `/user/profile` → renders `profile` component

## net/http (Standard Library)

### Wildcard Pattern

```go
package main

import (
    "net/http"

    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    // Mount all components with path prefix
    // The trailing slash makes it match all paths under /component/
    http.HandleFunc("/component/", registry.Handler)

    http.ListenAndServe(":8080", nil)
}
```

**URLs:**
- `/component/search` → renders `search` component
- `/component/login` → renders `login` component
- `/component/profile` → renders `profile` component

**Note:** With `net/http`, the handler will accept all HTTP methods. The registry's `Handler` method internally checks for GET or POST and returns a `405 Method Not Allowed` for other methods.

### Specific URLs

```go
package main

import (
    "net/http"

    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)
    components.Register(registry, "profile", mycomponent.Profile)

    // Mount components at specific URLs
    http.HandleFunc("/search", registry.HandlerFor("search"))
    http.HandleFunc("/auth/login", registry.HandlerFor("login"))
    http.HandleFunc("/user/profile", registry.HandlerFor("profile"))

    http.ListenAndServe(":8080", nil)
}
```

**URLs:**
- `/search` → renders `search` component
- `/auth/login` → renders `login` component
- `/user/profile` → renders `profile` component

## Mixing Components with Custom Handlers

You can easily mix component handlers with your custom route handlers:

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/ocomsoft/HxComponents/components"
    "myapp/mycomponent"
    "myapp/pages"
)

func main() {
    registry := components.NewRegistry()
    components.Register(registry, "search", mycomponent.Search)
    components.Register(registry, "login", mycomponent.Login)

    router := chi.NewRouter()

    // Custom page handlers
    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        pages.HomePage().Render(r.Context(), w)
    })

    router.Get("/about", func(w http.ResponseWriter, r *http.Request) {
        pages.AboutPage().Render(r.Context(), w)
    })

    router.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
        pages.DashboardPage().Render(r.Context(), w)
    })

    // Component handlers with wildcard
    router.Get("/component/*", registry.Handler)
    router.Post("/component/*", registry.Handler)

    // Or specific component URLs
    router.Get("/search", registry.HandlerFor("search"))
    router.Post("/search", registry.HandlerFor("search"))

    http.ListenAndServe(":8080", router)
}
```

This approach allows you to:
- Serve static pages with custom handlers
- Use HTMX components for dynamic interactions
- Organize your routes however you prefer
- Mix different routing patterns as needed

## How URL Extraction Works

The `Handler` method extracts the component name from the last segment of the URL path:

| URL Path | Extracted Component Name |
|----------|-------------------------|
| `/component/search` | `search` |
| `/component/login` | `login` |
| `/api/search` | `search` |
| `/search` | `search` |
| `/auth/login` | `login` |

The component name is always the text after the last `/` in the URL path.

## Best Practices

1. **Use `Handler` with wildcard patterns** - Simplest approach for mounting all components
2. **Use `HandlerFor()` for explicit URLs** - When you need full control over each component's URL
3. **Register both GET and POST** - Components work best when they support both methods (GET for initial state, POST for form submissions)
4. **Custom URL paths** - Choose URL paths that match your application's routing conventions
5. **Method restrictions** - Use your router's method filtering (`.Methods()`, `.Get()`, `.Post()`) to restrict HTTP methods if needed
6. **Consistent patterns** - Pick one pattern (wildcard or explicit) and use it consistently throughout your app

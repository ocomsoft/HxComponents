// Package components provides a type-safe, HTMX-based component registry
// for building dynamic web applications with Go and templ.
//
// The registry eliminates boilerplate by providing automatic form parsing,
// HTMX header handling, event-driven processing, and type-safe rendering.
//
// Example usage:
//
//	registry := components.NewRegistry()
//	components.Register[*MyComponent](registry, "mycomponent")
//	http.HandleFunc("/component/", registry.Handler)
package components

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
)

var defaultDecoder = form.NewDecoder()

// componentEntry stores the type information for a registered component.
type componentEntry struct {
	structType reflect.Type
}

// ErrorHandler is a function that renders error responses
type ErrorHandler func(w http.ResponseWriter, req *http.Request, title string, message string, code int)

// Registry manages component registration and handles HTTP requests for component rendering.
// It is safe for concurrent use by multiple goroutines.
type Registry struct {
	mu           sync.RWMutex
	components   map[string]componentEntry
	errorHandler ErrorHandler
	debugMode    bool
}

// NewRegistry creates a new component registry with the default error handler.
func NewRegistry() *Registry {
	return &Registry{
		components:   make(map[string]componentEntry),
		errorHandler: defaultErrorHandler,
	}
}

// SetErrorHandler sets a custom error handler for the registry.
// The error handler is responsible for rendering error responses.
func (r *Registry) SetErrorHandler(handler ErrorHandler) {
	r.errorHandler = handler
}

// EnableDebugMode enables debug mode for the registry.
// When enabled, additional debugging headers are added to responses:
//   - X-HxComponent-Name: The component name
//   - X-HxComponent-FormFields: Number of form fields received
//   - X-HxComponent-HasEvent: Whether an event was processed
//
// This is useful during development to understand component rendering.
// WARNING: Do not enable in production as it exposes internal details.
func (r *Registry) EnableDebugMode() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debugMode = true
	slog.Info("debug mode enabled for component registry")
}

// DisableDebugMode disables debug mode for the registry.
func (r *Registry) DisableDebugMode() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debugMode = false
	slog.Info("debug mode disabled for component registry")
}

// IsDebugMode returns whether debug mode is currently enabled.
func (r *Registry) IsDebugMode() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.debugMode
}

// defaultErrorHandler is the default error handler that renders the ErrorComponent
func defaultErrorHandler(w http.ResponseWriter, req *http.Request, title string, message string, code int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	if err := ErrorComponent(title, message, code).Render(req.Context(), w); err != nil {
		slog.Error("failed to render error component",
			"error", err,
			"title", title,
			"message", message,
			"code", code,
			"path", req.URL.Path)

		// Fallback to plain HTML with proper escaping
		fmt.Fprintf(w,
			`<div style="border:2px solid red;padding:1em;margin:1em">
				<h3>%s</h3>
				<p>%s</p>
				<small>Error Code: %d</small>
				<hr>
				<small>Additionally, the error template failed to render.</small>
			</div>`,
			html.EscapeString(title),
			html.EscapeString(message),
			code)
	}
}

// Register registers a component type that implements templ.Component.
// The name parameter is used in the URL path: /component/{name}
// The component type T must implement templ.Component's Render method.
//
// If the component type implements the Processor interface, its Process() method
// will be called after form decoding and before rendering, allowing you to perform
// validation, business logic, or set response headers.
//
// Example:
//
//	components.Register[*login.LoginComponent](registry, "login")
//
// Why is this a package-level function instead of a method on Registry?
//
// Go methods cannot have type parameters (as of Go 1.23). This is a fundamental
// limitation of Go generics. The Go team made this design decision to avoid
// complications with method dispatch, interface implementation, and type system complexity.
//
// We considered alternatives:
//   - Builder pattern: registry.Component("name").As[*Type]() - Still needs generic method
//   - Reflection-only: registry.Register("name", &Component{}) - Loses type safety
//   - Package function: components.Register[*Type](registry, "name") - Works! ✅
//
// The package-level generic function is the idiomatic Go approach for this pattern.
// See: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
func Register[T templ.Component](r *Registry, name string) {
	// Validate component name
	if name == "" {
		panic("component name cannot be empty")
	}

	// Get the type - T is already a pointer type
	var zero T
	structType := reflect.TypeOf(zero)

	// Validate that T is a pointer type
	if structType == nil {
		panic(fmt.Sprintf("component type cannot be nil (component name: %s)", name))
	}

	if structType.Kind() != reflect.Ptr {
		typeName := structType.Name()
		if typeName == "" {
			typeName = structType.String()
		}
		panic(fmt.Sprintf(
			"component type must be a pointer type, got %T\n"+
				"Hint: Use Register[*%s](registry, %q) instead of Register[%s](...)",
			zero, typeName, name, structType.String()))
	}

	// Validate that the pointer points to a struct
	if structType.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf(
			"component must point to a struct, got pointer to %s (component name: %s)\n"+
				"Hint: Components must be struct types that implement templ.Component",
			structType.Elem().Kind(), name))
	}

	// Validate that the component implements templ.Component
	// This is enforced at compile time by the generic constraint,
	// but we verify it here for runtime safety
	if _, ok := interface{}(zero).(templ.Component); !ok {
		structName := structType.Elem().Name()
		panic(fmt.Sprintf(
			"component type %T does not implement templ.Component (component name: %s)\n"+
				"Hint: Add a Render(ctx context.Context, w io.Writer) error method to %s",
			zero, name, structName))
	}

	// Thread-safe registration
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate registration
	if _, exists := r.components[name]; exists {
		panic(fmt.Sprintf("component '%s' already registered", name))
	}

	structType = structType.Elem()
	r.components[name] = componentEntry{
		structType: structType,
	}
}

// HandlerFor returns an http.HandlerFunc for rendering a specific component.
// This allows you to mount components at any URL path using any router.
//
// Component Request Lifecycle:
//
//	┌──────────────────────┐
//	│   HTTP Request       │
//	│  (GET/POST)          │
//	└──────────┬───────────┘
//	           │
//	           ▼
//	┌──────────────────────┐
//	│   Parse Form Data    │
//	│  (query/body params) │
//	└──────────┬───────────┘
//	           │
//	           ▼
//	┌──────────────────────┐
//	│  Decode into Struct  │
//	│  (form tags)         │
//	└──────────┬───────────┘
//	           │
//	           ▼
//	┌──────────────────────┐
//	│  Apply Request       │
//	│  Headers (HX-*)      │
//	└──────────┬───────────┘
//	           │
//	           ▼
//	     ┌─────────────┐
//	     │ Has Event?  │
//	     │ (hxc-event) │
//	     └──┬──────┬───┘
//	        │ YES  │ NO
//	        │      └─────────────────┐
//	        ▼                        │
//	┌──────────────────────┐         │
//	│  BeforeEvent(ctx)    │         │
//	│  (optional hook)     │         │
//	└──────────┬───────────┘         │
//	           │                     │
//	           ▼                     │
//	┌──────────────────────┐         │
//	│  On{EventName}()     │         │
//	│  (event handler)     │         │
//	└──────────┬───────────┘         │
//	           │                     │
//	           ▼                     │
//	┌──────────────────────┐         │
//	│  AfterEvent(ctx)     │         │
//	│  (optional hook)     │         │
//	└──────────┬───────────┘         │
//	           │                     │
//	           └──────┬──────────────┘
//	                  │
//	                  ▼
//	         ┌──────────────────────┐
//	         │  Process(ctx)        │
//	         │  (optional)          │
//	         └──────────┬───────────┘
//	                    │
//	                    ▼
//	         ┌──────────────────────┐
//	         │  Apply Response      │
//	         │  Headers (HX-*)      │
//	         └──────────┬───────────┘
//	                    │
//	                    ▼
//	         ┌──────────────────────┐
//	         │  Render(ctx, w)      │
//	         │  (templ.Component)   │
//	         └──────────┬───────────┘
//	                    │
//	                    ▼
//	         ┌──────────────────────┐
//	         │   HTTP Response      │
//	         │   (HTML)             │
//	         └──────────────────────┘
//
// Example with net/http:
//
//	http.HandleFunc("/search", registry.HandlerFor("search"))
//
// Example with chi:
//
//	router.Get("/search", registry.HandlerFor("search"))
//	router.Post("/search", registry.HandlerFor("search"))
//
// Example with gorilla/mux:
//
//	router.HandleFunc("/search", registry.HandlerFor("search"))
func (r *Registry) HandlerFor(componentName string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Panic recovery
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic in component handler",
					"component", componentName,
					"error", err,
					"stack", string(debug.Stack()))
				r.renderError(w, req, "Internal Server Error",
					"Component encountered an unexpected error",
					http.StatusInternalServerError)
			}
		}()

		if req.Method != http.MethodPost && req.Method != http.MethodGet {
			slog.Warn("method not allowed",
				"method", req.Method,
				"path", req.URL.Path,
				"component", componentName)
			r.renderError(w, req, "Method Not Allowed", fmt.Sprintf("Method %s is not allowed", req.Method), http.StatusMethodNotAllowed)
			return
		}

		// Thread-safe component lookup
		r.mu.RLock()
		entry, exists := r.components[componentName]
		r.mu.RUnlock()

		if !exists {
			slog.Warn("component not found",
				"component", componentName,
				"path", req.URL.Path)
			r.renderError(w, req, "Component Not Found", fmt.Sprintf("Component '%s' not found", componentName), http.StatusNotFound)
			return
		}

		slog.Debug("rendering component",
			"component", componentName,
			"method", req.Method,
			"remote_addr", req.RemoteAddr,
			"user_agent", req.UserAgent(),
			"content_type", req.Header.Get("Content-Type"))

		if err := req.ParseForm(); err != nil {
			slog.Error("form parse error",
				"component", componentName,
				"error", err)
			r.renderError(w, req, "Bad Request", fmt.Sprintf("Failed to parse form data: %v", err), http.StatusBadRequest)
			return
		}

		// Create instance and decode form
		instance := reflect.New(entry.structType)

		// For POST, use PostForm; for GET, use Form (which includes query params)
		var formData map[string][]string
		if req.Method == http.MethodPost {
			formData = req.PostForm
		} else {
			formData = req.Form
		}

		// Use component's custom decoder if provided, otherwise use default
		decoder := defaultDecoder
		if customDecoder, ok := instance.Interface().(FormDecoder); ok {
			decoder = customDecoder.GetFormDecoder()
			slog.Debug("using custom form decoder",
				"component", componentName)
		}

		if err := decoder.Decode(instance.Interface(), formData); err != nil {
			slog.Error("form decode error",
				"component", componentName,
				"error", err)
			r.renderError(w, req, "Decode Error", fmt.Sprintf("Failed to decode form data: %v", err), http.StatusBadRequest)
			return
		}

		// Apply request headers
		applyHxHeaders(instance.Interface(), req)

		// Initialize component if it implements Initializer interface
		if initializer, ok := instance.Interface().(Initializer); ok {
			if err := initializer.Init(req.Context()); err != nil {
				slog.Error("component init error",
					"component", componentName,
					"error", err)
				r.renderError(w, req, "Initialization Error", fmt.Sprintf("Component initialization failed: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Validate if component implements Validator interface
		if validator, ok := instance.Interface().(Validator); ok {
			if errs := validator.Validate(req.Context()); len(errs) > 0 {
				slog.Debug("validation errors",
					"component", componentName,
					"errors", errs)
				// Validation errors don't stop processing - they're stored in the component
				// and can be rendered in the template. Components can choose to handle
				// validation errors differently by checking in their Process() method.
			}
		}

		// Handle event-driven processing if hxc-event parameter is present
		hasEvent := false
		if eventNames, ok := formData["hxc-event"]; ok && len(eventNames) > 0 {
			hasEvent = true
			eventName := eventNames[0]
			slog.Debug("processing event",
				"component", componentName,
				"event", eventName)
			if err := r.handleEvent(req.Context(), instance.Interface(), eventName, componentName); err != nil {
				slog.Error("event handler error",
					"component", componentName,
					"event", eventName,
					"error", err,
					"remote_addr", req.RemoteAddr)
				r.renderError(w, req, "Event Error", fmt.Sprintf("Event '%s' failed: %v", eventName, err), http.StatusInternalServerError)
				return
			}
		}

		// Call Process if the component implements the Processor interface
		if processor, ok := instance.Interface().(Processor); ok {
			if err := processor.Process(req.Context()); err != nil {
				slog.Error("component process error",
					"component", componentName,
					"error", err)
				r.renderError(w, req, "Processing Error", fmt.Sprintf("Component processing failed: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Apply response headers (after processing, so we capture any changes made during Process)
		applyHxResponseHeaders(w, instance.Interface())

		// Add debug headers if debug mode is enabled
		if r.IsDebugMode() {
			w.Header().Set("X-HxComponent-Name", componentName)
			w.Header().Set("X-HxComponent-FormFields", fmt.Sprintf("%d", len(req.Form)))
			if hasEvent {
				w.Header().Set("X-HxComponent-HasEvent", "true")
			} else {
				w.Header().Set("X-HxComponent-HasEvent", "false")
			}
		}

		// Render component - the instance itself implements templ.Component
		w.Header().Set("Content-Type", "text/html")
		component, ok := instance.Interface().(templ.Component)
		if !ok {
			slog.Error("component does not implement templ.Component",
				"component", componentName)
			r.renderError(w, req, "Configuration Error", "Component does not implement templ.Component", http.StatusInternalServerError)
			return
		}

		if err := component.Render(req.Context(), w); err != nil {
			slog.Error("component render error",
				"component", componentName,
				"error", err)
			r.renderError(w, req, "Render Error", fmt.Sprintf("Component rendering failed: %v", err), http.StatusInternalServerError)
			return
		}

		slog.Debug("component rendered successfully",
			"component", componentName,
			"has_event", hasEvent,
			"form_fields", len(req.Form))
	}
}

// handleEvent processes event-driven method calls on a component.
// It implements the lifecycle: BeforeEvent → On{EventName} → AfterEvent
// Returns an error if any step fails, stopping further processing.
func (r *Registry) handleEvent(ctx context.Context, instance interface{}, eventName, componentName string) error {
	// Call BeforeEvent hook if component implements it
	if beforeHandler, ok := instance.(BeforeEventHandler); ok {
		slog.Debug("calling BeforeEvent hook",
			"component", componentName,
			"event", eventName)
		if err := beforeHandler.BeforeEvent(ctx, eventName); err != nil {
			return fmt.Errorf("BeforeEvent failed: %w", err)
		}
	}

	// Find and call the event handler method: On{EventName}
	// Convert event name to method name (e.g., "increment" -> "OnIncrement")
	methodName := "On" + capitalize(eventName)

	value := reflect.ValueOf(instance)
	method := value.MethodByName(methodName)

	if !method.IsValid() {
		return &ErrEventNotFound{
			ComponentName: componentName,
			EventName:     eventName,
		}
	}

	// Validate event handler signature: On{Event}(ctx context.Context) error
	methodType := method.Type()
	if methodType.NumIn() != 1 {
		return fmt.Errorf("event handler '%s' must have signature On%s(ctx context.Context) error", methodName, capitalize(eventName))
	}

	// Check that first parameter is context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(0).Implements(ctxType) {
		return fmt.Errorf("event handler '%s' first parameter must be context.Context", methodName)
	}

	// Call the event handler method with context
	slog.Debug("calling event handler",
		"component", componentName,
		"event", eventName,
		"method", methodName)

	results := method.Call([]reflect.Value{reflect.ValueOf(ctx)})

	// Check if method returns an error
	if len(results) > 0 {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			return fmt.Errorf("event handler failed: %w", err)
		}
	}

	// Call AfterEvent hook if component implements it
	if afterHandler, ok := instance.(AfterEventHandler); ok {
		slog.Debug("calling AfterEvent hook",
			"component", componentName,
			"event", eventName)
		if err := afterHandler.AfterEvent(ctx, eventName); err != nil {
			return fmt.Errorf("AfterEvent failed: %w", err)
		}
	}

	return nil
}

// capitalize converts the first character of a string to uppercase.
// Used to convert event names to method names (e.g., "increment" -> "OnIncrement").
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Handler extracts the component name from the URL path and renders the component.
// The component name is extracted from the last segment of the URL path after the last slash.
// This allows for wildcard routing patterns.
//
// Example with chi:
//
//	router.Get("/component/*", registry.Handler)
//	router.Post("/component/*", registry.Handler)
//
// Example with gorilla/mux:
//
//	router.PathPrefix("/component/").HandlerFunc(registry.Handler).Methods("GET", "POST")
//
// Example with net/http:
//
//	http.HandleFunc("/component/", registry.Handler)
//
// For URL "/component/search", the component name will be "search".
// For URL "/api/components/login", the component name will be "login".
func (r *Registry) Handler(w http.ResponseWriter, req *http.Request) {
	// Extract component name from URL path (last segment after last slash)
	path := req.URL.Path
	lastSlash := len(path)
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastSlash = i
			break
		}
	}
	componentName := path[lastSlash+1:]

	// Remove trailing slash if present
	if componentName == "" && lastSlash > 0 {
		// Path ends with slash, try again
		for i := lastSlash - 1; i >= 0; i-- {
			if path[i] == '/' {
				componentName = path[i+1 : lastSlash]
				break
			}
		}
	}

	if componentName == "" {
		slog.Warn("empty component name in URL path",
			"path", req.URL.Path)
		r.renderError(w, req, "Bad Request", "Component name cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate component name (alphanumeric, dash, underscore only)
	if !isValidComponentName(componentName) {
		err := &ErrInvalidComponentName{
			ComponentName: componentName,
			Reason:        "component names must contain only alphanumeric characters, dashes, and underscores, and be less than 100 characters",
		}
		slog.Warn("invalid component name",
			"component", componentName,
			"path", req.URL.Path,
			"error", err)
		r.renderError(w, req, "Bad Request", err.Error(), http.StatusBadRequest)
		return
	}

	// Use HandlerFor to handle the actual request
	r.HandlerFor(componentName)(w, req)
}

// renderError renders error responses using the configured error handler
func (r *Registry) renderError(w http.ResponseWriter, req *http.Request, title string, message string, code int) {
	r.errorHandler(w, req, title, message, code)
}

// ListComponents returns the names of all registered components in alphabetical order.
func (r *Registry) ListComponents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsRegistered checks if a component name is registered.
func (r *Registry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.components[name]
	return exists
}

// ComponentInfo contains metadata about a registered component.
type ComponentInfo struct {
	Name       string
	StructType string
}

// GetComponentInfo returns metadata about a registered component.
func (r *Registry) GetComponentInfo(name string) (ComponentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, exists := r.components[name]
	if !exists {
		return ComponentInfo{}, &ErrComponentNotFound{ComponentName: name}
	}

	return ComponentInfo{
		Name:       name,
		StructType: meta.structType.String(),
	}, nil
}

// isValidComponentName validates that a component name contains only
// alphanumeric characters, dashes, and underscores, and is not too long.
func isValidComponentName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_') {
			return false
		}
	}
	return true
}

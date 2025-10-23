package components

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
)

var decoder = form.NewDecoder()

// componentEntry stores the type information for a registered component.
type componentEntry struct {
	structType reflect.Type
}

// ErrorHandler is a function that renders error responses
type ErrorHandler func(w http.ResponseWriter, req *http.Request, title string, message string, code int)

// Registry manages component registration and handles HTTP requests for component rendering.
type Registry struct {
	components   map[string]componentEntry
	errorHandler ErrorHandler
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

// defaultErrorHandler is the default error handler that renders the ErrorComponent
func defaultErrorHandler(w http.ResponseWriter, req *http.Request, title string, message string, code int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	if err := ErrorComponent(title, message, code).Render(req.Context(), w); err != nil {
		slog.Error("failed to render error component", "error", err)
		// Fallback to plain text error
		fmt.Fprintf(w, "<div>%s: %s (Code: %d)</div>", title, message, code)
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
	// Get the type - T is already a pointer type
	var zero T
	structType := reflect.TypeOf(zero)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	r.components[name] = componentEntry{
		structType: structType,
	}
}

// HandlerFor returns an http.HandlerFunc for rendering a specific component.
// This allows you to mount components at any URL path using any router.
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
		if req.Method != http.MethodPost && req.Method != http.MethodGet {
			slog.Warn("method not allowed",
				"method", req.Method,
				"path", req.URL.Path,
				"component", componentName)
			r.renderError(w, req, "Method Not Allowed", fmt.Sprintf("Method %s is not allowed", req.Method), http.StatusMethodNotAllowed)
			return
		}

		entry, exists := r.components[componentName]
		if !exists {
			slog.Warn("component not found",
				"component", componentName,
				"path", req.URL.Path)
			r.renderError(w, req, "Component Not Found", fmt.Sprintf("Component '%s' not found", componentName), http.StatusNotFound)
			return
		}

		slog.Debug("rendering component",
			"component", componentName,
			"method", req.Method)

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

		if err := decoder.Decode(instance.Interface(), formData); err != nil {
			slog.Error("form decode error",
				"component", componentName,
				"error", err)
			r.renderError(w, req, "Decode Error", fmt.Sprintf("Failed to decode form data: %v", err), http.StatusBadRequest)
			return
		}

		// Apply request headers
		applyHxHeaders(instance.Interface(), req)

		// Handle event-driven processing if hxc-event parameter is present
		if eventNames, ok := formData["hxc-event"]; ok && len(eventNames) > 0 {
			eventName := eventNames[0]
			if err := r.handleEvent(instance.Interface(), eventName, componentName); err != nil {
				slog.Error("event handler error",
					"component", componentName,
					"event", eventName,
					"error", err)
				r.renderError(w, req, "Event Error", fmt.Sprintf("Event '%s' failed: %v", eventName, err), http.StatusInternalServerError)
				return
			}
		}

		// Call Process if the component implements the Processor interface
		if processor, ok := instance.Interface().(Processor); ok {
			if err := processor.Process(); err != nil {
				slog.Error("component process error",
					"component", componentName,
					"error", err)
				r.renderError(w, req, "Processing Error", fmt.Sprintf("Component processing failed: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Apply response headers (after processing, so we capture any changes made during Process)
		applyHxResponseHeaders(w, instance.Interface())

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
			"component", componentName)
	}
}

// handleEvent processes event-driven method calls on a component.
// It implements the lifecycle: BeforeEvent → On{EventName} → AfterEvent
// Returns an error if any step fails, stopping further processing.
func (r *Registry) handleEvent(instance interface{}, eventName, componentName string) error {
	// Call BeforeEvent hook if component implements it
	if beforeHandler, ok := instance.(BeforeEventHandler); ok {
		slog.Debug("calling BeforeEvent hook",
			"component", componentName,
			"event", eventName)
		if err := beforeHandler.BeforeEvent(eventName); err != nil {
			return fmt.Errorf("BeforeEvent failed: %w", err)
		}
	}

	// Find and call the event handler method: On{EventName}
	// Convert event name to method name (e.g., "increment" -> "OnIncrement")
	methodName := "On" + capitalize(eventName)

	value := reflect.ValueOf(instance)
	method := value.MethodByName(methodName)

	if !method.IsValid() {
		return fmt.Errorf("event handler method '%s' not found", methodName)
	}

	// Call the event handler method
	slog.Debug("calling event handler",
		"component", componentName,
		"event", eventName,
		"method", methodName)

	results := method.Call(nil)

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
		if err := afterHandler.AfterEvent(eventName); err != nil {
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
	runes := []rune(s)
	runes[0] = rune(s[0] - 32) // Convert first char to uppercase
	// Only capitalize if it was lowercase
	if runes[0] >= 'A' && runes[0] <= 'Z' {
		return string(runes)
	}
	// Return original if first char wasn't lowercase
	return s
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

	// Use HandlerFor to handle the actual request
	r.HandlerFor(componentName)(w, req)
}

// renderError renders error responses using the configured error handler
func (r *Registry) renderError(w http.ResponseWriter, req *http.Request, title string, message string, code int) {
	r.errorHandler(w, req, title, message, code)
}

package components

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
)

var decoder = form.NewDecoder()

// componentEntry stores the type information and render function for a registered component.
type componentEntry struct {
	structType reflect.Type
	render     func(interface{}) templ.Component
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

// Register registers a component type with its render function.
// The name parameter is used in the URL path: /component/{name}
// The render function takes a typed instance and returns a templ.Component.
//
// If the component type implements the Processor interface, its Process() method
// will be called after form decoding and before rendering, allowing you to perform
// validation, business logic, or set response headers.
//
// Example:
//
//	components.Register(registry, "search", Search)
func Register[T any](r *Registry, name string, render func(T) templ.Component) {
	r.components[name] = componentEntry{
		structType: reflect.TypeOf((*T)(nil)).Elem(),
		render: func(v interface{}) templ.Component {
			// Convert pointer to value for render function
			ptr := v.(*T)
			return render(*ptr)
		},
	}
}

// Handler handles GET and POST requests for component rendering.
// For POST requests:
//   - The component name is expected as a URL parameter "component_name"
//   - Form data is parsed and bound to the registered component type
//   - HTMX request headers are applied if the component implements the appropriate interfaces
//   - HTMX response headers are set if the component implements the appropriate interfaces
//
// For GET requests:
//   - The component name is expected as a URL parameter "component_name"
//   - Query parameters are parsed and bound to the registered component type
//   - HTMX request headers are applied if the component implements the appropriate interfaces
//   - HTMX response headers are set if the component implements the appropriate interfaces
//   - Useful for rendering components with default/initial state
func (r *Registry) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost && req.Method != http.MethodGet {
		slog.Warn("method not allowed",
			"method", req.Method,
			"path", req.URL.Path)
		r.renderError(w, req, "Method Not Allowed", fmt.Sprintf("Method %s is not allowed", req.Method), http.StatusMethodNotAllowed)
		return
	}

	componentName := chi.URLParam(req, "component_name")
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

	// Render component
	w.Header().Set("Content-Type", "text/html")
	component := entry.render(instance.Interface())
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

// renderError renders error responses using the configured error handler
func (r *Registry) renderError(w http.ResponseWriter, req *http.Request, title string, message string, code int) {
	r.errorHandler(w, req, title, message, code)
}

// Mount registers the handler with the chi router at GET and POST /component/{component_name}.
// GET requests are useful for rendering components with initial/default state or query parameters.
// POST requests are the standard HTMX pattern for form submissions.
func (r *Registry) Mount(router *chi.Mux) {
	router.Get("/component/{component_name}", r.Handler)
	router.Post("/component/{component_name}", r.Handler)
}

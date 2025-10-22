package components

import (
	"fmt"
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

// Registry manages component registration and handles HTTP requests for component rendering.
type Registry struct {
	components map[string]componentEntry
}

// NewRegistry creates a new component registry.
func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]componentEntry),
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
//	components.Register(registry, "search", SearchComponent)
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

// Register is a convenience method that calls the package-level Register function.
func (r *Registry) Register(name string, render interface{}) {
	// This method is intentionally left generic to allow type inference
	// Users should use the package-level Register function for type safety
	panic("Use components.Register[T](registry, name, renderFunc) instead of registry.Register()")
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
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	componentName := chi.URLParam(req, "component_name")
	entry, exists := r.components[componentName]
	if !exists {
		http.Error(w, "component not found", http.StatusNotFound)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("decode error: %v", err), http.StatusBadRequest)
		return
	}

	// Apply request headers
	applyHxHeaders(instance.Interface(), req)

	// Call Process if the component implements the Processor interface
	if processor, ok := instance.Interface().(Processor); ok {
		if err := processor.Process(); err != nil {
			http.Error(w, fmt.Sprintf("process error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Apply response headers (after processing, so we capture any changes made during Process)
	applyHxResponseHeaders(w, instance.Interface())

	// Render component
	w.Header().Set("Content-Type", "text/html")
	component := entry.render(instance.Interface())
	if err := component.Render(req.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("render error: %v", err), http.StatusInternalServerError)
		return
	}
}

// Mount registers the handler with the chi router at GET and POST /component/{component_name}.
// GET requests are useful for rendering components with initial/default state or query parameters.
// POST requests are the standard HTMX pattern for form submissions.
func (r *Registry) Mount(router *chi.Mux) {
	router.Get("/component/{component_name}", r.Handler)
	router.Post("/component/{component_name}", r.Handler)
}

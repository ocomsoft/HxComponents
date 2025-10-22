package components

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/a-h/templ"
	"github.com/go-playground/form/v4"
	"github.com/go-chi/chi/v5"
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
// Example:
//
//	registry.Register("search", SearchComponent)
func Register[T any](r *Registry, name string, render func(T) templ.Component) {
	r.components[name] = componentEntry{
		structType: reflect.TypeOf((*T)(nil)).Elem(),
		render: func(v interface{}) templ.Component {
			return render(v.(T))
		},
	}
}

// Register is a convenience method that calls the package-level Register function.
func (r *Registry) Register(name string, render interface{}) {
	// This method is intentionally left generic to allow type inference
	// Users should use the package-level Register function for type safety
	panic("Use components.Register[T](registry, name, renderFunc) instead of registry.Register()")
}

// Handler handles POST requests for component rendering.
// It expects the component name as a URL parameter "component_name".
// The form data is parsed and bound to the registered component type.
// HTMX request headers are applied if the component implements the appropriate interfaces.
// HTMX response headers are set if the component implements the appropriate interfaces.
func (r *Registry) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
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
	if err := decoder.Decode(instance.Interface(), req.PostForm); err != nil {
		http.Error(w, fmt.Sprintf("decode error: %v", err), http.StatusBadRequest)
		return
	}

	// Apply request headers
	applyHxHeaders(instance.Interface(), req)

	// Apply response headers BEFORE rendering
	applyHxResponseHeaders(w, instance.Interface())

	// Render component
	w.Header().Set("Content-Type", "text/html")
	component := entry.render(instance.Elem().Interface())
	component.Render(req.Context(), w)
}

// Mount registers the handler with the chi router at POST /component/{component_name}.
func (r *Registry) Mount(router *chi.Mux) {
	router.Post("/component/{component_name}", r.Handler)
}

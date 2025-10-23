package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ocomsoft/HxComponents/components"
	"github.com/ocomsoft/HxComponents/examples/login"
	"github.com/ocomsoft/HxComponents/examples/pages"
	"github.com/ocomsoft/HxComponents/examples/profile"
	"github.com/ocomsoft/HxComponents/examples/search"
)

func main() {
	// Create the component registry
	registry := components.NewRegistry()

	// Register components
	// The registry will automatically call Process() if the component implements the Processor interface
	components.Register(registry, "search", search.Search)
	components.Register(registry, "login", login.Login)
	components.Register(registry, "profile", profile.Profile)

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount component handlers with wildcard pattern
	router.Get("/component/*", registry.Handler)
	router.Post("/component/*", registry.Handler)

	// Serve pages using templ
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pages.IndexPage().Render(r.Context(), w)
	})
	router.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		pages.DashboardPage().Render(r.Context(), w)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

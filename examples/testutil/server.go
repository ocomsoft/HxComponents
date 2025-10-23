package testutil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ocomsoft/HxComponents/components"
	"github.com/ocomsoft/HxComponents/examples/counter"
	"github.com/ocomsoft/HxComponents/examples/login"
	"github.com/ocomsoft/HxComponents/examples/pages"
	"github.com/ocomsoft/HxComponents/examples/profile"
	"github.com/ocomsoft/HxComponents/examples/search"
	"github.com/ocomsoft/HxComponents/examples/todolist"
	"github.com/stretchr/testify/require"
)

// TestServer wraps an HTTP server for testing.
type TestServer struct {
	Server   *http.Server
	URL      string
	Registry *components.Registry
	t        *testing.T
}

// NewTestServer creates and starts a new test server.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create the component registry
	registry := components.NewRegistry()

	// Register components
	components.Register[*search.SearchComponent](registry, "search")
	components.Register[*login.LoginComponent](registry, "login")
	components.Register[*profile.ProfileComponent](registry, "profile")
	components.Register[*counter.CounterComponent](registry, "counter")
	components.Register[*todolist.TodoListComponent](registry, "todolist")

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount component handlers with wildcard pattern
	router.Get("/component/*", registry.Handler)
	router.Post("/component/*", registry.Handler)

	// Serve pages using templ
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := pages.IndexPage().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	router.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if err := pages.DashboardPage().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "failed to find available port")

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	server := &http.Server{
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	ts := &TestServer{
		Server:   server,
		URL:      url,
		Registry: registry,
		t:        t,
	}

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	require.Eventually(t, func() bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "server did not start in time")

	t.Logf("Test server started at %s", url)

	return ts
}

// Close shuts down the test server.
func (ts *TestServer) Close() {
	ts.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ts.Server.Shutdown(ctx); err != nil {
		ts.t.Logf("Server shutdown error: %v", err)
	}
}

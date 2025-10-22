package components

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

// Test struct that implements Processor and HxRedirectResponse
type TestLoginForm struct {
	Username   string `form:"username"`
	Password   string `form:"password"`
	RedirectTo string
	Error      string
}

func (f *TestLoginForm) Process() error {
	if f.Username == "demo" && f.Password == "password" {
		f.RedirectTo = "/dashboard"
		return nil
	}
	f.Error = "Invalid credentials"
	return nil
}

func (f *TestLoginForm) GetHxRedirect() string {
	return f.RedirectTo
}

// Test templ component
func renderTestLoginComponent(data TestLoginForm) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if data.Error != "" {
			_, err := w.Write([]byte("<div class=\"error\">" + data.Error + "</div>"))
			return err
		}
		if data.RedirectTo != "" {
			_, err := w.Write([]byte("<div class=\"success\">Login successful!</div>"))
			return err
		}
		return nil
	})
}

func TestProcessorAndRedirect(t *testing.T) {
	// Create registry
	registry := NewRegistry()

	// Register login component
	Register(registry, "login", renderTestLoginComponent)

	// Create router
	router := chi.NewRouter()
	registry.Mount(router)

	// Test successful login with redirect
	t.Run("successful login sets HX-Redirect header", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "demo")
		formData.Set("password", "password")

		req := httptest.NewRequest(http.MethodPost, "/component/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check status code
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check HX-Redirect header
		redirectHeader := w.Header().Get("HX-Redirect")
		if redirectHeader != "/dashboard" {
			t.Errorf("expected HX-Redirect header to be '/dashboard', got '%s'", redirectHeader)
		}

		// Check body contains success message
		body := w.Body.String()
		if !strings.Contains(body, "Login successful!") {
			t.Errorf("expected body to contain success message, got: %s", body)
		}
	})

	// Test failed login
	t.Run("failed login does not set HX-Redirect header", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "wrong")
		formData.Set("password", "wrong")

		req := httptest.NewRequest(http.MethodPost, "/component/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check status code
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check HX-Redirect header is NOT set
		redirectHeader := w.Header().Get("HX-Redirect")
		if redirectHeader != "" {
			t.Errorf("expected no HX-Redirect header, got '%s'", redirectHeader)
		}

		// Check body contains error message
		body := w.Body.String()
		if !strings.Contains(body, "Invalid credentials") {
			t.Errorf("expected body to contain error message, got: %s", body)
		}
	})
}

// Test form for HttpMethod interface
type TestMethodForm struct {
	Query  string `form:"q"`
	Method string
}

// Implement HttpMethod interface
func (f *TestMethodForm) SetHttpMethod(method string) {
	f.Method = method
}

func testMethodComponent(data TestMethodForm) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Method: " + data.Method + "</div>"))
		return err
	})
}

func TestHttpMethodInterface(t *testing.T) {
	registry := NewRegistry()
	Register(registry, "test", testMethodComponent)

	router := chi.NewRouter()
	registry.Mount(router)

	t.Run("POST request sets method", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("q", "test")

		req := httptest.NewRequest(http.MethodPost, "/component/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		body := w.Body.String()
		if !strings.Contains(body, "Method: POST") {
			t.Errorf("expected body to contain 'Method: POST', got: %s", body)
		}
	})

	t.Run("GET request sets method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/component/test?q=test", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		body := w.Body.String()
		if !strings.Contains(body, "Method: GET") {
			t.Errorf("expected body to contain 'Method: GET', got: %s", body)
		}
	})
}

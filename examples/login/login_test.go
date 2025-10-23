package login_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginComponent(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("login form renders correctly", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find login form
		loginForm := pt.Page.Locator("form[hx-post='/component/login']")
		require.NotNil(t, loginForm)

		// Verify username input exists
		usernameInput := loginForm.Locator("input[name='username']")
		count, err := usernameInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify password input exists
		passwordInput := loginForm.Locator("input[name='password']")
		count, err = passwordInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify password input is type password
		inputType, err := passwordInput.GetAttribute("type")
		require.NoError(t, err)
		assert.Equal(t, "password", inputType)

		// Verify submit button exists
		submitBtn := loginForm.Locator("button[type='submit']")
		count, err = submitBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("login fails with wrong credentials", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in login form with wrong credentials
		usernameInput := pt.Page.Locator("input[name='username']")
		err := usernameInput.Fill("wronguser")
		require.NoError(t, err)

		passwordInput := pt.Page.Locator("input[name='password']")
		err = passwordInput.Fill("wrongpass")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/login'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify error message is displayed
		errorDiv := pt.Page.Locator("#login-result .alert-danger")
		count, err := errorDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify error message text
		errorText, err := errorDiv.TextContent()
		require.NoError(t, err)
		assert.Contains(t, errorText, "Invalid credentials")
	})

	t.Run("login fails with empty username", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in only password
		passwordInput := pt.Page.Locator("input[name='password']")
		err := passwordInput.Fill("password")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/login'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify error message is displayed
		errorDiv := pt.Page.Locator("#login-result .alert-danger")
		count, err := errorDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify error message text
		errorText, err := errorDiv.TextContent()
		require.NoError(t, err)
		assert.Contains(t, errorText, "Username and password are required")
	})

	t.Run("login fails with empty password", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in only username
		usernameInput := pt.Page.Locator("input[name='username']")
		err := usernameInput.Fill("demo")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/login'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify error message is displayed
		errorDiv := pt.Page.Locator("#login-result .alert-danger")
		count, err := errorDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify error message text
		errorText, err := errorDiv.TextContent()
		require.NoError(t, err)
		assert.Contains(t, errorText, "Username and password are required")
	})

	t.Run("login succeeds with correct credentials and redirects", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in login form with correct credentials
		usernameInput := pt.Page.Locator("input[name='username']")
		err := usernameInput.Fill("demo")
		require.NoError(t, err)

		passwordInput := pt.Page.Locator("input[name='password']")
		err = passwordInput.Fill("password")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/login'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for redirect (HTMX will redirect via HX-Redirect header)
		err = pt.Page.WaitForURL(server.URL+"/dashboard", playwright.PageWaitForURLOptions{})
		require.NoError(t, err)

		// Verify we're on the dashboard page
		currentURL := pt.Page.URL()
		assert.Contains(t, currentURL, "/dashboard")

		// Verify dashboard content
		dashboardHeading := pt.Page.Locator("h1")
		text, err := dashboardHeading.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Welcome to the Dashboard")

		// Verify success message
		successMsg := pt.Page.Locator(".success-message")
		count, err := successMsg.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("login sets HX-Redirect response header", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Set up response listener to capture headers
		var responseHeaders map[string]string
		pt.Page.OnResponse(func(response playwright.Response) {
			url := response.URL()
			if url == server.URL+"/component/login" {
				headers := response.Headers()
				responseHeaders = headers
			}
		})

		// Fill in login form with correct credentials
		usernameInput := pt.Page.Locator("input[name='username']")
		err := usernameInput.Fill("demo")
		require.NoError(t, err)

		passwordInput := pt.Page.Locator("input[name='password']")
		err = passwordInput.Fill("password")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/login'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for redirect
		err = pt.Page.WaitForURL(server.URL+"/dashboard", playwright.PageWaitForURLOptions{})
		require.NoError(t, err)

		// Verify HX-Redirect header was set
		assert.NotNil(t, responseHeaders)
		assert.Equal(t, "/dashboard", responseHeaders["hx-redirect"])
	})

	t.Run("login hint text is displayed", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify hint text is present
		hintText := pt.Page.Locator("text=Hint: Use username \"demo\" and password \"password\"")
		count, err := hintText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

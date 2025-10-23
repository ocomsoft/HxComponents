//go:build integration
// +build integration

package examples_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration tests the overall application integration.
func TestIntegration(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("homepage renders all components", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify page title
		title, err := pt.Page.Title()
		require.NoError(t, err)
		assert.Equal(t, "HTMX Component Registry Demo", title)

		// Verify main heading
		heading := pt.Page.Locator("h1")
		text, err := heading.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "HTMX Generic Component Registry")

		// Verify counter component exists
		counter := pt.Page.Locator(".counter-component")
		count, err := counter.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify embedded search component exists
		embeddedSearch := pt.Page.Locator(".search-results")
		count, err = embeddedSearch.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)

		// Verify search form exists
		searchForm := pt.Page.Locator("form[hx-post='/component/search']")
		count, err = searchForm.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify login form exists
		loginForm := pt.Page.Locator("form[hx-post='/component/login']")
		count, err = loginForm.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify profile form exists
		profileForm := pt.Page.Locator("form[hx-post='/component/profile']")
		count, err = profileForm.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("htmx script is loaded", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify HTMX script tag exists
		htmxScript := pt.Page.Locator("script[src*='htmx.org']")
		count, err := htmxScript.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify htmx is available in window
		htmxExists, err := pt.Page.Evaluate("() => typeof htmx !== 'undefined'")
		require.NoError(t, err)
		assert.True(t, htmxExists.(bool), "htmx should be loaded")
	})

	t.Run("all result containers exist", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify search result container
		searchResult := pt.Page.Locator("#search-result")
		count, err := searchResult.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify login result container
		loginResult := pt.Page.Locator("#login-result")
		count, err = loginResult.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify profile result container
		profileResult := pt.Page.Locator("#profile-result")
		count, err = profileResult.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify GET demo container
		getDemo := pt.Page.Locator("#get-demo")
		count, err = getDemo.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("how it works section is present", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify "How It Works" heading
		howItWorks := pt.Page.Locator("h2:has-text('How It Works')")
		count, err := howItWorks.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify explanation text contains key points
		explanation := pt.Page.Locator("text=/component/")
		count, err = explanation.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})

	t.Run("multiple components work together", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Use counter
		counterBtn := pt.Page.Locator(".counter-component button:has-text('+')")
		err := counterBtn.Click()
		require.NoError(t, err)
		pt.WaitForHTMX()

		// Use search
		searchQuery := pt.Page.Locator("input[name='q']")
		err = searchQuery.Fill("test search")
		require.NoError(t, err)

		searchSubmit := pt.Page.Locator("form[hx-post='/component/search'] button[type='submit']")
		err = searchSubmit.Click()
		require.NoError(t, err)
		pt.WaitForHTMX()

		// Verify search result appeared
		searchResult := pt.Page.Locator("#search-result .search-results")
		count, err := searchResult.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Use profile
		profileName := pt.Page.Locator("input[name='name']")
		err = profileName.Fill("Integration Test")
		require.NoError(t, err)

		profileEmail := pt.Page.Locator("input[name='email']")
		err = profileEmail.Fill("integration@test.com")
		require.NoError(t, err)

		profileSubmit := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = profileSubmit.Click()
		require.NoError(t, err)
		pt.WaitForHTMX()

		// Verify profile result appeared
		profileResult := pt.Page.Locator("#profile-result .alert-success")
		count, err = profileResult.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify counter is still working
		counterSpan := pt.Page.Locator(".counter-component span")
		text, err := counterSpan.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "1", text)
	})

	t.Run("component registry handles not found", func(t *testing.T) {
		// Navigate to non-existent component
		pt.Goto(server.URL + "/component/nonexistent")

		// Verify error message
		errorComponent := pt.Page.Locator("text=Component Not Found")
		count, err := errorComponent.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("dashboard page is accessible", func(t *testing.T) {
		// Navigate directly to dashboard
		pt.Goto(server.URL + "/dashboard")

		// Verify dashboard heading
		heading := pt.Page.Locator("h1")
		text, err := heading.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Welcome to the Dashboard")

		// Verify success message
		successMsg := pt.Page.Locator(".success-message")
		count, err := successMsg.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify back link exists
		backLink := pt.Page.Locator("a[href='/']")
		count, err = backLink.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("responsive grid layout", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify grid container exists
		grid := pt.Page.Locator(".grid")
		count, err := grid.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify grid contains all three component forms
		gridContainers := grid.Locator(".container")
		count, err = gridContainers.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, count, "Grid should contain search, login, and profile forms")
	})
}

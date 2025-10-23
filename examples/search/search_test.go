package search_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchComponent(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("search form renders correctly", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find search form
		searchForm := pt.Page.Locator("form[hx-post='/component/search']")
		require.NotNil(t, searchForm)

		// Verify query input exists
		queryInput := searchForm.Locator("input[name='q']")
		count, err := queryInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify limit input exists
		limitInput := searchForm.Locator("input[name='limit']")
		count, err = limitInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify submit button exists
		submitBtn := searchForm.Locator("button[type='submit']")
		count, err = submitBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("search submits and displays results", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in search form
		queryInput := pt.Page.Locator("input[name='q']")
		err := queryInput.Fill("golang htmx")
		require.NoError(t, err)

		limitInput := pt.Page.Locator("input[name='limit']")
		err = limitInput.Fill("15")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/search'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify results are displayed in the search-result div
		results := pt.Page.Locator("#search-result")

		// Verify query is displayed
		queryText := results.Locator("text=golang htmx")
		count, err := queryText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Query should be displayed in results")

		// Verify limit is displayed
		limitText := results.Locator("text=15 results")
		count, err = limitText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Limit should be displayed in results")
	})

	t.Run("search captures HTMX request headers", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in search form
		queryInput := pt.Page.Locator("input[name='q']")
		err := queryInput.Fill("test query")
		require.NoError(t, err)

		// Submit form (with hx-boost="true")
		submitBtn := pt.Page.Locator("form[hx-post='/component/search'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify "This is an HTMX request" is displayed
		results := pt.Page.Locator("#search-result")
		htmxIndicator := results.Locator("text=This is an HTMX request")
		count, err := htmxIndicator.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "HTMX request indicator should be present")

		// Note: hx-boost="true" on the form may not always set HX-Boosted header
		// depending on how the form is submitted, so we don't assert on boosted indicator
	})

	t.Run("search works via GET request", func(t *testing.T) {
		// Navigate directly to search component with query params
		pt.Goto(server.URL + "/component/search?q=get-test&limit=7")

		// Verify query is displayed
		queryText := pt.Page.Locator("text=get-test")
		count, err := queryText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Query should be displayed")

		// Verify limit is displayed
		limitText := pt.Page.Locator("text=7 results")
		count, err = limitText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Limit should be displayed")
	})

	t.Run("search component embedded in template works", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the embedded search component (rendered server-side)
		embeddedSearch := pt.Page.Locator(".search-results", playwright.PageLocatorOptions{
			HasText: "Initial search",
		})
		count, err := embeddedSearch.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Embedded search component should be present")

		// Verify it shows the initial query
		queryText := embeddedSearch.First().Locator("text=Initial search")
		count, err = queryText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Initial query should be displayed")

		// Verify it shows the initial limit
		limitText := embeddedSearch.First().Locator("text=5 results")
		count, err = limitText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Initial limit should be displayed")
	})

	t.Run("GET demo link works", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Click the GET demo link
		demoLink := pt.Page.Locator("a[hx-get='/component/search?q=golang&limit=5']")
		err := demoLink.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify results are in the get-demo div
		getDemo := pt.Page.Locator("#get-demo")
		queryText := getDemo.Locator("text=golang")
		count, err := queryText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Query should be in get-demo div")

		// Verify limit
		limitText := getDemo.Locator("text=5 results")
		count, err = limitText.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Limit should be in get-demo div")
	})
}

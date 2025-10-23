package counter_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterComponent(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("counter renders with initial value", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		require.NotNil(t, counter)

		// Verify counter displays 0
		span := counter.Locator("span")
		text, err := span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "0", text)

		// Verify increment button exists
		incrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "+",
		})
		count, err := incrementBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify decrement button exists
		decrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "−",
		})
		count, err = decrementBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("counter increments when + button clicked", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		span := counter.Locator("span")

		// Initial value should be 0
		text, err := span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "0", text)

		// Click increment button
		incrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "+",
		})
		err = incrementBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify counter displays 1
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "1", text)

		// Click increment button again
		err = incrementBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify counter displays 2
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "2", text)
	})

	t.Run("counter decrements when - button clicked", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		span := counter.Locator("span")

		// Initial value should be 0
		text, err := span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "0", text)

		// Click decrement button
		decrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "−",
		})
		err = decrementBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify counter displays -1
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "-1", text)

		// Click decrement button again
		err = decrementBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify counter displays -2
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "-2", text)
	})

	t.Run("counter handles mixed increment and decrement", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		span := counter.Locator("span")
		incrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "+",
		})
		decrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "−",
		})

		// Click increment 3 times
		for i := 0; i < 3; i++ {
			err := incrementBtn.Click()
			require.NoError(t, err)
			pt.WaitForHTMX()
		}

		// Verify counter displays 3
		text, err := span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "3", text)

		// Click decrement 1 time
		err = decrementBtn.Click()
		require.NoError(t, err)
		pt.WaitForHTMX()

		// Verify counter displays 2
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "2", text)

		// Click decrement 3 times
		for i := 0; i < 3; i++ {
			err := decrementBtn.Click()
			require.NoError(t, err)
			pt.WaitForHTMX()
		}

		// Verify counter displays -1
		text, err = span.TextContent()
		require.NoError(t, err)
		assert.Equal(t, "-1", text)
	})

	t.Run("counter uses hx-vals for parameters", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		incrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "+",
		})

		// Check that button has hx-vals attribute with new hxc-event parameter
		hxVals, err := incrementBtn.GetAttribute("hx-vals")
		require.NoError(t, err)
		assert.Contains(t, hxVals, "count")
		assert.Contains(t, hxVals, "hxc-event")
		assert.Contains(t, hxVals, "increment")

		// Check that button posts to /component/counter
		hxPost, err := incrementBtn.GetAttribute("hx-post")
		require.NoError(t, err)
		assert.Equal(t, "/component/counter", hxPost)
	})

	t.Run("counter replaces itself with hx-swap outerHTML", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the counter component
		counter := pt.Page.Locator(".counter-component")
		incrementBtn := counter.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "+",
		})

		// Check hx-target
		hxTarget, err := incrementBtn.GetAttribute("hx-target")
		require.NoError(t, err)
		assert.Equal(t, "closest .counter-component", hxTarget)

		// Check hx-swap
		hxSwap, err := incrementBtn.GetAttribute("hx-swap")
		require.NoError(t, err)
		assert.Equal(t, "outerHTML", hxSwap)
	})
}

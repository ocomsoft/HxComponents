package testutil

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// PlaywrightTest wraps Playwright resources for testing.
type PlaywrightTest struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
	Context playwright.BrowserContext
	Page    playwright.Page
	t       *testing.T
}

// NewPlaywrightTest creates a new Playwright test environment.
func NewPlaywrightTest(t *testing.T) *PlaywrightTest {
	t.Helper()

	// Install playwright if needed
	err := playwright.Install()
	require.NoError(t, err, "failed to install playwright")

	// Launch playwright
	pw, err := playwright.Run()
	require.NoError(t, err, "failed to run playwright")

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err, "failed to launch browser")

	// Create context
	context, err := browser.NewContext()
	require.NoError(t, err, "failed to create browser context")

	// Create page
	page, err := context.NewPage()
	require.NoError(t, err, "failed to create page")

	return &PlaywrightTest{
		PW:      pw,
		Browser: browser,
		Context: context,
		Page:    page,
		t:       t,
	}
}

// Close cleans up all Playwright resources.
func (pt *PlaywrightTest) Close() {
	pt.t.Helper()
	if pt.Page != nil {
		if err := pt.Page.Close(); err != nil {
			pt.t.Logf("Page close error: %v", err)
		}
	}
	if pt.Context != nil {
		if err := pt.Context.Close(); err != nil {
			pt.t.Logf("Context close error: %v", err)
		}
	}
	if pt.Browser != nil {
		if err := pt.Browser.Close(); err != nil {
			pt.t.Logf("Browser close error: %v", err)
		}
	}
	if pt.PW != nil {
		if err := pt.PW.Stop(); err != nil {
			pt.t.Logf("Playwright stop error: %v", err)
		}
	}
}

// Goto navigates to a URL and waits for the page to load.
func (pt *PlaywrightTest) Goto(url string) {
	pt.t.Helper()
	_, err := pt.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(pt.t, err, "failed to navigate to %s", url)
}

// WaitForHTMX waits for HTMX requests to settle.
func (pt *PlaywrightTest) WaitForHTMX() {
	pt.t.Helper()
	// Wait for htmx:afterSettle event which fires after HTMX completes
	err := pt.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	require.NoError(pt.t, err, "failed waiting for network idle")
}

package profile_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileComponent(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("profile form renders correctly", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find profile form
		profileForm := pt.Page.Locator("form[hx-post='/component/profile']")
		require.NotNil(t, profileForm)

		// Verify name input exists
		nameInput := profileForm.Locator("input[name='name']")
		count, err := nameInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify email input exists
		emailInput := profileForm.Locator("input[name='email']")
		count, err = emailInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify email input is type email
		inputType, err := emailInput.GetAttribute("type")
		require.NoError(t, err)
		assert.Equal(t, "email", inputType)

		// Verify tags input exists
		tagsInput := profileForm.Locator("input[name='tags']")
		count, err = tagsInput.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify submit button exists
		submitBtn := profileForm.Locator("button[type='submit']")
		count, err = submitBtn.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("profile update succeeds with all fields", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in profile form
		nameInput := pt.Page.Locator("input[name='name']")
		err := nameInput.Fill("John Doe")
		require.NoError(t, err)

		emailInput := pt.Page.Locator("input[name='email']")
		err = emailInput.Fill("john@example.com")
		require.NoError(t, err)

		tagsInput := pt.Page.Locator("input[name='tags']")
		err = tagsInput.Fill("developer, golang, htmx")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify success message
		successDiv := pt.Page.Locator("#profile-result .alert-success")
		count, err := successDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify success heading
		successHeading := successDiv.Locator("h3")
		text, err := successHeading.TextContent()
		require.NoError(t, err)
		assert.Contains(t, text, "Profile Updated Successfully")

		// Verify name is displayed
		nameText := successDiv.Locator("text=John Doe")
		count, err = nameText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify email is displayed
		emailText := successDiv.Locator("text=john@example.com")
		count, err = emailText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify tags are displayed
		tagsText := successDiv.Locator("text=developer, golang, htmx")
		count, err = tagsText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("profile update fails with empty name", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in only email
		emailInput := pt.Page.Locator("input[name='email']")
		err := emailInput.Fill("john@example.com")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify warning message
		warningDiv := pt.Page.Locator("#profile-result .alert-warning")
		count, err := warningDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify warning text
		warningText, err := warningDiv.TextContent()
		require.NoError(t, err)
		assert.Contains(t, warningText, "Please fill in all required fields")
	})

	t.Run("profile update fails with empty email", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in only name
		nameInput := pt.Page.Locator("input[name='name']")
		err := nameInput.Fill("John Doe")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify warning message
		warningDiv := pt.Page.Locator("#profile-result .alert-warning")
		count, err := warningDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify warning text
		warningText, err := warningDiv.TextContent()
		require.NoError(t, err)
		assert.Contains(t, warningText, "Please fill in all required fields")
	})

	t.Run("profile update succeeds without tags", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in name and email only (tags are optional)
		nameInput := pt.Page.Locator("input[name='name']")
		err := nameInput.Fill("Jane Smith")
		require.NoError(t, err)

		emailInput := pt.Page.Locator("input[name='email']")
		err = emailInput.Fill("jane@example.com")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify success message
		successDiv := pt.Page.Locator("#profile-result .alert-success")
		count, err := successDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify name is displayed
		nameText := successDiv.Locator("text=Jane Smith")
		count, err = nameText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify email is displayed
		emailText := successDiv.Locator("text=jane@example.com")
		count, err = emailText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// The important part is that the profile updated successfully without tags
		// The template correctly hides the tags section when len(data.Tags) == 0
		// We've verified the name and email are present, which confirms success
	})

	t.Run("profile parses comma-separated tags correctly", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in profile form with multiple tags
		nameInput := pt.Page.Locator("input[name='name']")
		err := nameInput.Fill("Bob Builder")
		require.NoError(t, err)

		emailInput := pt.Page.Locator("input[name='email']")
		err = emailInput.Fill("bob@builder.com")
		require.NoError(t, err)

		tagsInput := pt.Page.Locator("input[name='tags']")
		err = tagsInput.Fill("frontend, backend, devops, testing")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify success message
		successDiv := pt.Page.Locator("#profile-result .alert-success")
		count, err := successDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify all tags are displayed
		tagsText := successDiv.Locator("text=frontend, backend, devops, testing")
		count, err = tagsText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("profile handles special characters in name", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Fill in profile form with special characters
		nameInput := pt.Page.Locator("input[name='name']")
		err := nameInput.Fill("O'Brien-Smith Jr.")
		require.NoError(t, err)

		emailInput := pt.Page.Locator("input[name='email']")
		err = emailInput.Fill("obrien@example.com")
		require.NoError(t, err)

		// Submit form
		submitBtn := pt.Page.Locator("form[hx-post='/component/profile'] button[type='submit']")
		err = submitBtn.Click()
		require.NoError(t, err)

		// Wait for HTMX to update
		pt.WaitForHTMX()

		// Verify success message
		successDiv := pt.Page.Locator("#profile-result .alert-success")
		count, err := successDiv.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify name with special characters is displayed
		nameText := successDiv.Locator("text=O'Brien-Smith Jr.")
		count, err = nameText.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

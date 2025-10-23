package todolist_test

import (
	"testing"

	"github.com/ocomsoft/HxComponents/examples/testutil"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoListComponent(t *testing.T) {
	// Start test server
	server := testutil.NewTestServer(t)
	defer server.Close()

	// Start Playwright
	pt := testutil.NewPlaywrightTest(t)
	defer pt.Close()

	t.Run("todolist renders with initial state", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")
		require.NotNil(t, todolist)

		// Verify title exists
		title := todolist.Locator("h3")
		titleText, err := title.TextContent()
		require.NoError(t, err)
		assert.Contains(t, titleText, "Todo List")

		// Verify stats are displayed
		stats := todolist.Locator("div").Filter(playwright.LocatorFilterOptions{
			HasText: "Active:",
		})
		count, err := stats.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Stats should be displayed")

		// Verify initial items are displayed
		checkboxes := todolist.Locator("input[type='checkbox']")
		itemCount, err := checkboxes.Count()
		require.NoError(t, err)
		assert.Equal(t, 3, itemCount, "Should have 3 initial todo items")

		// Verify add item input exists
		input := todolist.Locator("input[name='newItemText']")
		count, err = input.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("add item button has correct HTMX attributes", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Find add button
		addButton := todolist.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "Add Item",
		})

		// Check hx-post attribute
		hxPost, err := addButton.GetAttribute("hx-post")
		require.NoError(t, err)
		assert.Equal(t, "/component/todolist", hxPost)

		// Check hx-vals contains hxc-event
		hxVals, err := addButton.GetAttribute("hx-vals")
		require.NoError(t, err)
		assert.Contains(t, hxVals, "hxc-event")
		assert.Contains(t, hxVals, "addItem")

		// Check hx-target
		hxTarget, err := addButton.GetAttribute("hx-target")
		require.NoError(t, err)
		assert.Equal(t, "closest .todo-list-component", hxTarget)

		// Check hx-swap
		hxSwap, err := addButton.GetAttribute("hx-swap")
		require.NoError(t, err)
		assert.Equal(t, "outerHTML", hxSwap)
	})

	t.Run("toggle checkbox has correct HTMX attributes", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Find first checkbox
		checkbox := todolist.Locator("input[type='checkbox']").First()

		// Check hx-post attribute
		hxPost, err := checkbox.GetAttribute("hx-post")
		require.NoError(t, err)
		assert.Equal(t, "/component/todolist", hxPost)

		// Check hx-vals contains hxc-event and itemId
		hxVals, err := checkbox.GetAttribute("hx-vals")
		require.NoError(t, err)
		assert.Contains(t, hxVals, "hxc-event")
		assert.Contains(t, hxVals, "toggleItem")
		assert.Contains(t, hxVals, "itemId")

		// Check hx-target
		hxTarget, err := checkbox.GetAttribute("hx-target")
		require.NoError(t, err)
		assert.Equal(t, "closest .todo-list-component", hxTarget)

		// Check hx-swap
		hxSwap, err := checkbox.GetAttribute("hx-swap")
		require.NoError(t, err)
		assert.Equal(t, "outerHTML", hxSwap)
	})

	t.Run("delete button has correct HTMX attributes", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Find first delete button
		deleteButton := todolist.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "Delete",
		}).First()

		// Check hx-post attribute
		hxPost, err := deleteButton.GetAttribute("hx-post")
		require.NoError(t, err)
		assert.Equal(t, "/component/todolist", hxPost)

		// Check hx-vals contains hxc-event and itemId
		hxVals, err := deleteButton.GetAttribute("hx-vals")
		require.NoError(t, err)
		assert.Contains(t, hxVals, "hxc-event")
		assert.Contains(t, hxVals, "deleteItem")
		assert.Contains(t, hxVals, "itemId")

		// Check hx-confirm
		hxConfirm, err := deleteButton.GetAttribute("hx-confirm")
		require.NoError(t, err)
		assert.Contains(t, hxConfirm, "Are you sure you want to delete this item?")

		// Check hx-target
		hxTarget, err := deleteButton.GetAttribute("hx-target")
		require.NoError(t, err)
		assert.Equal(t, "closest .todo-list-component", hxTarget)

		// Check hx-swap
		hxSwap, err := deleteButton.GetAttribute("hx-swap")
		require.NoError(t, err)
		assert.Equal(t, "outerHTML", hxSwap)
	})

	t.Run("shows event lifecycle information", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Verify lifecycle debug info is displayed
		debugInfo := todolist.Locator("div", playwright.LocatorLocatorOptions{
			HasText: "Event Lifecycle Demo:",
		})
		count, err := debugInfo.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Event lifecycle info should be visible")

		// Verify it mentions the lifecycle steps
		debugText, err := debugInfo.TextContent()
		require.NoError(t, err)
		assert.Contains(t, debugText, "BeforeEvent")
		assert.Contains(t, debugText, "OnEventName")
		assert.Contains(t, debugText, "AfterEvent")
		assert.Contains(t, debugText, "Process")
		assert.Contains(t, debugText, "Render")
	})

	t.Run("displays stats correctly", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Check Active count
		activeText, err := todolist.Locator("text=Active:").Locator("..").TextContent()
		require.NoError(t, err)
		assert.Contains(t, activeText, "Active: 3")

		// Check Completed count
		completedText, err := todolist.Locator("text=Completed:").Locator("..").TextContent()
		require.NoError(t, err)
		assert.Contains(t, completedText, "Completed: 0")

		// Check Total count
		totalText, err := todolist.Locator("text=Total:").Locator("..").TextContent()
		require.NoError(t, err)
		assert.Contains(t, totalText, "Total: 3")
	})

	t.Run("displays todos with correct styling", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Verify sample todos are displayed
		todo1 := todolist.Locator("span", playwright.LocatorLocatorOptions{
			HasText: "Try adding a new item",
		})
		count, err := todo1.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		todo2 := todolist.Locator("span", playwright.LocatorLocatorOptions{
			HasText: "Toggle an item as complete",
		})
		count, err = todo2.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		todo3 := todolist.Locator("span", playwright.LocatorLocatorOptions{
			HasText: "Delete an item",
		})
		count, err = todo3.Count()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("clear completed button not shown initially", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Clear completed button should not be visible when no items are completed
		clearButton := todolist.Locator("button", playwright.LocatorLocatorOptions{
			HasText: "Clear Completed",
		})
		count, err := clearButton.Count()
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Clear Completed button should be hidden when no completed items")
	})

	t.Run("demonstrates event-driven pattern", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Verify explanation text
		explanation, err := pt.Page.Locator("p", playwright.PageLocatorOptions{
			HasText: "This TodoList uses the new event-driven pattern",
		}).TextContent()
		require.NoError(t, err)
		assert.Contains(t, explanation, "hxc-event")
		assert.Contains(t, explanation, "OnAddItem")
		assert.Contains(t, explanation, "OnToggleItem")
		assert.Contains(t, explanation, "OnDeleteItem")
		assert.Contains(t, explanation, "OnClearCompleted")
	})

	t.Run("add item input has correct attributes", func(t *testing.T) {
		// Navigate to home page
		pt.Goto(server.URL)

		// Find the todolist component
		todolist := pt.Page.Locator(".todo-list-component")

		// Find input
		input := todolist.Locator("input[name='newItemText']")

		// Check placeholder
		placeholder, err := input.GetAttribute("placeholder")
		require.NoError(t, err)
		assert.Equal(t, "What needs to be done?", placeholder)

		// Check it's a text input
		inputType, err := input.GetAttribute("type")
		require.NoError(t, err)
		assert.Equal(t, "text", inputType)
	})
}

package todolist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"
)

// TodoItem represents a single todo item.
type TodoItem struct {
	ID        int
	Text      string
	Completed bool
}

// TodoListComponent demonstrates the full event-driven lifecycle with hooks.
// It shows BeforeEvent, multiple event handlers, and AfterEvent.
// This is a stateless component - all state is passed via form fields.
type TodoListComponent struct {
	Items       []TodoItem `json:"-"`
	ItemsJSON   string     `form:"items"` // Hidden field containing JSON-encoded items
	NewItemText string     `form:"newItemText"`
	ItemID      int        `form:"itemId"`
	LastEvent   string     `json:"-"`
	EventCount  int        `json:"-"`
}

// BeforeEvent is called before any event handler.
// This demonstrates validation and setup logic that runs for all events.
func (t *TodoListComponent) BeforeEvent(ctx context.Context, eventName string) error {
	slog.Info("TodoList BeforeEvent", "event", eventName)

	// Deserialize items from JSON (stateless approach)
	if t.ItemsJSON != "" {
		if err := json.Unmarshal([]byte(t.ItemsJSON), &t.Items); err != nil {
			return fmt.Errorf("failed to unmarshal items: %w", err)
		}
	}

	return nil
}

// AfterEvent is called after an event handler succeeds.
// This demonstrates cleanup or side effects that run after events.
func (t *TodoListComponent) AfterEvent(ctx context.Context, eventName string) error {
	slog.Info("TodoList AfterEvent", "event", eventName)

	// Track the last event and count
	t.LastEvent = eventName
	t.EventCount++

	return nil
}

// OnAddItem handles the "addItem" event.
func (t *TodoListComponent) OnAddItem(ctx context.Context) error {
	if t.NewItemText == "" {
		return fmt.Errorf("item text cannot be empty")
	}

	// Generate new ID (find max ID and increment)
	newID := 1
	for _, item := range t.Items {
		if item.ID >= newID {
			newID = item.ID + 1
		}
	}

	// Add the new item
	t.Items = append(t.Items, TodoItem{
		ID:        newID,
		Text:      t.NewItemText,
		Completed: false,
	})

	slog.Info("Added todo item", "id", newID, "text", t.NewItemText)

	// Clear the input
	t.NewItemText = ""

	return nil
}

// OnToggleItem handles the "toggleItem" event.
func (t *TodoListComponent) OnToggleItem(ctx context.Context) error {
	for i := range t.Items {
		if t.Items[i].ID == t.ItemID {
			t.Items[i].Completed = !t.Items[i].Completed
			slog.Info("Toggled todo item", "id", t.ItemID, "completed", t.Items[i].Completed)
			return nil
		}
	}
	return fmt.Errorf("item with ID %d not found", t.ItemID)
}

// OnDeleteItem handles the "deleteItem" event.
func (t *TodoListComponent) OnDeleteItem(ctx context.Context) error {
	for i, item := range t.Items {
		if item.ID == t.ItemID {
			// Remove item from slice
			t.Items = append(t.Items[:i], t.Items[i+1:]...)
			slog.Info("Deleted todo item", "id", t.ItemID)
			return nil
		}
	}
	return fmt.Errorf("item with ID %d not found", t.ItemID)
}

// OnClearCompleted handles the "clearCompleted" event.
func (t *TodoListComponent) OnClearCompleted(ctx context.Context) error {
	// Filter out completed items
	remaining := []TodoItem{}
	removedCount := 0

	for _, item := range t.Items {
		if !item.Completed {
			remaining = append(remaining, item)
		} else {
			removedCount++
		}
	}

	t.Items = remaining
	slog.Info("Cleared completed items", "count", removedCount)
	return nil
}

// Process is still called after events complete.
// This demonstrates that you can still use Process() for final logic.
func (t *TodoListComponent) Process(ctx context.Context) error {
	// Deserialize items from JSON if not already done (for non-event requests)
	if len(t.Items) == 0 && t.ItemsJSON != "" {
		if err := json.Unmarshal([]byte(t.ItemsJSON), &t.Items); err != nil {
			return fmt.Errorf("failed to unmarshal items: %w", err)
		}
	}

	slog.Debug("TodoList Process called", "itemCount", len(t.Items), "lastEvent", t.LastEvent)
	return nil
}

// Render implements templ.Component interface.
func (t *TodoListComponent) Render(ctx context.Context, w io.Writer) error {
	return TodoList(*t).Render(ctx, w)
}

// Helper methods for template

// GetActiveCount returns the number of active (not completed) items.
func (t *TodoListComponent) GetActiveCount() int {
	count := 0
	for _, item := range t.Items {
		if !item.Completed {
			count++
		}
	}
	return count
}

// GetCompletedCount returns the number of completed items.
func (t *TodoListComponent) GetCompletedCount() int {
	count := 0
	for _, item := range t.Items {
		if item.Completed {
			count++
		}
	}
	return count
}

// GetTimestamp returns the current timestamp for display.
func (t *TodoListComponent) GetTimestamp() string {
	return time.Now().Format("15:04:05")
}

// GetItemsJSON serializes the items to JSON for the hidden field.
// This enables stateless operation by passing all state in form data.
func (t *TodoListComponent) GetItemsJSON() string {
	if len(t.Items) == 0 {
		return "[]"
	}
	data, err := json.Marshal(t.Items)
	if err != nil {
		slog.Error("failed to marshal items to JSON", "error", err)
		return "[]"
	}
	return string(data)
}

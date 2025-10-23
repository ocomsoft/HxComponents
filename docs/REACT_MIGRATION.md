# Migrating from React to HxComponents

This guide demonstrates how to convert React components to the HxComponents architecture, mapping React concepts to Go/HTMX patterns.

## Architecture Comparison

| React Concept | HxComponents Equivalent |
|---------------|-------------------------|
| `useState()` | Struct fields with `form` tags |
| `props` | Struct fields (passed via form data or initial state) |
| `methods/functions` | Struct methods (event handlers, helpers) |
| `useMemo()`, `useCallback()` | Struct methods called from templates |
| `onClick`, `onChange` | HTMX attributes (`hx-post`, `hx-vals`, `hxc-event`) |
| Controlled inputs | Form inputs with `name` attribute + HTMX |
| `useEffect()` | `BeforeEvent()` or `AfterEvent()` hooks |
| `useEffect(..., [])` | `BeforeEvent()` with conditional logic |
| Context API | Session storage, context values, or service layer |
| Event handlers | Event methods (`On{EventName}`) |

## Example 1: Simple Counter

### React Component

```jsx
import React, { useState, useMemo } from 'react';

function Counter() {
  const [count, setCount] = useState(0);

  // Computed value similar to useMemo
  const doubled = useMemo(() => count * 2, [count]);

  const increment = () => setCount(count + 1);
  const decrement = () => setCount(count - 1);

  return (
    <div className="counter">
      <button onClick={decrement}>−</button>
      <span>{count}</span>
      <button onClick={increment}>+</button>
      <p>Double: {doubled}</p>
    </div>
  );
}

export default Counter;
```

### HxComponents Equivalent

**counter.go:**
```go
package counter

import (
	"context"
	"io"
)

// CounterComponent - equivalent to React state and functions
type CounterComponent struct {
	Count int `form:"count"` // useState(0)
}

// OnIncrement - equivalent to React's increment function
// Called automatically when hxc-event=increment
func (c *CounterComponent) OnIncrement() error {
	c.Count++
	return nil
}

// OnDecrement - equivalent to React's decrement function
// Called automatically when hxc-event=decrement
func (c *CounterComponent) OnDecrement() error {
	c.Count--
	return nil
}

// Doubled - equivalent to React's useMemo for derived state
// Called from template as needed
func (c *CounterComponent) Doubled() int {
	return c.Count * 2
}

// Render implements templ.Component interface
func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {
	return Counter(*c).Render(ctx, w)
}
```

**counter.templ:**
```templ
package counter

import "fmt"

templ Counter(data CounterComponent) {
	<div class="counter">
		<!-- onClick={decrement} becomes hx-post with hxc-event -->
		<button
			hx-post="/component/counter"
			hx-vals={ fmt.Sprintf(`{"count": %d, "hxc-event": "decrement"}`, data.Count) }
			hx-target="closest .counter"
			hx-swap="outerHTML"
		>
			−
		</button>

		<!-- {count} becomes { fmt.Sprint(data.Count) } -->
		<span>{ fmt.Sprint(data.Count) }</span>

		<!-- onClick={increment} becomes hx-post with hxc-event -->
		<button
			hx-post="/component/counter"
			hx-vals={ fmt.Sprintf(`{"count": %d, "hxc-event": "increment"}`, data.Count) }
			hx-target="closest .counter"
			hx-swap="outerHTML"
		>
			+
		</button>

		<!-- {doubled} becomes method call { fmt.Sprint(data.Doubled()) } -->
		<p>Double: { fmt.Sprint(data.Doubled()) }</p>
	</div>
}
```

## Example 2: Todo List with Effects

### React Component

```jsx
import React, { useState, useEffect } from 'react';

function TodoList({ title = 'My Todos' }) {
  const [items, setItems] = useState([]);
  const [newItemText, setNewItemText] = useState('');

  // useEffect on mount
  useEffect(() => {
    console.log('TodoList mounted');
    // Could load data from API here
  }, []);

  // useEffect on update
  useEffect(() => {
    console.log('TodoList updated', items.length);
  }, [items]);

  const activeCount = items.filter(i => !i.completed).length;
  const completedCount = items.filter(i => i.completed).length;

  const addItem = () => {
    if (!newItemText) return;

    setItems([...items, {
      id: Date.now(),
      text: newItemText,
      completed: false
    }]);
    setNewItemText('');
  };

  const toggleItem = (id) => {
    setItems(items.map(item =>
      item.id === id ? { ...item, completed: !item.completed } : item
    ));
  };

  const deleteItem = (id) => {
    setItems(items.filter(i => i.id !== id));
  };

  const clearCompleted = () => {
    setItems(items.filter(i => !i.completed));
  };

  return (
    <div className="todo-list">
      <h3>{title}</h3>

      <input
        value={newItemText}
        onChange={(e) => setNewItemText(e.target.value)}
        onKeyPress={(e) => e.key === 'Enter' && addItem()}
        placeholder="Add item"
      />
      <button onClick={addItem}>Add</button>

      <ul>
        {items.map(item => (
          <li key={item.id}>
            <input
              type="checkbox"
              checked={item.completed}
              onChange={() => toggleItem(item.id)}
            />
            <span className={item.completed ? 'completed' : ''}>
              {item.text}
            </span>
            <button onClick={() => deleteItem(item.id)}>Delete</button>
          </li>
        ))}
      </ul>

      <p>Active: {activeCount} | Completed: {completedCount}</p>
      {completedCount > 0 && (
        <button onClick={clearCompleted}>Clear Completed</button>
      )}
    </div>
  );
}

export default TodoList;
```

### HxComponents Equivalent

**todolist.go:**
```go
package todolist

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type TodoItem struct {
	ID        int
	Text      string
	Completed bool
}

// TodoListComponent - combines React's props, state, and functions
type TodoListComponent struct {
	// Props (like React props)
	Title string `json:"-"`

	// State (like useState)
	Items       []TodoItem `json:"-"`
	NewItemText string     `form:"newItemText"`
	ItemID      int        `form:"itemId"`

	// Tracking (used in lifecycle hooks)
	LastEvent  string `json:"-"`
	EventCount int    `json:"-"`
}

// BeforeEvent - equivalent to React's useEffect(..., [])
// Called before any event handler
func (t *TodoListComponent) BeforeEvent(eventName string) error {
	slog.Info("TodoList event starting", "event", eventName)

	// This is like useEffect on mount
	// You could validate authentication here
	// You could load data from database here

	return nil
}

// AfterEvent - equivalent to React's useEffect(..., [items])
// Called after successful event handler
func (t *TodoListComponent) AfterEvent(eventName string) error {
	slog.Info("TodoList event completed", "event", eventName, "itemCount", len(t.Items))

	t.LastEvent = eventName
	t.EventCount++

	// You could trigger webhooks, send notifications, etc.
	// This is like useEffect that runs after state changes

	return nil
}

// OnAddItem - equivalent to React's addItem function
// Automatically called when hxc-event=addItem
func (t *TodoListComponent) OnAddItem() error {
	if t.NewItemText == "" {
		return fmt.Errorf("item text cannot be empty")
	}

	// Generate new ID
	newID := len(t.Items) + 1
	for _, item := range t.Items {
		if item.ID >= newID {
			newID = item.ID + 1
		}
	}

	// Add the new item (like setItems([...items, newItem]))
	t.Items = append(t.Items, TodoItem{
		ID:        newID,
		Text:      t.NewItemText,
		Completed: false,
	})

	// Clear the input (like setNewItemText(''))
	t.NewItemText = ""

	slog.Info("Added todo item", "id", newID)
	return nil
}

// OnToggleItem - equivalent to React's toggleItem function
func (t *TodoListComponent) OnToggleItem() error {
	for i := range t.Items {
		if t.Items[i].ID == t.ItemID {
			t.Items[i].Completed = !t.Items[i].Completed
			slog.Info("Toggled todo item", "id", t.ItemID, "completed", t.Items[i].Completed)
			return nil
		}
	}
	return fmt.Errorf("item with ID %d not found", t.ItemID)
}

// OnDeleteItem - equivalent to React's deleteItem function
func (t *TodoListComponent) OnDeleteItem() error {
	for i, item := range t.Items {
		if item.ID == t.ItemID {
			t.Items = append(t.Items[:i], t.Items[i+1:]...)
			slog.Info("Deleted todo item", "id", t.ItemID)
			return nil
		}
	}
	return fmt.Errorf("item with ID %d not found", t.ItemID)
}

// OnClearCompleted - equivalent to React's clearCompleted function
func (t *TodoListComponent) OnClearCompleted() error {
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

// GetActiveCount - equivalent to React's derived state (activeCount)
func (t *TodoListComponent) GetActiveCount() int {
	count := 0
	for _, item := range t.Items {
		if !item.Completed {
			count++
		}
	}
	return count
}

// GetCompletedCount - equivalent to React's derived state (completedCount)
func (t *TodoListComponent) GetCompletedCount() int {
	count := 0
	for _, item := range t.Items {
		if item.Completed {
			count++
		}
	}
	return count
}

// Render implements templ.Component interface
func (t *TodoListComponent) Render(ctx context.Context, w io.Writer) error {
	return TodoList(*t).Render(ctx, w)
}
```

**todolist.templ:**
```templ
package todolist

import "fmt"

templ TodoList(data TodoListComponent) {
	<div class="todo-list">
		<!-- {title} becomes { data.Title } -->
		<h3>{ data.Title }</h3>

		<!-- value + onChange becomes name attribute + hx-include -->
		<input
			type="text"
			name="newItemText"
			value={ data.NewItemText }
			placeholder="Add item"
		/>

		<!-- onClick becomes hx-post with hxc-event -->
		<button
			hx-post="/component/todolist"
			hx-include="[name='newItemText']"
			hx-vals='{"hxc-event": "addItem"}'
			hx-target="closest .todo-list"
			hx-swap="outerHTML"
		>
			Add
		</button>

		<ul>
			<!-- items.map() becomes Go for loop -->
			for _, item := range data.Items {
				<li>
					<!-- checked and onChange become hx-post -->
					<input
						type="checkbox"
						checked?={ item.Completed }
						hx-post="/component/todolist"
						hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "toggleItem"}`, item.ID) }
						hx-target="closest .todo-list"
						hx-swap="outerHTML"
					/>

					<!-- className conditional becomes style attribute -->
					<span style={ fmt.Sprintf("%s",
						func() string {
							if item.Completed {
								return "text-decoration: line-through;"
							}
							return ""
						}()) }>
						{ item.Text }
					</span>

					<button
						hx-post="/component/todolist"
						hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "deleteItem"}`, item.ID) }
						hx-target="closest .todo-list"
						hx-swap="outerHTML"
					>
						Delete
					</button>
				</li>
			}
		</ul>

		<!-- {activeCount} becomes method call { fmt.Sprint(data.GetActiveCount()) } -->
		<p>
			Active: { fmt.Sprint(data.GetActiveCount()) } |
			Completed: { fmt.Sprint(data.GetCompletedCount()) }
		</p>

		<!-- conditional rendering becomes Go if statement -->
		if data.GetCompletedCount() > 0 {
			<button
				hx-post="/component/todolist"
				hx-vals='{"hxc-event": "clearCompleted"}'
				hx-target="closest .todo-list"
				hx-swap="outerHTML"
			>
				Clear Completed
			</button>
		}
	</div>
}
```

## Example 3: Form with Custom Hooks Pattern

### React Component

```jsx
import React, { useState, useEffect } from 'react';

// Custom hook pattern
function useFormValidation(initialEmail, initialPassword) {
  const [email, setEmail] = useState(initialEmail);
  const [password, setPassword] = useState(initialPassword);
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  useEffect(() => {
    validateEmail(email);
  }, [email]);

  useEffect(() => {
    validatePassword(password);
  }, [password]);

  const validateEmail = (value) => {
    if (!value) {
      setEmailError('Email is required');
    } else if (!value.includes('@')) {
      setEmailError('Invalid email');
    } else {
      setEmailError('');
    }
  };

  const validatePassword = (value) => {
    if (!value) {
      setPasswordError('Password is required');
    } else if (value.length < 8) {
      setPasswordError('Password must be 8+ characters');
    } else {
      setPasswordError('');
    }
  };

  const isValid = !emailError && !passwordError && email && password;

  return {
    email,
    setEmail,
    password,
    setPassword,
    emailError,
    passwordError,
    isValid
  };
}

function UserForm() {
  const {
    email,
    setEmail,
    password,
    setPassword,
    emailError,
    passwordError,
    isValid
  } = useFormValidation('', '');

  const handleSubmit = () => {
    if (isValid) {
      console.log('Submit', { email, password });
    }
  };

  return (
    <div className="user-form">
      <input
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      {emailError && <p className="error">{emailError}</p>}

      <input
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        type="password"
        placeholder="Password"
      />
      {passwordError && <p className="error">{passwordError}</p>}

      <button onClick={handleSubmit} disabled={!isValid}>
        Submit
      </button>
    </div>
  );
}

export default UserForm;
```

### HxComponents Equivalent

**userform.go:**
```go
package userform

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type UserFormComponent struct {
	Email         string `form:"email"`
	Password      string `form:"password"`
	EmailError    string `json:"-"`
	PasswordError string `json:"-"`
}

// BeforeEvent - equivalent to React's useEffect for validation
// This replaces the custom hook pattern
func (f *UserFormComponent) BeforeEvent(eventName string) error {
	f.validateEmail()
	f.validatePassword()
	return nil
}

// OnSubmit - equivalent to React's handleSubmit
func (f *UserFormComponent) OnSubmit() error {
	// Validation already done in BeforeEvent
	if !f.IsValid() {
		return fmt.Errorf("form is not valid")
	}

	// Process the submission
	// This is where you'd save to database, send email, etc.

	return nil
}

// OnValidate - trigger validation manually (like React input events)
func (f *UserFormComponent) OnValidate() error {
	// Validation happens in BeforeEvent
	// This event just triggers a re-render
	return nil
}

// validateEmail - equivalent to React's validateEmail function
func (f *UserFormComponent) validateEmail() {
	if f.Email == "" {
		f.EmailError = "Email is required"
	} else if !strings.Contains(f.Email, "@") {
		f.EmailError = "Invalid email"
	} else {
		f.EmailError = ""
	}
}

// validatePassword - equivalent to React's validatePassword function
func (f *UserFormComponent) validatePassword() {
	if f.Password == "" {
		f.PasswordError = "Password is required"
	} else if len(f.Password) < 8 {
		f.PasswordError = "Password must be 8+ characters"
	} else {
		f.PasswordError = ""
	}
}

// IsValid - equivalent to React's derived isValid state
func (f *UserFormComponent) IsValid() bool {
	return f.EmailError == "" && f.PasswordError == "" &&
		f.Email != "" && f.Password != ""
}

func (f *UserFormComponent) Render(ctx context.Context, w io.Writer) error {
	return UserForm(*f).Render(ctx, w)
}
```

**userform.templ:**
```templ
package userform

templ UserForm(data UserFormComponent) {
	<div class="user-form">
		<!-- value + onChange with validation -->
		<input
			type="email"
			name="email"
			value={ data.Email }
			placeholder="Email"
			hx-post="/component/userform"
			hx-trigger="change"
			hx-vals='{"hxc-event": "validate"}'
			hx-include="closest .user-form input"
			hx-target="closest .user-form"
			hx-swap="outerHTML"
		/>

		<!-- {emailError && <p>} becomes if statement -->
		if data.EmailError != "" {
			<p class="error">{ data.EmailError }</p>
		}

		<input
			type="password"
			name="password"
			value={ data.Password }
			placeholder="Password"
			hx-post="/component/userform"
			hx-trigger="change"
			hx-vals='{"hxc-event": "validate"}'
			hx-include="closest .user-form input"
			hx-target="closest .user-form"
			hx-swap="outerHTML"
		/>

		if data.PasswordError != "" {
			<p class="error">{ data.PasswordError }</p>
		}

		<!-- disabled={!isValid} becomes disabled? -->
		<button
			disabled?={ !data.IsValid() }
			hx-post="/component/userform"
			hx-vals='{"hxc-event": "submit"}'
			hx-include="closest .user-form input"
			hx-target="closest .user-form"
			hx-swap="outerHTML"
		>
			Submit
		</button>
	</div>
}
```

## Key Migration Patterns

### 1. State Management (useState)

**React:**
```jsx
const [count, setCount] = useState(0);
setCount(count + 1);
```

**HxComponents:**
```go
type Component struct {
	Count int `form:"count"`
}

func (c *Component) OnIncrement() error {
	c.Count++
	return nil
}
```

### 2. Props

**React:**
```jsx
function MyComponent({ title, count }) {
  return <h1>{title}: {count}</h1>;
}
```

**HxComponents:**
```go
type MyComponent struct {
	Title string `json:"-"`
	Count int    `json:"-"`
}
```

```templ
templ MyComponent(data MyComponent) {
	<h1>{ data.Title }: { fmt.Sprint(data.Count) }</h1>
}
```

### 3. Event Handlers

**React:**
```jsx
<button onClick={handleClick}>Click</button>
```

**HxComponents:**
```templ
<button
	hx-post="/component/mycomponent"
	hx-vals='{"hxc-event": "handleClick"}'
	hx-target="closest .container"
	hx-swap="outerHTML"
>
	Click
</button>
```

```go
func (c *MyComponent) OnHandleClick() error {
	// Handle click
	return nil
}
```

### 4. Conditional Rendering

**React:**
```jsx
{showMessage && <p>{message}</p>}
```

**HxComponents:**
```templ
if data.ShowMessage {
	<p>{ data.Message }</p>
}
```

### 5. List Rendering

**React:**
```jsx
{items.map(item => (
  <li key={item.id}>{item.name}</li>
))}
```

**HxComponents:**
```templ
for _, item := range data.Items {
	<li>{ item.Name }</li>
}
```

### 6. Derived State (useMemo)

**React:**
```jsx
const fullName = useMemo(() =>
  firstName + ' ' + lastName,
  [firstName, lastName]
);
```

**HxComponents:**
```go
func (c *Component) FullName() string {
	return c.FirstName + " " + c.LastName
}
```

```templ
<p>{ data.FullName() }</p>
```

### 7. Effects (useEffect)

**React:**
```jsx
useEffect(() => {
  console.log('Component mounted');
  // Load data
}, []);

useEffect(() => {
  console.log('Count changed');
}, [count]);
```

**HxComponents:**
```go
func (c *Component) BeforeEvent(eventName string) error {
	// Like useEffect on mount or before updates
	log.Println("Event starting:", eventName)
	return nil
}

func (c *Component) AfterEvent(eventName string) error {
	// Like useEffect after updates
	log.Println("Event completed:", eventName)
	return nil
}
```

### 8. Context API

**React:**
```jsx
const UserContext = React.createContext();

function Parent() {
  return (
    <UserContext.Provider value={user}>
      <Child />
    </UserContext.Provider>
  );
}

function Child() {
  const user = useContext(UserContext);
}
```

**HxComponents:**
```go
// Use context.Context or service layer
func (c *Component) BeforeEvent(eventName string) error {
	// Get user from session, database, or context
	user := getUserFromSession()
	c.CurrentUser = user
	return nil
}
```

## Benefits of HxComponents over React

1. **Type Safety**: Go's type system catches errors at compile time
2. **Simpler Mental Model**: No hooks rules, no dependency arrays to manage
3. **Performance**: Server-side rendering with minimal JavaScript
4. **SEO Friendly**: All content rendered on server
5. **Simpler Deployment**: Single binary, no build step
6. **Better Security**: Business logic stays on server
7. **No Virtual DOM**: Direct HTML updates via HTMX

## When to Use Each

**Use React when:**
- You need rich client-side interactivity
- You want offline-first capabilities
- You need the ecosystem (npm packages, UI libraries)
- Complex client-side state management is required

**Use HxComponents when:**
- You want simpler architecture
- SEO is important
- You prefer type safety and compile-time errors
- You want to avoid JavaScript complexity
- Server-side rendering fits your use case
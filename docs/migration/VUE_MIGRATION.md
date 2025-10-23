# Migrating from Vue 2 to HxComponents

This guide demonstrates how to convert Vue 2 components to the HxComponents architecture, mapping Vue concepts to Go/HTMX patterns.

## Architecture Comparison

| Vue 2 Concept | HxComponents Equivalent |
|---------------|-------------------------|
| `data()` | Struct fields with `form` tags |
| `props` | Struct fields (passed via form data or initial state) |
| `methods` | Struct methods (event handlers, helpers) |
| `computed` | Struct methods called from templates |
| `@click`, `@input` | HTMX attributes (`hx-post`, `hx-vals`, `hxc-event`) |
| `v-model` | Form inputs with `name` attribute + HTMX |
| `emit()` | Server-side event handlers (`On{EventName}`) |
| `beforeCreate`, `created` | `BeforeEvent()` hook |
| `beforeUpdate`, `updated` | `AfterEvent()` hook |
| `watch` | Event handlers that compare values |

## Example 1: Simple Counter

### Vue 2 Component

```vue
<template>
  <div class="counter">
    <button @click="decrement">−</button>
    <span>{{ count }}</span>
    <button @click="increment">+</button>
    <p>Double: {{ doubled }}</p>
  </div>
</template>

<script>
export default {
  name: 'Counter',
  data() {
    return {
      count: 0
    }
  },
  computed: {
    doubled() {
      return this.count * 2
    }
  },
  methods: {
    increment() {
      this.count++
    },
    decrement() {
      this.count--
    }
  }
}
</script>
```

### HxComponents Equivalent

**counter.go:**
```go
package counter

import (
	"context"
	"io"
)

// CounterComponent - equivalent to Vue's data() and methods
type CounterComponent struct {
	Count int `form:"count"` // data() property
}

// OnIncrement - equivalent to Vue's methods.increment
// Called automatically when hxc-event=increment
func (c *CounterComponent) OnIncrement() error {
	c.Count++
	return nil
}

// OnDecrement - equivalent to Vue's methods.decrement
// Called automatically when hxc-event=decrement
func (c *CounterComponent) OnDecrement() error {
	c.Count--
	return nil
}

// Doubled - equivalent to Vue's computed.doubled
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
		<!-- @click="decrement" becomes hx-post with hxc-event -->
		<button
			hx-post="/component/counter"
			hx-vals={ fmt.Sprintf(`{"count": %d, "hxc-event": "decrement"}`, data.Count) }
			hx-target="closest .counter"
			hx-swap="outerHTML"
		>
			−
		</button>

		<!-- {{ count }} becomes { fmt.Sprint(data.Count) } -->
		<span>{ fmt.Sprint(data.Count) }</span>

		<!-- @click="increment" becomes hx-post with hxc-event -->
		<button
			hx-post="/component/counter"
			hx-vals={ fmt.Sprintf(`{"count": %d, "hxc-event": "increment"}`, data.Count) }
			hx-target="closest .counter"
			hx-swap="outerHTML"
		>
			+
		</button>

		<!-- {{ doubled }} becomes method call { fmt.Sprint(data.Doubled()) } -->
		<p>Double: { fmt.Sprint(data.Doubled()) }</p>
	</div>
}
```

## Example 2: Todo List with Props and Events

### Vue 2 Component

```vue
<template>
  <div class="todo-list">
    <h3>{{ title }}</h3>

    <input v-model="newItemText" @keyup.enter="addItem" placeholder="Add item">
    <button @click="addItem">Add</button>

    <ul>
      <li v-for="item in items" :key="item.id">
        <input
          type="checkbox"
          :checked="item.completed"
          @change="toggleItem(item.id)"
        >
        <span :class="{ completed: item.completed }">{{ item.text }}</span>
        <button @click="deleteItem(item.id)">Delete</button>
      </li>
    </ul>

    <p>Active: {{ activeCount }} | Completed: {{ completedCount }}</p>
    <button v-if="completedCount > 0" @click="clearCompleted">
      Clear Completed
    </button>
  </div>
</template>

<script>
export default {
  name: 'TodoList',
  props: {
    title: {
      type: String,
      default: 'My Todos'
    }
  },
  data() {
    return {
      items: [],
      newItemText: ''
    }
  },
  computed: {
    activeCount() {
      return this.items.filter(i => !i.completed).length
    },
    completedCount() {
      return this.items.filter(i => i.completed).length
    }
  },
  methods: {
    addItem() {
      if (!this.newItemText) return

      this.items.push({
        id: Date.now(),
        text: this.newItemText,
        completed: false
      })
      this.newItemText = ''
      this.$emit('item-added')
    },
    toggleItem(id) {
      const item = this.items.find(i => i.id === id)
      if (item) {
        item.completed = !item.completed
        this.$emit('item-toggled', id)
      }
    },
    deleteItem(id) {
      this.items = this.items.filter(i => i.id !== id)
      this.$emit('item-deleted', id)
    },
    clearCompleted() {
      this.items = this.items.filter(i => !i.completed)
      this.$emit('completed-cleared')
    }
  },
  created() {
    console.log('TodoList created')
  },
  updated() {
    console.log('TodoList updated')
  }
}
</script>
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

// TodoListComponent - combines Vue's props, data, methods, and computed
type TodoListComponent struct {
	// Props (could be set from initial state or parent component)
	Title string `json:"-"`

	// Data
	Items       []TodoItem `json:"-"`
	NewItemText string     `form:"newItemText"`
	ItemID      int        `form:"itemId"`

	// Tracking (used in lifecycle hooks)
	LastEvent  string `json:"-"`
	EventCount int    `json:"-"`
}

// BeforeEvent - equivalent to Vue's beforeCreate/created hooks
// Called before any event handler
func (t *TodoListComponent) BeforeEvent(ctx context.Context, eventName string) error {
	slog.Info("TodoList event starting", "event", eventName)

	// You could validate authentication here
	// You could load data from database here

	return nil
}

// AfterEvent - equivalent to Vue's beforeUpdate/updated hooks
// Called after successful event handler
func (t *TodoListComponent) AfterEvent(ctx context.Context, eventName string) error {
	slog.Info("TodoList event completed", "event", eventName)

	t.LastEvent = eventName
	t.EventCount++

	// Equivalent to Vue's this.$emit()
	// You could trigger webhooks, send notifications, etc.

	return nil
}

// OnAddItem - equivalent to Vue's methods.addItem
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

	// Add the new item
	t.Items = append(t.Items, TodoItem{
		ID:        newID,
		Text:      t.NewItemText,
		Completed: false,
	})

	// Clear the input (like Vue's this.newItemText = '')
	t.NewItemText = ""

	slog.Info("Added todo item", "id", newID)
	return nil
}

// OnToggleItem - equivalent to Vue's methods.toggleItem
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

// OnDeleteItem - equivalent to Vue's methods.deleteItem
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

// OnClearCompleted - equivalent to Vue's methods.clearCompleted
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

// GetActiveCount - equivalent to Vue's computed.activeCount
func (t *TodoListComponent) GetActiveCount() int {
	count := 0
	for _, item := range t.Items {
		if !item.Completed {
			count++
		}
	}
	return count
}

// GetCompletedCount - equivalent to Vue's computed.completedCount
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
		<!-- {{ title }} becomes { data.Title } -->
		<h3>{ data.Title }</h3>

		<!-- v-model becomes name attribute + hx-include -->
		<input
			type="text"
			name="newItemText"
			value={ data.NewItemText }
			placeholder="Add item"
		/>

		<!-- @click becomes hx-post with hxc-event -->
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
			<!-- v-for becomes Go for loop -->
			for _, item := range data.Items {
				<li>
					<!-- :checked and @change become hx-post -->
					<input
						type="checkbox"
						checked?={ item.Completed }
						hx-post="/component/todolist"
						hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "toggleItem"}`, item.ID) }
						hx-target="closest .todo-list"
						hx-swap="outerHTML"
					/>

					<!-- :class becomes style attribute with conditional -->
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

		<!-- {{ activeCount }} becomes method call { fmt.Sprint(data.GetActiveCount()) } -->
		<p>
			Active: { fmt.Sprint(data.GetActiveCount()) } |
			Completed: { fmt.Sprint(data.GetCompletedCount()) }
		</p>

		<!-- v-if becomes Go if statement -->
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

## Example 3: Form with Validation (Watchers)

### Vue 2 Component

```vue
<template>
  <div class="user-form">
    <input v-model="email" placeholder="Email">
    <p v-if="emailError" class="error">{{ emailError }}</p>

    <input v-model="password" type="password" placeholder="Password">
    <p v-if="passwordError" class="error">{{ passwordError }}</p>

    <button @click="submit" :disabled="!isValid">Submit</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      email: '',
      password: '',
      emailError: '',
      passwordError: ''
    }
  },
  computed: {
    isValid() {
      return !this.emailError && !this.passwordError &&
             this.email && this.password
    }
  },
  watch: {
    email(newVal) {
      this.validateEmail(newVal)
    },
    password(newVal) {
      this.validatePassword(newVal)
    }
  },
  methods: {
    validateEmail(email) {
      if (!email) {
        this.emailError = 'Email is required'
      } else if (!email.includes('@')) {
        this.emailError = 'Invalid email'
      } else {
        this.emailError = ''
      }
    },
    validatePassword(password) {
      if (!password) {
        this.passwordError = 'Password is required'
      } else if (password.length < 8) {
        this.passwordError = 'Password must be 8+ characters'
      } else {
        this.passwordError = ''
      }
    },
    submit() {
      if (this.isValid) {
        this.$emit('submit', { email: this.email, password: this.password })
      }
    }
  }
}
</script>
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

// BeforeEvent - validate on every interaction (like Vue watchers)
func (f *UserFormComponent) BeforeEvent(ctx context.Context, eventName string) error {
	f.validateEmail()
	f.validatePassword()
	return nil
}

// OnSubmit - equivalent to Vue's methods.submit
func (f *UserFormComponent) OnSubmit() error {
	// Validation already done in BeforeEvent
	if !f.IsValid() {
		return fmt.Errorf("form is not valid")
	}

	// Process the submission
	// This is where you'd save to database, send email, etc.

	return nil
}

// OnValidate - trigger validation manually (like Vue input events)
func (f *UserFormComponent) OnValidate() error {
	// Validation happens in BeforeEvent
	// This event just triggers a re-render
	return nil
}

// validateEmail - equivalent to Vue's methods.validateEmail
func (f *UserFormComponent) validateEmail() {
	if f.Email == "" {
		f.EmailError = "Email is required"
	} else if !strings.Contains(f.Email, "@") {
		f.EmailError = "Invalid email"
	} else {
		f.EmailError = ""
	}
}

// validatePassword - equivalent to Vue's methods.validatePassword
func (f *UserFormComponent) validatePassword() {
	if f.Password == "" {
		f.PasswordError = "Password is required"
	} else if len(f.Password) < 8 {
		f.PasswordError = "Password must be 8+ characters"
	} else {
		f.PasswordError = ""
	}
}

// IsValid - equivalent to Vue's computed.isValid
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
		<!-- v-model with validation on input -->
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

		<!-- v-if for error display -->
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

		<!-- :disabled becomes disabled? -->
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

### 1. Data Binding (v-model)

**Vue:**
```vue
<input v-model="username">
```

**HxComponents:**
```templ
<input
	name="username"
	value={ data.Username }
	hx-post="/component/form"
	hx-trigger="change"
	hx-include="closest form"
	hx-target="closest form"
	hx-swap="outerHTML"
/>
```

### 2. Events (@click, @change)

**Vue:**
```vue
<button @click="handleClick">Click</button>
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

**Go method:**
```go
func (c *MyComponent) OnHandleClick() error {
	// Handle click
	return nil
}
```

### 3. Conditional Rendering (v-if, v-show)

**Vue:**
```vue
<p v-if="showMessage">{{ message }}</p>
```

**HxComponents:**
```templ
if data.ShowMessage {
	<p>{ data.Message }</p>
}
```

### 4. List Rendering (v-for)

**Vue:**
```vue
<li v-for="item in items" :key="item.id">{{ item.name }}</li>
```

**HxComponents:**
```templ
for _, item := range data.Items {
	<li>{ item.Name }</li>
}
```

### 5. Computed Properties

**Vue:**
```js
computed: {
	fullName() {
		return this.firstName + ' ' + this.lastName
	}
}
```

**HxComponents:**
```go
func (c *Component) FullName() string {
	return c.FirstName + " " + c.LastName
}
```

**Template:**
```templ
<p>{ data.FullName() }</p>
```

### 6. Lifecycle Hooks

**Vue:**
```js
created() { },
mounted() { },
updated() { },
destroyed() { }
```

**HxComponents:**
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
	// Like created/beforeUpdate
	return nil
}

func (c *Component) AfterEvent(ctx context.Context, eventName string) error {
	// Like updated
	return nil
}

func (c *Component) Process(ctx context.Context) error {
	// Final processing before render
	return nil
}
```

## State Management

In Vue, you might use Vuex for state. In HxComponents:

1. **Component State**: Store in struct fields
2. **Session State**: Use session storage or cookies
3. **Database State**: Load in BeforeEvent, save in AfterEvent
4. **Shared State**: Use a service layer or context values

## Benefits of HxComponents over Vue

1. **Type Safety**: Go's type system catches errors at compile time
2. **Performance**: Server-side rendering is faster for many use cases
3. **SEO Friendly**: All content rendered on server
4. **Simpler Deployment**: Single binary, no build step for frontend
5. **Better Security**: Business logic stays on server
6. **No JavaScript Required**: Progressive enhancement approach

## When to Use Each

**Use Vue/React/Angular when:**
- You need rich client-side interactivity (real-time games, drawing apps)
- You want offline-first capabilities
- You have complex client-side state management needs

**Use HxComponents when:**
- You want simpler architecture
- SEO is important
- You prefer type safety
- You want to avoid JavaScript complexity
- Server-side rendering fits your use case

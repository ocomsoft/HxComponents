# Migrating from Vue 3 to HxComponents

This guide demonstrates how to convert Vue 3 components (including Composition API) to the HxComponents architecture, mapping Vue 3 concepts to Go/HTMX patterns.

## Architecture Comparison

| Vue 3 Concept | HxComponents Equivalent |
|---------------|-------------------------|
| `ref()`, `reactive()` | Struct fields with `form` tags |
| `props` | Struct fields (passed via form data or initial state) |
| `methods` | Struct methods (event handlers, helpers) |
| `computed()` | Struct methods called from templates |
| `@click`, `@input` | HTMX attributes (`hx-post`, `hx-vals`, `hxc-event`) |
| `v-model` | Form inputs with `name` attribute + HTMX |
| `emit()` | Server-side event handlers (`On{EventName}`) |
| `onMounted()` | `BeforeEvent()` hook with conditional logic |
| `onUpdated()` | `AfterEvent()` hook |
| `watch()`, `watchEffect()` | Event handlers that compare values |
| `provide()`/`inject()` | Context values or service layer |

## Example 1: Simple Counter (Composition API)

### Vue 3 Component (Composition API)

```vue
<script setup>
import { ref, computed } from 'vue'

const count = ref(0)

// Computed value
const doubled = computed(() => count.value * 2)

const increment = () => count.value++
const decrement = () => count.value--
</script>

<template>
  <div class="counter">
    <button @click="decrement">−</button>
    <span>{{ count }}</span>
    <button @click="increment">+</button>
    <p>Double: {{ doubled }}</p>
  </div>
</template>
```

### Vue 3 Component (Options API)

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

// CounterComponent - equivalent to Vue's ref()/reactive() or data()
type CounterComponent struct {
	Count int `form:"count"` // ref(0) or data() { return { count: 0 } }
}

// OnIncrement - equivalent to Vue's increment method
// Called automatically when hxc-event=increment
func (c *CounterComponent) OnIncrement() error {
	c.Count++
	return nil
}

// OnDecrement - equivalent to Vue's decrement method
// Called automatically when hxc-event=decrement
func (c *CounterComponent) OnDecrement() error {
	c.Count--
	return nil
}

// Doubled - equivalent to Vue's computed property
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

## Example 2: Todo List with Lifecycle Hooks

### Vue 3 Component (Composition API)

```vue
<script setup>
import { ref, computed, onMounted, onUpdated } from 'vue'

const props = defineProps({
  title: {
    type: String,
    default: 'My Todos'
  }
})

const emit = defineEmits(['itemAdded', 'itemToggled', 'itemDeleted', 'completedCleared'])

const items = ref([])
const newItemText = ref('')

onMounted(() => {
  console.log('TodoList mounted')
  // Could load data from API here
})

onUpdated(() => {
  console.log('TodoList updated', items.value.length)
})

const activeCount = computed(() =>
  items.value.filter(i => !i.completed).length
)

const completedCount = computed(() =>
  items.value.filter(i => i.completed).length
)

const addItem = () => {
  if (!newItemText.value) return

  items.value.push({
    id: Date.now(),
    text: newItemText.value,
    completed: false
  })
  newItemText.value = ''
  emit('itemAdded')
}

const toggleItem = (id) => {
  const item = items.value.find(i => i.id === id)
  if (item) {
    item.completed = !item.completed
    emit('itemToggled', id)
  }
}

const deleteItem = (id) => {
  items.value = items.value.filter(i => i.id !== id)
  emit('itemDeleted', id)
}

const clearCompleted = () => {
  items.value = items.value.filter(i => !i.completed)
  emit('completedCleared')
}
</script>

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

// TodoListComponent - combines Vue's props, ref/reactive, and methods
type TodoListComponent struct {
	// Props (like Vue props)
	Title string `json:"-"`

	// Reactive state (like ref() or reactive())
	Items       []TodoItem `json:"-"`
	NewItemText string     `form:"newItemText"`
	ItemID      int        `form:"itemId"`

	// Tracking (used in lifecycle hooks)
	LastEvent  string `json:"-"`
	EventCount int    `json:"-"`
}

// BeforeEvent - equivalent to Vue's onMounted/onBeforeUpdate
// Called before any event handler
func (t *TodoListComponent) BeforeEvent(eventName string) error {
	slog.Info("TodoList event starting", "event", eventName)

	// This is like onMounted() or onBeforeUpdate()
	// You could validate authentication here
	// You could load data from database here

	return nil
}

// AfterEvent - equivalent to Vue's onUpdated/onUnmounted
// Called after successful event handler
func (t *TodoListComponent) AfterEvent(eventName string) error {
	slog.Info("TodoList event completed", "event", eventName, "itemCount", len(t.Items))

	t.LastEvent = eventName
	t.EventCount++

	// Equivalent to Vue's emit()
	// You could trigger webhooks, send notifications, etc.

	return nil
}

// OnAddItem - equivalent to Vue's addItem method
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

	// Add the new item (like items.value.push(...))
	t.Items = append(t.Items, TodoItem{
		ID:        newID,
		Text:      t.NewItemText,
		Completed: false,
	})

	// Clear the input (like newItemText.value = '')
	t.NewItemText = ""

	slog.Info("Added todo item", "id", newID)
	return nil
}

// OnToggleItem - equivalent to Vue's toggleItem method
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

// OnDeleteItem - equivalent to Vue's deleteItem method
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

// OnClearCompleted - equivalent to Vue's clearCompleted method
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

// GetActiveCount - equivalent to Vue's computed property
func (t *TodoListComponent) GetActiveCount() int {
	count := 0
	for _, item := range t.Items {
		if !item.Completed {
			count++
		}
	}
	return count
}

// GetCompletedCount - equivalent to Vue's computed property
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

## Example 3: Form with Watchers (Composition API)

### Vue 3 Component (Composition API)

```vue
<script setup>
import { ref, computed, watch } from 'vue'

const email = ref('')
const password = ref('')
const emailError = ref('')
const passwordError = ref('')

// Watch for changes
watch(email, (newVal) => {
  validateEmail(newVal)
})

watch(password, (newVal) => {
  validatePassword(newVal)
})

const isValid = computed(() =>
  !emailError.value && !passwordError.value &&
  email.value && password.value
)

const validateEmail = (value) => {
  if (!value) {
    emailError.value = 'Email is required'
  } else if (!value.includes('@')) {
    emailError.value = 'Invalid email'
  } else {
    emailError.value = ''
  }
}

const validatePassword = (value) => {
  if (!value) {
    passwordError.value = 'Password is required'
  } else if (value.length < 8) {
    passwordError.value = 'Password must be 8+ characters'
  } else {
    passwordError.value = ''
  }
}

const submit = () => {
  if (isValid.value) {
    console.log('Submit', { email: email.value, password: password.value })
  }
}
</script>

<template>
  <div class="user-form">
    <input v-model="email" placeholder="Email">
    <p v-if="emailError" class="error">{{ emailError }}</p>

    <input v-model="password" type="password" placeholder="Password">
    <p v-if="passwordError" class="error">{{ passwordError }}</p>

    <button @click="submit" :disabled="!isValid">Submit</button>
  </div>
</template>
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

// BeforeEvent - equivalent to Vue's watch() for validation
// This replaces watch() by running before every event
func (f *UserFormComponent) BeforeEvent(eventName string) error {
	f.validateEmail()
	f.validatePassword()
	return nil
}

// OnSubmit - equivalent to Vue's submit method
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

// validateEmail - equivalent to Vue's validateEmail function
func (f *UserFormComponent) validateEmail() {
	if f.Email == "" {
		f.EmailError = "Email is required"
	} else if !strings.Contains(f.Email, "@") {
		f.EmailError = "Invalid email"
	} else {
		f.EmailError = ""
	}
}

// validatePassword - equivalent to Vue's validatePassword function
func (f *UserFormComponent) validatePassword() {
	if f.Password == "" {
		f.PasswordError = "Password is required"
	} else if len(f.Password) < 8 {
		f.PasswordError = "Password must be 8+ characters"
	} else {
		f.PasswordError = ""
	}
}

// IsValid - equivalent to Vue's computed property
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
		<!-- v-model with validation on change -->
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

### 1. Reactivity (ref/reactive)

**Vue 3 Composition API:**
```vue
<script setup>
import { ref, reactive } from 'vue'

const count = ref(0)
const user = reactive({ name: 'John', age: 30 })

count.value++
user.name = 'Jane'
</script>
```

**HxComponents:**
```go
type Component struct {
	Count int  `form:"count"`
	Name  string `form:"name"`
	Age   int  `form:"age"`
}

func (c *Component) OnUpdate() error {
	c.Count++
	c.Name = "Jane"
	return nil
}
```

### 2. Props (defineProps)

**Vue 3:**
```vue
<script setup>
const props = defineProps({
  title: String,
  count: Number
})
</script>

<template>
  <h1>{{ props.title }}: {{ props.count }}</h1>
</template>
```

**HxComponents:**
```go
type Component struct {
	Title string `json:"-"`
	Count int    `json:"-"`
}
```

```templ
templ Component(data Component) {
	<h1>{ data.Title }: { fmt.Sprint(data.Count) }</h1>
}
```

### 3. Emits (defineEmits)

**Vue 3:**
```vue
<script setup>
const emit = defineEmits(['update', 'delete'])

const handleUpdate = () => {
  emit('update', data)
}
</script>
```

**HxComponents:**
```go
// Emits are replaced by event handlers and lifecycle hooks
func (c *Component) OnUpdate() error {
	// Handle update
	return nil
}

func (c *Component) AfterEvent(eventName string) error {
	// This is where you'd trigger side effects
	// Like calling webhooks, sending notifications, etc.
	return nil
}
```

### 4. Computed Properties

**Vue 3:**
```vue
<script setup>
import { ref, computed } from 'vue'

const firstName = ref('John')
const lastName = ref('Doe')
const fullName = computed(() => `${firstName.value} ${lastName.value}`)
</script>
```

**HxComponents:**
```go
type Component struct {
	FirstName string `form:"firstName"`
	LastName  string `form:"lastName"`
}

func (c *Component) FullName() string {
	return c.FirstName + " " + c.LastName
}
```

```templ
<p>{ data.FullName() }</p>
```

### 5. Watchers

**Vue 3:**
```vue
<script setup>
import { ref, watch } from 'vue'

const count = ref(0)

watch(count, (newVal, oldVal) => {
  console.log('Count changed from', oldVal, 'to', newVal)
})
</script>
```

**HxComponents:**
```go
type Component struct {
	Count    int `form:"count"`
	OldCount int `json:"-"`
}

func (c *Component) BeforeEvent(eventName string) error {
	c.OldCount = c.Count
	return nil
}

func (c *Component) AfterEvent(eventName string) error {
	if c.Count != c.OldCount {
		log.Printf("Count changed from %d to %d", c.OldCount, c.Count)
	}
	return nil
}
```

### 6. Lifecycle Hooks

**Vue 3:**
```vue
<script setup>
import { onMounted, onUpdated, onUnmounted } from 'vue'

onMounted(() => {
  console.log('Component mounted')
})

onUpdated(() => {
  console.log('Component updated')
})

onUnmounted(() => {
  console.log('Component unmounted')
})
</script>
```

**HxComponents:**
```go
func (c *Component) BeforeEvent(eventName string) error {
	// Like onMounted or onBeforeUpdate
	log.Println("Event starting:", eventName)
	return nil
}

func (c *Component) AfterEvent(eventName string) error {
	// Like onUpdated
	log.Println("Event completed:", eventName)
	return nil
}

// No direct equivalent to onUnmounted since components are stateless
// Use context cancellation or cleanup in AfterEvent if needed
```

### 7. Provide/Inject

**Vue 3:**
```vue
<!-- Parent -->
<script setup>
import { provide } from 'vue'

provide('user', { name: 'John' })
</script>

<!-- Child -->
<script setup>
import { inject } from 'vue'

const user = inject('user')
</script>
```

**HxComponents:**
```go
// Use context.Context for passing data
func (c *Component) BeforeEvent(eventName string) error {
	// Get user from session, database, or context
	user := getUserFromContext(c.ctx)
	c.CurrentUser = user
	return nil
}
```

### 8. Conditional Rendering

**Vue 3:**
```vue
<p v-if="show">Visible</p>
<p v-else>Hidden</p>
```

**HxComponents:**
```templ
if data.Show {
	<p>Visible</p>
} else {
	<p>Hidden</p>
}
```

### 9. List Rendering

**Vue 3:**
```vue
<li v-for="item in items" :key="item.id">{{ item.name }}</li>
```

**HxComponents:**
```templ
for _, item := range data.Items {
	<li>{ item.Name }</li>
}
```

### 10. Event Handling

**Vue 3:**
```vue
<button @click="handleClick">Click</button>
<input @input="handleInput" v-model="text">
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

<input
	name="text"
	value={ data.Text }
	hx-post="/component/mycomponent"
	hx-trigger="input"
	hx-vals='{"hxc-event": "handleInput"}'
	hx-target="closest .container"
	hx-swap="outerHTML"
/>
```

## Benefits of HxComponents over Vue 3

1. **Type Safety**: Go's type system catches errors at compile time
2. **Simpler Mental Model**: No reactivity system to understand
3. **No Build Step**: No Vite/Webpack configuration needed
4. **Performance**: Server-side rendering with minimal JavaScript
5. **SEO Friendly**: All content rendered on server
6. **Single Binary**: Easier deployment
7. **Better Security**: Business logic stays on server

## When to Use Each

**Use Vue 3 when:**
- You need rich client-side interactivity
- You want offline-first capabilities
- You need the Vue ecosystem (Vuetify, Nuxt, etc.)
- Complex client-side state management is required
- You prefer frontend-focused development

**Use HxComponents when:**
- You want simpler architecture
- SEO is important
- You prefer type safety and compile-time errors
- You want to avoid JavaScript complexity
- Server-side rendering fits your use case
- You're already working in Go
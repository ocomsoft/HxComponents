# Component Composition

This guide covers patterns for composing and combining HxComponents to build complex UIs.

## Parent-Child Relationships

Components can render other components directly in their templates.

### Basic Parent-Child

**parent.go:**
```go
package parent

import (
	"context"
	"io"
)

type ParentComponent struct {
	ParentData string
}

func (p *ParentComponent) Render(ctx context.Context, w io.Writer) error {
	return Parent(*p).Render(ctx, w)
}
```

**parent.templ:**
```templ
package parent

import "myproject/components/child"

templ Parent(data ParentComponent) {
	<div class="parent">
		<h1>Parent Component</h1>
		<p>Parent data: { data.ParentData }</p>

		<!-- Render child component directly -->
		@child.Child(child.ChildComponent{
			Message: data.ParentData,
		})
	</div>
}
```

**child.go:**
```go
package child

import (
	"context"
	"io"
)

type ChildComponent struct {
	Message string
}

func (c *ChildComponent) Render(ctx context.Context, w io.Writer) error {
	return Child(*c).Render(ctx, w)
}
```

**child.templ:**
```templ
package child

templ Child(data ChildComponent) {
	<div class="child">
		<p>Child received: { data.Message }</p>
	</div>
}
```

## Nested Interactive Components

Each component can manage its own state and handle its own events independently.

```templ
// dashboard.templ
package dashboard

import (
	"myproject/components/counter"
	"myproject/components/todolist"
	"myproject/components/chart"
)

templ Dashboard(data DashboardComponent) {
	<div class="dashboard">
		<div class="sidebar">
			<h2>Counter</h2>
			<!-- Counter handles its own increment/decrement events -->
			@counter.Counter(counter.CounterComponent{
				Count: data.CounterValue,
			})
		</div>

		<div class="main">
			<h2>Tasks</h2>
			<!-- TodoList handles its own add/delete/toggle events -->
			@todolist.TodoList(todolist.TodoListComponent{
				Title: "My Tasks",
				Items: data.Tasks,
			})
		</div>

		<div class="stats">
			<h2>Statistics</h2>
			<!-- Chart renders data visualization -->
			@chart.Chart(chart.ChartComponent{
				Data: data.ChartData,
			})
		</div>
	</div>
}
```

## Slot/Children Pattern

Pass dynamic content to child components using `templ.Component`.

```go
package card

import (
	"context"
	"io"
	"github.com/a-h/templ"
)

type CardComponent struct {
	HeaderContent templ.Component
	BodyContent   templ.Component
	FooterContent templ.Component
}

func (c *CardComponent) Render(ctx context.Context, w io.Writer) error {
	return Card(*c).Render(ctx, w)
}
```

```templ
package card

templ Card(data CardComponent) {
	<div class="card">
		if data.HeaderContent != nil {
			<div class="card-header">
				@data.HeaderContent
			</div>
		}

		<div class="card-body">
			@data.BodyContent
		</div>

		if data.FooterContent != nil {
			<div class="card-footer">
				@data.FooterContent
			</div>
		}
	</div>
}
```

### Using the Card Component

```templ
@card.Card(card.CardComponent{
	HeaderContent: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<h2>Card Title</h2>")
		return err
	}),
	BodyContent: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<p>Card body content goes here</p>")
		return err
	}),
	FooterContent: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<button>Action</button>")
		return err
	}),
})
```

### Helper for Cleaner Syntax

```go
// Helper to create templ.Component from HTML string
func HTML(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, html)
		return err
	})
}

// Usage
@card.Card(card.CardComponent{
	HeaderContent: HTML("<h2>Title</h2>"),
	BodyContent:   HTML("<p>Content</p>"),
})
```

## Shared State Between Components

### Option 1: Session Storage

```go
package shared

type SharedState struct {
	UserID    int
	Theme     string
	CartItems int
}

func GetSharedState(ctx context.Context) *SharedState {
	session := getSessionFromContext(ctx)
	return &SharedState{
		UserID:    session.GetInt("userID"),
		Theme:     session.GetString("theme"),
		CartItems: session.GetInt("cartItems"),
	}
}

func SaveSharedState(ctx context.Context, state *SharedState) error {
	session := getSessionFromContext(ctx)
	session.Set("userID", state.UserID)
	session.Set("theme", state.Theme)
	session.Set("cartItems", state.CartItems)
	return session.Save()
}
```

```go
// In component
type HeaderComponent struct {
	ctx       context.Context `json:"-"`
	UserID    int
	Theme     string
	CartItems int
}

func (c *HeaderComponent) BeforeEvent(ctx context.Context, eventName string) error {
	state := shared.GetSharedState(ctx)
	c.UserID = state.UserID
	c.Theme = state.Theme
	c.CartItems = state.CartItems
	return nil
}
```

### Option 2: Service Layer

```go
package services

type UserService struct {
	db *sql.DB
}

func (s *UserService) GetUser(id int) (*User, error) {
	// Load from database
	return user, nil
}

func (s *UserService) UpdatePreferences(id int, prefs *Preferences) error {
	// Save to database
	return nil
}

// Global service instance (or use dependency injection)
var Users = &UserService{db: db}
```

```go
// In component
type ProfileComponent struct {
	UserID int    `form:"userId"`
	User   *User  `json:"-"`
}

func (c *ProfileComponent) BeforeEvent(ctx context.Context, eventName string) error {
	user, err := services.Users.GetUserWithContext(ctx, c.UserID)
	if err != nil {
		return err
	}
	c.User = user
	return nil
}
```

### Option 3: Context Values

```go
// Middleware to inject user into context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from session/token
		user := getUserFromRequest(r)

		// Add to context
		ctx := context.WithValue(r.Context(), "user", user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// In component
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
	user := ctx.Value("user").(*User)
	c.CurrentUser = user
	return nil
}
```

## Component Communication

### Pattern 1: HTMX Events

One component triggers an event, another listens for it.

**Action Component:**
```go
func (c *ActionComponent) OnDoAction() error {
	// Perform action
	c.ActionCompleted = true

	// Trigger HTMX event
	c.GetHTMXResponseHeaders()["HX-Trigger"] = "actionCompleted"
	return nil
}
```

```templ
<button
	hx-post="/component/action"
	hx-vals='{"hxc-event": "doAction"}'
>
	Do Action
</button>
```

**Listener Component:**
```templ
<div
	hx-get="/component/listener"
	hx-trigger="actionCompleted from:body"
	hx-swap="innerHTML"
>
	Waiting for action...
</div>
```

### Pattern 2: Out-of-Band Swaps

Update multiple components from a single request.

```templ
templ ShoppingCart(data CartComponent) {
	<!-- Main cart component -->
	<div id="cart">
		<h2>Shopping Cart</h2>
		for _, item := range data.Items {
			<div>{ item.Name } - ${ fmt.Sprintf("%.2f", item.Price) }</div>
		}
		<p>Total: ${ fmt.Sprintf("%.2f", data.Total) }</p>
	</div>

	<!-- Out-of-band update for header cart count -->
	<div id="cart-count" hx-swap-oob="true">
		{ fmt.Sprint(len(data.Items)) }
	</div>

	<!-- Out-of-band update for cart total in footer -->
	<div id="cart-total-footer" hx-swap-oob="true">
		Total: ${ fmt.Sprintf("%.2f", data.Total) }
	</div>
}
```

### Pattern 3: Server-Side Events

Broadcast changes to all listening components.

```go
type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

func (e *EventBus) Publish(eventType string, data interface{}) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if subs, ok := e.subscribers[eventType]; ok {
		for _, ch := range subs {
			ch <- Event{Type: eventType, Data: data}
		}
	}
}

// Component publishes event
func (c *Component) OnUpdate() error {
	// Update data
	c.UpdateData()

	// Publish to event bus
	eventBus.Publish("dataUpdated", c.Data)

	return nil
}
```

## Layout Components

Create reusable layout components.

```go
package layouts

type PageLayoutComponent struct {
	Title   string
	Content templ.Component
	User    *User
}
```

```templ
package layouts

templ PageLayout(data PageLayoutComponent) {
	<!DOCTYPE html>
	<html>
		<head>
			<title>{ data.Title }</title>
			<script src="https://unpkg.com/htmx.org@1.9.10"></script>
			<link rel="stylesheet" href="/static/css/app.css" />
		</head>
		<body>
			<nav>
				<a href="/">Home</a>
				if data.User != nil {
					<span>Welcome, { data.User.Name }</span>
					<a href="/logout">Logout</a>
				} else {
					<a href="/login">Login</a>
				}
			</nav>

			<main>
				@data.Content
			</main>

			<footer>
				<p>&copy; 2024 My App</p>
			</footer>
		</body>
	</html>
}
```

### Using Layout

```templ
package pages

import "myproject/layouts"
import "myproject/components/dashboard"

templ DashboardPage(user *User, dashData dashboard.DashboardComponent) {
	@layouts.PageLayout(layouts.PageLayoutComponent{
		Title: "Dashboard",
		User:  user,
		Content: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return dashboard.Dashboard(dashData).Render(ctx, w)
		}),
	})
}
```

## Component Factories

Create components dynamically based on type.

```go
package components

type WidgetType string

const (
	WidgetTypeChart   WidgetType = "chart"
	WidgetTypeTable   WidgetType = "table"
	WidgetTypeCounter WidgetType = "counter"
)

type Widget struct {
	Type WidgetType
	Data interface{}
}

func CreateWidget(widget Widget) templ.Component {
	switch widget.Type {
	case WidgetTypeChart:
		data := widget.Data.(chart.ChartData)
		return chart.Chart(chart.ChartComponent{Data: data})

	case WidgetTypeTable:
		data := widget.Data.(table.TableData)
		return table.Table(table.TableComponent{Data: data})

	case WidgetTypeCounter:
		data := widget.Data.(counter.CounterData)
		return counter.Counter(counter.CounterComponent{Count: data.Count})

	default:
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return fmt.Errorf("unknown widget type: %s", widget.Type)
		})
	}
}
```

```templ
templ Dashboard(data DashboardComponent) {
	<div class="dashboard">
		for _, widget := range data.Widgets {
			<div class="widget">
				@CreateWidget(widget)
			</div>
		}
	</div>
}
```

## Higher-Order Components

Wrap components with additional functionality.

```go
// WithAuth wraps a component and checks authentication
func WithAuth(comp templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		user := ctx.Value("user").(*User)
		if user == nil {
			return HTML("<div>Please log in</div>").Render(ctx, w)
		}
		return comp.Render(ctx, w)
	})
}

// WithLoading wraps a component with loading indicator
func WithLoading(comp templ.Component, loading bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if loading {
			return HTML(`<div class="spinner">Loading...</div>`).Render(ctx, w)
		}
		return comp.Render(ctx, w)
	})
}
```

```templ
// Usage
@WithAuth(dashboard.Dashboard(data))
@WithLoading(chart.Chart(chartData), isLoading)
```

## Component Registry Pattern

Register components dynamically for plugin-like architecture.

```go
type ComponentRegistry struct {
	components map[string]func(data interface{}) templ.Component
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]func(data interface{}) templ.Component),
	}
}

func (r *ComponentRegistry) Register(name string, factory func(data interface{}) templ.Component) {
	r.components[name] = factory
}

func (r *ComponentRegistry) Get(name string, data interface{}) templ.Component {
	if factory, ok := r.components[name]; ok {
		return factory(data)
	}
	return HTML(fmt.Sprintf("<div>Component '%s' not found</div>", name))
}

// Global registry
var Components = NewComponentRegistry()

// Register components at startup
func init() {
	Components.Register("counter", func(data interface{}) templ.Component {
		return counter.Counter(data.(counter.CounterComponent))
	})

	Components.Register("chart", func(data interface{}) templ.Component {
		return chart.Chart(data.(chart.ChartComponent))
	})
}
```

```templ
// Dynamic component rendering
@Components.Get(widgetType, widgetData)
```

## Best Practices

1. **Keep Components Focused**
   - Each component should have a single responsibility
   - Split large components into smaller, reusable pieces

2. **Use Composition Over Inheritance**
   - Compose complex UIs from simple components
   - Pass components as props for flexibility

3. **Manage State at the Right Level**
   - Local state: Component fields
   - Shared state: Session or database
   - Global state: Context or service layer

4. **Document Component Interfaces**
   - Document expected props/data
   - Document events and behavior
   - Provide usage examples

5. **Test Components in Isolation**
   - Unit test component logic
   - Integration test component interactions
   - E2E test critical user flows

6. **Consider Performance**
   - Lazy load heavy components
   - Cache expensive computations
   - Use efficient data structures

7. **Handle Errors Gracefully**
   - Validate props/data
   - Return meaningful error messages
   - Provide fallback UI for errors

8. **Make Components Reusable**
   - Parameterize behavior
   - Avoid hard-coding values
   - Support customization through props

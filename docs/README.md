# HxComponents Documentation

Welcome to the HxComponents documentation! This guide will help you migrate from popular frontend frameworks or get started from scratch.

## What is HxComponents?

HxComponents is a Go-based component framework that uses HTMX for interactivity and templ for templating. It allows you to build interactive web applications with type-safe, server-side components while minimizing JavaScript.

## Quick Start

### 1. Choose Your Path

**Migrating from another framework?**
- [React Migration Guide](migration/REACT_MIGRATION.md) - Coming from React
- [Vue 3 Migration Guide](migration/VUE3_MIGRATION.md) - Coming from Vue 3
- [Vue 2 Migration Guide](migration/VUE_MIGRATION.md) - Coming from Vue 2
- [Svelte Migration Guide](migration/SVELTE_MIGRATION.md) - Coming from Svelte

**Starting fresh?**
- [Getting Started](GETTING_STARTED.md) - Setup and first component

### 2. Learn the Fundamentals

**Core Concepts:**
- [Getting Started](GETTING_STARTED.md) - Project setup, registration, and workflow
- [Component Lifecycle](#component-lifecycle) - Understanding BeforeEvent, AfterEvent, and event handlers
- [State Management](#state-management) - How to handle component state

**Building UIs:**
- [Component Composition](COMPONENT_COMPOSITION.md) - Parent-child relationships, slots, and patterns
- [Advanced Patterns](ADVANCED_PATTERNS.md) - HTMX techniques, animations, and optimizations

**Quality Assurance:**
- [Testing](TESTING.md) - Unit, integration, and E2E testing strategies
- [Common Gotchas](COMMON_GOTCHAS.md) - Troubleshooting and solutions

## Component Lifecycle

HxComponents follow a predictable lifecycle for each request:

```
Request → Parse Form Data → BeforeEvent → On{EventName} → AfterEvent → Process → Render → Response
```

### Lifecycle Hooks

**BeforeEvent(ctx context.Context, eventName string) error**
- Called before any event handler
- Use for loading data, authentication
- Return error to abort request
- Context provides request-scoped values and cancellation

**On{EventName}() error**
- Event handler (e.g., `OnSubmit`, `OnAddItem`)
- Called when `hxc-event` parameter matches
- Return error to indicate failure

**AfterEvent(ctx context.Context, eventName string) error**
- Called after successful event handler
- Use for saving data, side effects
- Return error to indicate failure
- Context provides request-scoped values and cancellation

**Process(ctx context.Context) error**
- Called after all events, before render
- Use for final transformations
- Return error to indicate failure
- Context provides request-scoped values and cancellation

## State Management

Unlike client-side frameworks where state lives in memory, HxComponents are **stateless** - each request creates a new instance. You have several options for managing state:

### Option 1: Hidden Form Fields
Best for: Simple state that doesn't need persistence
```templ
<input type="hidden" name="count" value={ fmt.Sprint(data.Count) } />
```

### Option 2: Session Storage
Best for: User-specific state during a session
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
	// Load from session
	c.Data = getFromSession("data")
	return nil
}

func (c *Component) AfterEvent(ctx context.Context, eventName string) error {
	// Save to session
	saveToSession("data", c.Data)
	return nil
}
```

### Option 3: Database
Best for: Persistent state that survives sessions
```go
func (c *Component) BeforeEvent(ctx context.Context, eventName string) error {
	// Load from database
	c.Items = db.GetItems(ctx, c.UserID)
	return nil
}

func (c *Component) AfterEvent(ctx context.Context, eventName string) error {
	// Save to database
	return db.SaveItems(ctx, c.UserID, c.Items)
}
```

## Key Concepts

### Components

Components are Go structs that:
1. Hold state in fields with `form` tags
2. Implement event handlers as methods (`On{EventName}()`)
3. Implement `Render()` method
4. Optionally implement lifecycle hooks

```go
type CounterComponent struct {
	Count int `form:"count"`
}

func (c *CounterComponent) OnIncrement() error {
	c.Count++
	return nil
}

func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {
	return Counter(*c).Render(ctx, w)
}
```

### Templates

Templates are written in templ syntax:
- Type-safe
- Compiled to Go code
- Support Go expressions and control flow

```templ
templ Counter(data CounterComponent) {
	<div>
		<button hx-post="/component/counter" hx-vals='{"count": { fmt.Sprint(data.Count) }, "hxc-event": "increment"}'>
			+
		</button>
		<span>{ fmt.Sprint(data.Count) }</span>
	</div>
}
```

### HTMX

HTMX provides interactivity:
- `hx-post`, `hx-get` - Make requests
- `hx-target` - Where to update
- `hx-swap` - How to update
- `hx-vals` - Send additional data
- `hx-trigger` - When to trigger

## Architecture Benefits

### Compared to SPA Frameworks

**Advantages:**
- ✅ Type safety at compile time
- ✅ Server-side rendering (better SEO)
- ✅ Simpler deployment (single binary)
- ✅ Less JavaScript (better performance)
- ✅ Direct database access
- ✅ No build step (for application code)
- ✅ Better security (logic stays on server)

**Trade-offs:**
- ❌ Less rich client-side interactions
- ❌ Network latency for every interaction
- ❌ No offline functionality
- ❌ Different mental model

### When to Use HxComponents

**Perfect for:**
- Content-heavy applications
- Admin dashboards
- Internal tools
- Form-heavy applications
- SEO-critical pages
- Progressive enhancement

**Not ideal for:**
- Real-time collaborative editing
- Complex animations and transitions
- Offline-first applications
- Games or interactive graphics
- Applications requiring extensive client-side logic

## Examples

All examples are in the `/examples` directory:

- **Counter** - Simple increment/decrement
- **TodoList** - Full CRUD with lifecycle hooks
- **Search** - Form submission and results
- **Login** - Authentication and redirects
- **Profile** - Complex forms with arrays

Run the examples:
```bash
cd examples
templ generate
go run main.go
```

Visit http://localhost:8080

## Common Patterns

### Form Handling
```templ
<form hx-post="/component/myform" hx-vals='{"hxc-event": "submit"}'>
	<input name="email" type="email" />
	<input name="password" type="password" />
	<button type="submit">Submit</button>
</form>
```

### Loading Indicators
```templ
<button hx-post="/component/action" hx-indicator="#spinner">
	Submit
</button>
<span id="spinner" class="htmx-indicator">Loading...</span>
```

### Optimistic Updates
```templ
<input
	type="checkbox"
	hx-post="/component/toggle"
	hx-swap="outerHTML swap:0.1s"
/>
```

### Polling
```templ
<div hx-get="/component/data" hx-trigger="every 5s">
	{ data.Value }
</div>
```

## Development Workflow

1. **Create component struct** with `form` tags
2. **Add event handlers** as `On{EventName}()` methods
3. **Create templ template** for rendering
4. **Add lifecycle hooks** if needed
5. **Register component** in main.go
6. **Run `templ generate`**
7. **Test**

```bash
# Watch for changes and regenerate
templ generate --watch

# In another terminal
go run main.go

# In another terminal (optional)
# Run tests
go test ./...
```

## Best Practices

1. **Keep components focused** - Single responsibility
2. **Use lifecycle hooks** - Load data in BeforeEvent, save in AfterEvent
3. **Test at multiple levels** - Unit, integration, E2E
4. **Handle errors gracefully** - Validate inputs, return meaningful errors
5. **Use `closest` for targeting** - Avoid ambiguous selectors
6. **Escape JSON properly** - Use fmt.Sprintf with %q or json.Marshal
7. **Include form fields** - Use hx-include when needed
8. **Name events carefully** - Must match On{EventName} method
9. **Use sessions wisely** - For temporary state
10. **Persist to database** - For permanent state

## Getting Help

- **Documentation**: You're reading it!
- **Examples**: Check `/examples` directory
- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share ideas

## Contributing

Contributions are welcome! Please:
1. Read the migration guides to understand the patterns
2. Check existing issues and PRs
3. Write tests for new features
4. Follow the existing code style
5. Update documentation

## Next Steps

1. **Start with a migration guide** or [Getting Started](GETTING_STARTED.md)
2. **Build your first component**
3. **Read [Advanced Patterns](ADVANCED_PATTERNS.md)** for HTMX techniques
4. **Study the examples** in `/examples`
5. **Set up testing** using [Testing Guide](TESTING.md)
6. **Refer to [Common Gotchas](COMMON_GOTCHAS.md)** when stuck

Happy building with HxComponents! 🚀

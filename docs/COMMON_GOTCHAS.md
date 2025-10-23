# Common Gotchas and Troubleshooting

This guide covers common issues you might encounter when working with HxComponents and how to solve them.

## 1. State Doesn't Persist Between Requests

**Problem:** Components are stateless - each request creates a new instance.

**Why:** Unlike client-side frameworks where state lives in browser memory, HxComponents run on the server and are recreated for each request.

**Solutions:**

### Option A: Hidden Form Fields
```templ
templ TodoList(data TodoListComponent) {
	<div class="todo-list">
		<!-- Preserve state with hidden fields -->
		<input type="hidden" name="itemsJson" value={ data.SerializeItems() } />

		<!-- Your component UI -->
	</div>
}
```

```go
func (t *TodoListComponent) SerializeItems() string {
	data, _ := json.Marshal(t.Items)
	return string(data)
}

func (t *TodoListComponent) BeforeEvent(eventName string) error {
	if t.ItemsJson != "" {
		json.Unmarshal([]byte(t.ItemsJson), &t.Items)
	}
	return nil
}
```

### Option B: Session Storage
```go
func (t *TodoListComponent) BeforeEvent(eventName string) error {
	// Load state from session
	session := getSessionFromContext(ctx)
	if itemsJson, ok := session["items"]; ok {
		json.Unmarshal([]byte(itemsJson), &t.Items)
	}
	return nil
}

func (t *TodoListComponent) AfterEvent(eventName string) error {
	// Save state to session
	itemsJson, _ := json.Marshal(t.Items)
	saveToSession(ctx, "items", string(itemsJson))
	return nil
}
```

### Option C: Database Persistence
```go
func (t *TodoListComponent) BeforeEvent(eventName string) error {
	// Load from database
	items, err := t.db.GetTodoItems(t.UserID)
	if err != nil {
		return err
	}
	t.Items = items
	return nil
}

func (t *TodoListComponent) AfterEvent(eventName string) error {
	// Save to database
	return t.db.SaveTodoItems(t.UserID, t.Items)
}
```

## 2. HTMX Targeting Issues

**Problem:** `hx-target` selector doesn't find the element.

**Common Causes:**
- Multiple elements with the same class
- Selector doesn't match actual DOM structure
- Element doesn't exist when HTMX tries to target it

**Solutions:**

### Bad: Multiple elements with same class
```templ
<div class="component">
	<button hx-target=".component">Update</button>
</div>
<div class="component">
	<button hx-target=".component">Update</button>
</div>
<!-- Which .component will be updated? Ambiguous! -->
```

### Good: Use `closest` to target parent
```templ
<div class="component">
	<button hx-target="closest .component">Update</button>
</div>
<div class="component">
	<button hx-target="closest .component">Update</button>
</div>
<!-- Each button targets its closest parent .component -->
```

### Best: Use unique IDs when needed
```templ
<div id="component-{ data.ID }">
	<button hx-target="#component-{ data.ID }">Update</button>
</div>
```

## 3. JSON Escaping in hx-vals

**Problem:** Special characters in JSON break the attribute.

### Bad: Quotes not escaped
```templ
<!-- BROKEN: data.Name with quotes will break JSON -->
<button hx-vals='{"name": "{ data.Name }"}'>Click</button>
```

### Good: Use fmt.Sprintf with proper escaping
```templ
<button hx-vals={ fmt.Sprintf(`{"name": %q, "id": %d}`, data.Name, data.ID) }>
	Click
</button>
```

### Better: Use helper function
```go
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
```

```templ
<button hx-vals={ toJSON(map[string]interface{}{
	"name": data.Name,
	"id": data.ID,
}) }>
	Click
</button>
```

## 4. Form Data vs Query Parameters

**Problem:** Confusion about when to use GET vs POST.

**GET:** For loading components with initial state (idempotent, cacheable)
```html
<!-- Good for links, loading data -->
<a href="/component/search?q=golang&limit=10">Search</a>
```

**POST:** For mutations and form submissions
```html
<!-- Good for changing data -->
<form hx-post="/component/search">
	<input name="q" />
	<button type="submit">Search</button>
</form>
```

## 5. Forgetting to Include Form Fields

**Problem:** `hx-vals` doesn't automatically include other form fields.

### Bad: Only sends hxc-event
```templ
<input name="email" type="email" />
<input name="password" type="password" />

<button hx-post="/component/login" hx-vals='{"hxc-event": "submit"}'>
	Login
</button>
<!-- Only hxc-event is sent, email and password are missing! -->
```

### Good: Use hx-include
```templ
<input name="email" type="email" />
<input name="password" type="password" />

<button
	hx-post="/component/login"
	hx-vals='{"hxc-event": "submit"}'
	hx-include="closest form input"
>
	Login
</button>
<!-- Now email, password, and hxc-event are all sent -->
```

### Alternative: Use a form with hx-post
```templ
<form hx-post="/component/login" hx-vals='{"hxc-event": "submit"}'>
	<input name="email" type="email" />
	<input name="password" type="password" />
	<button type="submit">Login</button>
</form>
<!-- All form fields are automatically included -->
```

## 6. Event Names Must Match Method Names

**Problem:** `hxc-event` doesn't trigger the handler.

**Rule:** The method MUST be named `On` + PascalCase(eventName)

```go
// Correct naming
func (c *Component) OnSubmit() error { }      // hxc-event: "submit"
func (c *Component) OnAddItem() error { }     // hxc-event: "addItem"
func (c *Component) OnDeleteUser() error { }  // hxc-event: "deleteUser"

// WRONG - these won't be called
func (c *Component) Submit() error { }        // Missing "On" prefix
func (c *Component) OnSubmitForm() error { }  // Event must be exactly "submitForm"
func (c *Component) Onsubmit() error { }      // Wrong case - should be OnSubmit
```

```templ
<!-- Event names must match -->
<button hx-vals='{"hxc-event": "submit"}'>Submit</button>       <!-- Calls OnSubmit -->
<button hx-vals='{"hxc-event": "addItem"}'>Add</button>        <!-- Calls OnAddItem -->
<button hx-vals='{"hxc-event": "deleteUser"}'>Delete</button>  <!-- Calls OnDeleteUser -->
```

## 7. Swap Timing and CSS Transitions

**Problem:** CSS transitions don't work because swap happens instantly.

### Without timing
```html
<!-- Swap happens immediately, no time for transition -->
<div hx-swap="outerHTML">
```

### With timing
```html
<!-- Give time for CSS transitions -->
<div hx-swap="outerHTML swap:0.5s settle:0.5s">

<style>
/* HTMX adds these classes during swap */
.htmx-swapping {
	opacity: 0;
	transition: opacity 0.5s;
}

.htmx-settling {
	opacity: 1;
	transition: opacity 0.5s;
}
</style>
```

## 8. Boosted Links Cause Full Page Reload

**Problem:** Regular links still cause full page reloads.

### Without hx-boost
```html
<!-- Full page reload -->
<a href="/page">Link</a>
```

### With hx-boost
```html
<!-- Apply to container to boost all links/forms -->
<body hx-boost="true">
	<a href="/page">This will use AJAX</a>
	<form action="/submit">This will use AJAX</form>
</body>
```

## 9. HTMX Not Loading

**Problem:** HTMX attributes don't work.

**Check:**
1. Is HTMX script loaded?
```html
<script src="https://unpkg.com/htmx.org@1.9.10"></script>
```

2. Is it loaded before your content?
```html
<!-- WRONG: HTMX loaded after content -->
<body>
	<div hx-get="/component/data">Content</div>
	<script src="https://unpkg.com/htmx.org@1.9.10"></script>
</body>

<!-- RIGHT: HTMX loaded in head -->
<head>
	<script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
	<div hx-get="/component/data">Content</div>
</body>
```

3. Check browser console for errors

## 10. Component Returns Empty Response

**Problem:** HTMX request succeeds but nothing happens.

**Common Causes:**

### Forgot to implement Render()
```go
type MyComponent struct {
	// fields...
}

// MISSING: Render() method
// Add this:
func (c *MyComponent) Render(ctx context.Context, w io.Writer) error {
	return MyComponent(*c).Render(ctx, w)
}
```

### Template name doesn't match
```go
// counter.go
type CounterComponent struct { }

func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {
	// Template name must match: Counter (exported)
	return Counter(*c).Render(ctx, w)
}
```

```templ
// counter.templ
package counter

// Template name must be exported and match
templ Counter(data CounterComponent) {
	<div>{ fmt.Sprint(data.Count) }</div>
}
```

### Forgot to run templ generate
```bash
# Run this after changing .templ files
templ generate
```

## 11. Form Tags Not Working

**Problem:** Form data isn't being parsed into struct fields.

### Check tag names
```go
type MyComponent struct {
	// Tag name must match form field name
	Email    string `form:"email"`     // Matches <input name="email">
	UserName string `form:"username"`  // Matches <input name="username">

	// WRONG - tag doesn't match
	Email string `form:"userEmail"`   // Won't match <input name="email">
}
```

### Check form field names
```templ
<form>
	<!-- name must match form tag -->
	<input name="email" type="email" />      <!-- Matches form:"email" -->
	<input name="username" type="text" />    <!-- Matches form:"username" -->
</form>
```

## 12. Context Not Available in Component

**Problem:** Need to access request context in component methods.

**Solution:** Store context in BeforeEvent

```go
type MyComponent struct {
	ctx context.Context `json:"-"`
}

func (c *MyComponent) BeforeEvent(eventName string) error {
	// Context is available here - store it if needed
	// (Note: You'll need to pass it from the registry)
	c.ctx = context.Background() // In reality, get from request
	return nil
}

func (c *MyComponent) OnSubmit() error {
	// Now you can use context
	user := getUserFromContext(c.ctx)
	return nil
}
```

## 13. CORS Issues with Components

**Problem:** CORS errors when accessing components from different origin.

**Solution:** Add CORS middleware

```go
import "github.com/go-chi/cors"

router.Use(cors.Handler(cors.Options{
	AllowedOrigins:   []string{"https://yourapp.com"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Content-Type", "HX-Request", "HX-Target"},
	ExposedHeaders:   []string{"HX-Trigger", "HX-Redirect"},
	AllowCredentials: true,
	MaxAge:           300,
}))
```

## 14. Infinite Request Loop

**Problem:** Component keeps making requests infinitely.

**Common Causes:**

### hx-trigger without proper conditions
```templ
<!-- BAD: Triggers on every change, which causes another change -->
<div
	hx-get="/component/data"
	hx-trigger="change"
	hx-target="this"
>
```

### Solution: Use specific triggers
```templ
<!-- GOOD: Only triggers once on load -->
<div
	hx-get="/component/data"
	hx-trigger="load once"
>

<!-- GOOD: Debounce user input -->
<input
	hx-get="/component/search"
	hx-trigger="keyup changed delay:500ms"
/>
```

## 15. Component Not Found Error

**Problem:** 404 error when accessing component.

**Checklist:**
1. Is component registered?
```go
components.Register[*MyComponent](registry, "mycomponent")
```

2. Does URL match registration name?
```go
// Registration: "mycomponent"
// URL must be: /component/mycomponent
```

3. Is wildcard route correct?
```go
router.Get("/component/*", registry.Handler)
router.Post("/component/*", registry.Handler)
```

4. Check route order (wildcards should be last)
```go
// WRONG: wildcard catches everything
router.Get("/component/*", registry.Handler)
router.Get("/component/special", specialHandler) // Never reached!

// RIGHT: specific routes first
router.Get("/component/special", specialHandler)
router.Get("/component/*", registry.Handler)
```

## Debugging Tips

### 1. Enable HTMX Logging
```html
<script>
	htmx.logAll();
</script>
```

### 2. Check Network Tab
- Look at request/response headers
- Check request payload
- Verify response HTML

### 3. Add Server Logging
```go
func (c *MyComponent) BeforeEvent(eventName string) error {
	log.Printf("BeforeEvent: %s, State: %+v", eventName, c)
	return nil
}
```

### 4. Use Browser DevTools
- Check console for JavaScript errors
- Inspect HTMX attributes on elements
- Monitor network requests

### 5. Test Components in Isolation
```bash
# Test component directly
curl -X POST http://localhost:8080/component/counter \
  -d "count=5&hxc-event=increment"
```

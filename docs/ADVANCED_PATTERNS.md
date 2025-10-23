# Advanced HTMX Patterns

This guide covers advanced patterns for building interactive HxComponents applications.

## Loading Indicators

Show loading state while requests are in progress:

```templ
<button
	hx-post="/component/myform"
	hx-vals='{"hxc-event": "submit"}'
	hx-indicator="#spinner"
	hx-disabled-elt="this"
>
	Submit
</button>
<span id="spinner" class="htmx-indicator">Loading...</span>

<style>
	.htmx-indicator {
		display: none;
	}
	.htmx-request .htmx-indicator {
		display: inline;
	}
	.htmx-request.htmx-indicator {
		display: inline;
	}
</style>
```

## Optimistic UI Updates

Update UI immediately, then sync with server:

```templ
templ TodoList(data TodoListComponent) {
	<div class="todo-list">
		<ul>
			for _, item := range data.Items {
				<li hx-swap="outerHTML swap:0.1s">
					<input
						type="checkbox"
						checked?={ item.Completed }
						hx-post="/component/todolist"
						hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "toggleItem"}`, item.ID) }
						hx-target="closest .todo-list"
						hx-swap="outerHTML swap:0.5s"
					/>
					<span>{ item.Text }</span>
				</li>
			}
		</ul>
	</div>
}
```

## Polling for Real-Time Updates

```templ
templ DataDisplay(data DataComponent) {
	<div
		hx-get="/component/data"
		hx-trigger="every 5s"
		hx-swap="outerHTML"
	>
		<p>Last updated: { data.UpdatedAt }</p>
		<p>Value: { data.Value }</p>
	</div>
}
```

## Server-Sent Events (SSE)

For real-time server push:

```templ
templ Notifications(data NotificationComponent) {
	<div
		hx-ext="sse"
		sse-connect="/events"
		sse-swap="notification"
		hx-swap="beforeend"
	>
		<ul id="notifications">
			for _, notif := range data.Notifications {
				<li>{ notif.Message }</li>
			}
		</ul>
	</div>
}
```

**Go SSE Handler:**
```go
func (r *Registry) HandleSSE(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-req.Context().Done():
			return
		case event := <-eventChannel:
			fmt.Fprintf(w, "event: notification\ndata: <li>%s</li>\n\n", event.Message)
			flusher.Flush()
		}
	}
}
```

## Out-of-Band Swaps

Update multiple parts of the page from a single request:

```templ
templ CartComponent(data CartComponent) {
	<div id="cart">
		<p>Items: { fmt.Sprint(len(data.Items)) }</p>
		<p>Total: ${ fmt.Sprintf("%.2f", data.Total) }</p>
	</div>

	<!-- Out-of-band update for header -->
	<div id="cart-count" hx-swap-oob="true">
		{ fmt.Sprint(len(data.Items)) }
	</div>
}
```

## Custom Request Headers

```go
type MyComponent struct {
	// Implement HTMXRequestHeaders interface
	HXRequest      bool   `json:"-"`
	HXTarget       string `json:"-"`
	HXTriggerName  string `json:"-"`
	HXCurrentURL   string `json:"-"`
}

func (c *MyComponent) SetHTMXRequestHeaders(headers map[string]string) {
	c.HXRequest = headers["HX-Request"] == "true"
	c.HXTarget = headers["HX-Target"]
	c.HXTriggerName = headers["HX-Trigger-Name"]
	c.HXCurrentURL = headers["HX-Current-URL"]
}
```

## Custom Response Headers

```go
type MyComponent struct {
	responseHeaders map[string]string
}

func (c *MyComponent) GetHTMXResponseHeaders() map[string]string {
	if c.responseHeaders == nil {
		c.responseHeaders = make(map[string]string)
	}
	return c.responseHeaders
}

func (c *MyComponent) OnSubmit() error {
	// Set response headers
	c.GetHTMXResponseHeaders()["HX-Redirect"] = "/success"
	c.GetHTMXResponseHeaders()["HX-Trigger"] = "itemUpdated"
	return nil
}
```

## CSS Transitions

```templ
templ MyComponent(data MyComponent) {
	<style>
		.fade-in {
			animation: fadeIn 0.3s ease-in;
		}

		@keyframes fadeIn {
			from { opacity: 0; }
			to { opacity: 1; }
		}

		.slide-down {
			animation: slideDown 0.3s ease-out;
		}

		@keyframes slideDown {
			from {
				opacity: 0;
				transform: translateY(-10px);
			}
			to {
				opacity: 1;
				transform: translateY(0);
			}
		}
	</style>

	if data.Visible {
		<div class="fade-in">Fading content</div>
		<div class="slide-down">Sliding content</div>
	}
}
```

## HTMX Swap Transitions

```templ
<!-- Settle timing for CSS transitions -->
<div
	hx-post="/component/mycomponent"
	hx-swap="outerHTML swap:0.5s settle:0.5s"
>
	Content
</div>

<!-- Scroll into view after swap -->
<div
	hx-post="/component/mycomponent"
	hx-swap="outerHTML scroll:top"
>
	Content
</div>

<!-- Show/focus after swap -->
<div
	hx-post="/component/mycomponent"
	hx-swap="outerHTML show:top focus-scroll:true"
>
	Content
</div>
```

## View Transitions API

Modern browsers support the View Transitions API:

```templ
<style>
	/* Define transition for specific elements */
	::view-transition-old(card),
	::view-transition-new(card) {
		animation-duration: 0.5s;
	}
</style>

<div
	style="view-transition-name: card;"
	hx-post="/component/card"
	hx-swap="outerHTML transition:true"
>
	Card content
</div>
```

## Alpine.js Integration

For complex client-side animations and interactions:

```templ
templ Dropdown(data DropdownComponent) {
	<div x-data="{ open: false }">
		<button @click="open = !open">Toggle</button>

		<div
			x-show="open"
			x-transition:enter="transition ease-out duration-300"
			x-transition:enter-start="opacity-0 transform scale-90"
			x-transition:enter-end="opacity-100 transform scale-100"
			x-transition:leave="transition ease-in duration-200"
			x-transition:leave-start="opacity-100 transform scale-100"
			x-transition:leave-end="opacity-0 transform scale-90"
		>
			Dropdown content
		</div>
	</div>
}
```

Include Alpine.js in your page:
```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
```

## Loading States with Skeleton Screens

```templ
<style>
	.skeleton {
		background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: loading 1.5s infinite;
	}

	@keyframes loading {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>

<div
	hx-get="/component/data"
	hx-trigger="load"
	hx-indicator=".skeleton"
>
	<div class="skeleton" style="height: 20px; width: 100%;"></div>
	<div class="htmx-indicator">Loading...</div>
</div>
```

## Debouncing and Throttling

```templ
<!-- Search with 500ms debounce -->
<input
	name="query"
	hx-get="/component/search"
	hx-trigger="keyup changed delay:500ms"
	hx-target="#search-results"
/>

<!-- Throttle to at most once per second -->
<input
	name="value"
	hx-post="/component/update"
	hx-trigger="keyup throttle:1s"
/>
```

## Progressive Enhancement

Start with a working HTML form, then enhance with HTMX:

```templ
<!-- Works without JavaScript -->
<form action="/component/search" method="POST">
	<input name="query" />
	<button type="submit">Search</button>
</form>

<!-- Enhanced with HTMX -->
<form
	action="/component/search"
	method="POST"
	hx-post="/component/search"
	hx-target="#results"
	hx-swap="innerHTML"
>
	<input name="query" />
	<button type="submit">Search</button>
</form>
```

## Infinite Scroll

```templ
templ ItemList(data ListComponent) {
	<div id="items">
		for _, item := range data.Items {
			<div class="item">{ item.Name }</div>
		}

		if data.HasMore {
			<!-- Load more when this element comes into view -->
			<div
				hx-get={ fmt.Sprintf("/component/items?page=%d", data.NextPage) }
				hx-trigger="revealed"
				hx-swap="outerHTML"
			>
				Loading more...
			</div>
		}
	</div>
}
```

## Confirming Actions

```templ
<button
	hx-delete="/component/item"
	hx-vals='{"itemId": 123}'
	hx-confirm="Are you sure you want to delete this item?"
>
	Delete
</button>
```

## File Uploads

```templ
<form
	hx-post="/component/upload"
	hx-encoding="multipart/form-data"
	hx-target="#upload-result"
>
	<input type="file" name="file" />
	<button type="submit">Upload</button>
</form>
```

```go
func (c *UploadComponent) OnUpload() error {
	// Access file from request in BeforeEvent
	file, header, err := req.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	// Process file
	// ...

	return nil
}
```

## WebSocket Support

```templ
<div
	hx-ext="ws"
	ws-connect="/ws"
>
	<div id="messages"></div>

	<form ws-send>
		<input name="message" />
		<button type="submit">Send</button>
	</form>
</div>
```

## History and Navigation

```templ
<!-- Push URL to history -->
<a
	href="/page/123"
	hx-get="/component/page"
	hx-vals='{"id": 123}'
	hx-push-url="true"
	hx-target="#content"
>
	View Page
</a>

<!-- Replace URL in history -->
<a
	href="/search?q=golang"
	hx-get="/component/search"
	hx-vals='{"q": "golang"}'
	hx-push-url="true"
	hx-replace-url="true"
>
	Search
</a>
```

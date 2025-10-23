package components

import "context"

// Initializer is an optional interface that components can implement to perform
// initialization logic. This is particularly useful when using constructor functions
// to create components in templ templates.
//
// The Init method is called:
// - After form decoding (in HTTP handlers)
// - Before validation
// - Before event handling
// - Before processing
//
// This allows you to:
// - Set default values
// - Load data from database
// - Initialize computed fields
// - Set up component state
//
// Example with Constructor Pattern:
//
//	// Constructor function for use in templ templates
//	func NewCard(ctx context.Context, title string, count int) *CardComponent {
//	    c := &CardComponent{
//	        Title: title,
//	        Count: count,
//	    }
//	    // Optionally call Init here for template usage
//	    c.Init(ctx)
//	    return c
//	}
//
//	// Component struct
//	type CardComponent struct {
//	    Title       string `form:"title"`
//	    Count       int    `form:"count"`
//	    Description string `json:"-"`
//	    Timestamp   string `json:"-"`
//	}
//
//	// Init sets defaults and initializes computed fields
//	func (c *CardComponent) Init(ctx context.Context) error {
//	    if c.Title == "" {
//	        c.Title = "Untitled"
//	    }
//	    c.Description = fmt.Sprintf("Card with %d items", c.Count)
//	    c.Timestamp = time.Now().Format("15:04:05")
//	    return nil
//	}
//
//	// Render implements templ.Component
//	func (c *CardComponent) Render(ctx context.Context, w io.Writer) error {
//	    return CardTemplate(*c).Render(ctx, w)
//	}
//
// Usage in templ:
//
//	templ MyPage() {
//	    <div class="container">
//	        // Constructor is called, Init() is called automatically
//	        @NewCard(ctx, "My Card", 42)
//	    </div>
//	}
//
// Usage in HTMX form:
//
//	<form hx-post="/component/card" hx-target="#result">
//	    <input name="title" value="Dynamic Card"/>
//	    <input name="count" value="10"/>
//	    <button>Submit</button>
//	</form>
//
// When the form is submitted, the handler will:
// 1. Decode form data into CardComponent
// 2. Call Init(ctx) to set defaults and computed fields
// 3. Call Validate(ctx) if implemented
// 4. Call event handlers if hxc-event is present
// 5. Call Process(ctx) if implemented
// 6. Render the component
type Initializer interface {
	Init(ctx context.Context) error
}

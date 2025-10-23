# HxComponents Patterns Guide

This guide demonstrates common patterns for properties, events, methods, and computed fields in HxComponents.

## Table of Contents

1. [Properties (Props)](#properties-props)
2. [Events](#events)
3. [Methods](#methods)
4. [Computed Fields](#computed-fields)
5. [Lifecycle Hooks](#lifecycle-hooks)
6. [Advanced Patterns](#advanced-patterns)

---

## Properties (Props)

Properties are data inputs to your component. They can come from form submissions, query parameters, or initial state.

### Basic Properties

```go
type ProductComponent struct {
	// Form properties - populated from POST/GET data
	ProductID   int    `form:"productId"`
	Quantity    int    `form:"quantity"`
	Color       string `form:"color"`

	// Internal properties - not exposed to forms
	Product     *Product `json:"-"`
	IsAvailable bool     `json:"-"`
}
```

**Key Points:**
- Use `form:"name"` tags for properties that come from HTMX requests
- Use `json:"-"` for internal state that shouldn't be serialized
- Properties are automatically populated by the form decoder

### Property Validation

```go
func (p *ProductComponent) BeforeEvent(eventName string) error {
	// Validate required properties
	if p.ProductID <= 0 {
		return fmt.Errorf("product ID is required")
	}

	if p.Quantity < 1 {
		return fmt.Errorf("quantity must be at least 1")
	}

	// Validate against business rules
	if p.Quantity > 100 {
		return fmt.Errorf("quantity cannot exceed 100")
	}

	return nil
}
```

### Default Values

```go
type SearchComponent struct {
	Query      string `form:"query"`
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	SortBy     string `form:"sortBy"`
	SortOrder  string `form:"sortOrder"`
}

func (s *SearchComponent) BeforeEvent(eventName string) error {
	// Set defaults if not provided
	if s.Page == 0 {
		s.Page = 1
	}

	if s.PageSize == 0 {
		s.PageSize = 20
	}

	if s.SortBy == "" {
		s.SortBy = "created_at"
	}

	if s.SortOrder == "" {
		s.SortOrder = "desc"
	}

	return nil
}
```

### Complex Properties (Nested Objects)

```go
type CheckoutComponent struct {
	// The form decoder can handle nested objects
	CustomerName    string `form:"customer.name"`
	CustomerEmail   string `form:"customer.email"`
	ShippingAddress string `form:"shipping.address"`
	ShippingCity    string `form:"shipping.city"`
	ShippingZip     string `form:"shipping.zip"`

	// Or use separate structs
	Items []CartItem `json:"-"`
	Total Money      `json:"-"`
}

type CartItem struct {
	ProductID int
	Quantity  int
	Price     Money
}

type Money struct {
	Amount   int64
	Currency string
}
```

---

## Events

Events are user interactions that trigger server-side handlers. Use the `hxc-event` parameter to specify which event to trigger.

### Basic Event Handlers

```go
type CounterComponent struct {
	Count int `form:"count"`
}

// Event handler methods follow the pattern: On{EventName}
// They must return error

func (c *CounterComponent) OnIncrement() error {
	c.Count++
	return nil
}

func (c *CounterComponent) OnDecrement() error {
	c.Count--
	return nil
}

func (c *CounterComponent) OnReset() error {
	c.Count = 0
	return nil
}
```

**Template Usage:**
```templ
<button
	hx-post="/component/counter"
	hx-vals={ fmt.Sprintf(`{"count": %d, "hxc-event": "increment"}`, data.Count) }
	hx-target="closest .counter"
	hx-swap="outerHTML"
>+</button>
```

### Events with Parameters

```go
type TodoComponent struct {
	Items  []TodoItem `json:"-"`
	ItemID int        `form:"itemId"`    // Parameter for events
	Text   string     `form:"text"`      // Parameter for events
}

func (t *TodoComponent) OnAddItem() error {
	if t.Text == "" {
		return fmt.Errorf("text is required")
	}

	newItem := TodoItem{
		ID:   generateID(),
		Text: t.Text,
	}

	t.Items = append(t.Items, newItem)
	t.Text = "" // Clear input after adding

	return nil
}

func (t *TodoComponent) OnDeleteItem() error {
	for i, item := range t.Items {
		if item.ID == t.ItemID {
			t.Items = append(t.Items[:i], t.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

func (t *TodoComponent) OnToggleItem() error {
	for i := range t.Items {
		if t.Items[i].ID == t.ItemID {
			t.Items[i].Completed = !t.Items[i].Completed
			return nil
		}
	}
	return fmt.Errorf("item not found")
}
```

**Template Usage:**
```templ
<!-- Add with text parameter -->
<button
	hx-post="/component/todo"
	hx-include="[name='text']"
	hx-vals='{"hxc-event": "addItem"}'
	hx-target="closest .todo"
	hx-swap="outerHTML"
>Add</button>

<!-- Delete with ID parameter -->
<button
	hx-post="/component/todo"
	hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "deleteItem"}`, item.ID) }
	hx-target="closest .todo"
	hx-swap="outerHTML"
>Delete</button>
```

### Event Lifecycle

The complete event lifecycle is:

1. **Form Decode** - Populate struct fields from request
2. **Apply Headers** - Set HTMX headers (HxRequest, HxTrigger, etc.)
3. **BeforeEvent** - Pre-event validation/setup
4. **On{EventName}** - The event handler itself
5. **AfterEvent** - Post-event cleanup/tracking
6. **Process** - Final processing before render
7. **Render** - Generate HTML response

```go
func (c *Component) BeforeEvent(eventName string) error {
	// Called before every event
	// Use for: validation, auth checks, loading data
	slog.Info("event starting", "event", eventName)
	return nil
}

func (c *Component) On{EventName}() error {
	// The specific event handler
	// Use for: business logic, state changes
	return nil
}

func (c *Component) AfterEvent(eventName string) error {
	// Called after successful event
	// Use for: logging, analytics, cache invalidation
	slog.Info("event completed", "event", eventName)
	return nil
}

func (c *Component) Process() error {
	// Called after event chain completes
	// Use for: final data loading, permission checks
	return nil
}
```

### Conditional Events

```go
type FormComponent struct {
	IsDraft bool `form:"isDraft"`
}

func (f *FormComponent) OnSave() error {
	if f.IsDraft {
		return f.saveDraft()
	}
	return f.savePublished()
}

func (f *FormComponent) saveDraft() error {
	// Save as draft
	return nil
}

func (f *FormComponent) savePublished() error {
	// Validate and save as published
	return nil
}
```

---

## Methods

Methods are helper functions that don't trigger events. They're used for internal logic, formatting, and utilities.

### Helper Methods

```go
type UserComponent struct {
	FirstName string `form:"firstName"`
	LastName  string `form:"lastName"`
	Email     string `form:"email"`
}

// Formatting helpers
func (u *UserComponent) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func (u *UserComponent) Initials() string {
	first := ""
	last := ""

	if len(u.FirstName) > 0 {
		first = string(u.FirstName[0])
	}

	if len(u.LastName) > 0 {
		last = string(u.LastName[0])
	}

	return strings.ToUpper(first + last)
}

// Validation helpers
func (u *UserComponent) IsEmailValid() bool {
	return strings.Contains(u.Email, "@") && strings.Contains(u.Email, ".")
}

func (u *UserComponent) HasCompleteName() bool {
	return u.FirstName != "" && u.LastName != ""
}
```

**Template Usage:**
```templ
<div class="user-card">
	<div class="initials">{ data.Initials() }</div>
	<div class="name">{ data.FullName() }</div>
	<div class="email">{ data.Email }</div>

	if !data.IsEmailValid() {
		<p class="error">Invalid email address</p>
	}
</div>
```

### Data Loading Methods

```go
type ProductListComponent struct {
	CategoryID int       `form:"categoryId"`
	Products   []Product `json:"-"`
	db         *sql.DB   `json:"-"`
}

func (p *ProductListComponent) Process() error {
	// Load products in Process, called before render
	return p.loadProducts()
}

func (p *ProductListComponent) loadProducts() error {
	var err error
	p.Products, err = p.db.GetProductsByCategory(p.CategoryID)
	return err
}

// Helper methods for template
func (p *ProductListComponent) HasProducts() bool {
	return len(p.Products) > 0
}

func (p *ProductListComponent) ProductCount() int {
	return len(p.Products)
}
```

### Formatting Methods

```go
type PriceComponent struct {
	Amount   int64  `form:"amount"`   // Cents
	Currency string `form:"currency"` // USD, EUR, etc.
}

func (p *PriceComponent) Formatted() string {
	dollars := float64(p.Amount) / 100.0

	switch p.Currency {
	case "USD":
		return fmt.Sprintf("$%.2f", dollars)
	case "EUR":
		return fmt.Sprintf("€%.2f", dollars)
	case "GBP":
		return fmt.Sprintf("£%.2f", dollars)
	default:
		return fmt.Sprintf("%.2f %s", dollars, p.Currency)
	}
}

func (p *PriceComponent) AmountInDollars() float64 {
	return float64(p.Amount) / 100.0
}
```

---

## Computed Fields

Computed fields are methods that calculate values based on other properties. They're called from templates as needed.

### Basic Computed Fields

```go
type OrderComponent struct {
	Items    []OrderItem `json:"-"`
	TaxRate  float64     `form:"taxRate"`
	Shipping int64       `form:"shipping"` // Cents
}

// Subtotal is computed from items
func (o *OrderComponent) Subtotal() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.Price * int64(item.Quantity)
	}
	return total
}

// Tax is computed from subtotal and rate
func (o *OrderComponent) Tax() int64 {
	return int64(float64(o.Subtotal()) * o.TaxRate)
}

// Total is computed from subtotal, tax, and shipping
func (o *OrderComponent) Total() int64 {
	return o.Subtotal() + o.Tax() + o.Shipping
}

// Formatted versions for display
func (o *OrderComponent) SubtotalFormatted() string {
	return formatCents(o.Subtotal())
}

func (o *OrderComponent) TaxFormatted() string {
	return formatCents(o.Tax())
}

func (o *OrderComponent) TotalFormatted() string {
	return formatCents(o.Total())
}

func formatCents(cents int64) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100.0)
}
```

**Template Usage:**
```templ
<div class="order-summary">
	<div class="line-item">
		<span>Subtotal:</span>
		<span>{ data.SubtotalFormatted() }</span>
	</div>
	<div class="line-item">
		<span>Tax:</span>
		<span>{ data.TaxFormatted() }</span>
	</div>
	<div class="line-item">
		<span>Shipping:</span>
		<span>{ formatCents(data.Shipping) }</span>
	</div>
	<div class="line-item total">
		<span>Total:</span>
		<span>{ data.TotalFormatted() }</span>
	</div>
</div>
```

### Cached Computed Fields

For expensive computations, you can cache results:

```go
type ReportComponent struct {
	Data      []DataPoint `json:"-"`

	// Cache fields
	avgCache  *float64    `json:"-"`
	sumCache  *int64      `json:"-"`
}

func (r *ReportComponent) Average() float64 {
	// Return cached value if available
	if r.avgCache != nil {
		return *r.avgCache
	}

	// Compute
	sum := 0.0
	for _, dp := range r.Data {
		sum += dp.Value
	}
	avg := sum / float64(len(r.Data))

	// Cache for next call
	r.avgCache = &avg

	return avg
}

func (r *ReportComponent) Sum() int64 {
	if r.sumCache != nil {
		return *r.sumCache
	}

	var sum int64
	for _, dp := range r.Data {
		sum += int64(dp.Value)
	}

	r.sumCache = &sum
	return sum
}

// Clear cache when data changes
func (r *ReportComponent) OnRefresh() error {
	r.avgCache = nil
	r.sumCache = nil

	// Reload data
	return r.loadData()
}
```

### Boolean Computed Fields

```go
type FormComponent struct {
	Email         string `form:"email"`
	Password      string `form:"password"`
	ConfirmPass   string `form:"confirmPassword"`
	AgreeToTerms  bool   `form:"agreeToTerms"`

	EmailError    string `json:"-"`
	PasswordError string `json:"-"`
}

// Validation computed fields
func (f *FormComponent) IsEmailValid() bool {
	return f.EmailError == ""
}

func (f *FormComponent) IsPasswordValid() bool {
	return f.PasswordError == ""
}

func (f *FormComponent) DoPasswordsMatch() bool {
	return f.Password == f.ConfirmPass && f.Password != ""
}

// Overall validity
func (f *FormComponent) IsValid() bool {
	return f.IsEmailValid() &&
		f.IsPasswordValid() &&
		f.DoPasswordsMatch() &&
		f.AgreeToTerms
}

// UI state computed fields
func (f *FormComponent) ShowEmailError() bool {
	return f.Email != "" && !f.IsEmailValid()
}

func (f *FormComponent) ShowPasswordError() bool {
	return f.Password != "" && !f.IsPasswordValid()
}

func (f *FormComponent) ShowPasswordMismatch() bool {
	return f.ConfirmPass != "" && !f.DoPasswordsMatch()
}
```

**Template Usage:**
```templ
<form>
	<input
		type="email"
		name="email"
		value={ data.Email }
		class={ templ.KV("error", data.ShowEmailError()) }
	/>
	if data.ShowEmailError() {
		<p class="error">{ data.EmailError }</p>
	}

	<input
		type="password"
		name="password"
		value={ data.Password }
		class={ templ.KV("error", data.ShowPasswordError()) }
	/>
	if data.ShowPasswordError() {
		<p class="error">{ data.PasswordError }</p>
	}

	<input
		type="password"
		name="confirmPassword"
		value={ data.ConfirmPass }
		class={ templ.KV("error", data.ShowPasswordMismatch()) }
	/>
	if data.ShowPasswordMismatch() {
		<p class="error">Passwords don't match</p>
	}

	<button
		type="submit"
		disabled?={ !data.IsValid() }
		hx-post="/component/form"
		hx-vals='{"hxc-event": "submit"}'
	>
		Submit
	</button>
</form>
```

### Collection Computed Fields

```go
type TaskListComponent struct {
	Tasks []Task `json:"-"`
}

// Filter computed fields
func (t *TaskListComponent) CompletedTasks() []Task {
	var completed []Task
	for _, task := range t.Tasks {
		if task.Completed {
			completed = append(completed, task)
		}
	}
	return completed
}

func (t *TaskListComponent) PendingTasks() []Task {
	var pending []Task
	for _, task := range t.Tasks {
		if !task.Completed {
			pending = append(pending, task)
		}
	}
	return pending
}

func (t *TaskListComponent) HighPriorityTasks() []Task {
	var high []Task
	for _, task := range t.Tasks {
		if task.Priority == "high" {
			high = append(high, task)
		}
	}
	return high
}

// Count computed fields
func (t *TaskListComponent) CompletedCount() int {
	return len(t.CompletedTasks())
}

func (t *TaskListComponent) PendingCount() int {
	return len(t.PendingTasks())
}

func (t *TaskListComponent) CompletionPercentage() float64 {
	if len(t.Tasks) == 0 {
		return 0
	}
	return float64(t.CompletedCount()) / float64(len(t.Tasks)) * 100
}
```

---

## Lifecycle Hooks

Lifecycle hooks let you run code at specific points in the request lifecycle.

### BeforeEvent - Validation and Setup

```go
func (c *Component) BeforeEvent(eventName string) error {
	// 1. Authentication
	if c.UserID == "" {
		return fmt.Errorf("user not authenticated")
	}

	// 2. Authorization
	if eventName == "delete" && !c.CanDelete() {
		return fmt.Errorf("user not authorized to delete")
	}

	// 3. Load required data
	if err := c.loadUser(); err != nil {
		return err
	}

	// 4. Validation
	if err := c.validate(); err != nil {
		return err
	}

	// 5. Logging
	slog.Info("event starting",
		"event", eventName,
		"user", c.UserID,
		"component", "Component")

	return nil
}
```

### AfterEvent - Cleanup and Side Effects

```go
func (c *Component) AfterEvent(eventName string) error {
	// 1. Track analytics
	c.trackEvent(eventName)

	// 2. Invalidate caches
	cache.Invalidate(c.CacheKey())

	// 3. Send notifications
	if eventName == "submit" {
		c.sendNotification()
	}

	// 4. Update timestamps
	c.LastModified = time.Now()
	c.ModifiedBy = c.UserID

	// 5. Logging
	slog.Info("event completed",
		"event", eventName,
		"duration", time.Since(c.StartTime))

	return nil
}
```

### Process - Final Setup

```go
func (c *Component) Process() error {
	// 1. Load data for rendering
	if err := c.loadData(); err != nil {
		return err
	}

	// 2. Apply permissions
	c.applyPermissions()

	// 3. Set response headers
	c.HxTrigger = "dataUpdated"

	// 4. Final validation
	return c.validateForDisplay()
}
```

---

## Advanced Patterns

### Dependency Injection

```go
type UserProfileComponent struct {
	UserID string `form:"userId"`
	User   *User  `json:"-"`

	// Injected dependencies
	userRepo UserRepository `json:"-"`
	logger   *slog.Logger   `json:"-"`
}

// Constructor function for dependency injection
func NewUserProfileComponent(userRepo UserRepository, logger *slog.Logger) *UserProfileComponent {
	return &UserProfileComponent{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *UserProfileComponent) Process() error {
	var err error
	u.User, err = u.userRepo.GetByID(u.UserID)
	if err != nil {
		u.logger.Error("failed to load user", "error", err)
		return err
	}
	return nil
}
```

### Error Handling

```go
type Component struct {
	Errors []string `json:"-"`
}

func (c *Component) AddError(msg string) {
	c.Errors = append(c.Errors, msg)
}

func (c *Component) HasErrors() bool {
	return len(c.Errors) > 0
}

func (c *Component) ClearErrors() {
	c.Errors = nil
}

// Use in event handlers
func (c *Component) OnSubmit() error {
	c.ClearErrors()

	if !c.IsValid() {
		c.AddError("Please fix validation errors")
		return nil // Don't return error, just show validation messages
	}

	if err := c.save(); err != nil {
		c.AddError("Failed to save: " + err.Error())
		return nil
	}

	return nil
}
```

**Template:**
```templ
if data.HasErrors() {
	<div class="errors">
		for _, err := range data.Errors {
			<p class="error">{ err }</p>
		}
	</div>
}
```

### State Machine Pattern

```go
type WorkflowComponent struct {
	State   string `form:"state"` // draft, pending, approved, rejected
	Comment string `form:"comment"`
}

func (w *WorkflowComponent) OnSubmit() error {
	if w.State != "draft" {
		return fmt.Errorf("can only submit from draft state")
	}
	w.State = "pending"
	return nil
}

func (w *WorkflowComponent) OnApprove() error {
	if w.State != "pending" {
		return fmt.Errorf("can only approve pending items")
	}
	w.State = "approved"
	return nil
}

func (w *WorkflowComponent) OnReject() error {
	if w.State != "pending" {
		return fmt.Errorf("can only reject pending items")
	}
	if w.Comment == "" {
		return fmt.Errorf("comment required for rejection")
	}
	w.State = "rejected"
	return nil
}

// Computed fields for UI state
func (w *WorkflowComponent) CanSubmit() bool {
	return w.State == "draft"
}

func (w *WorkflowComponent) CanApprove() bool {
	return w.State == "pending"
}

func (w *WorkflowComponent) CanReject() bool {
	return w.State == "pending"
}

func (w *WorkflowComponent) IsFinalized() bool {
	return w.State == "approved" || w.State == "rejected"
}
```

### Pagination Pattern

```go
type PaginatedListComponent struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`

	Items      []Item `json:"-"`
	TotalCount int    `json:"-"`
}

func (p *PaginatedListComponent) OnNextPage() error {
	p.Page++
	return nil
}

func (p *PaginatedListComponent) OnPrevPage() error {
	if p.Page > 1 {
		p.Page--
	}
	return nil
}

func (p *PaginatedListComponent) OnGoToPage() error {
	// Page number comes from form
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Page > p.TotalPages() {
		p.Page = p.TotalPages()
	}
	return nil
}

// Computed fields
func (p *PaginatedListComponent) TotalPages() int {
	if p.PageSize == 0 {
		return 0
	}
	return (p.TotalCount + p.PageSize - 1) / p.PageSize
}

func (p *PaginatedListComponent) HasNextPage() bool {
	return p.Page < p.TotalPages()
}

func (p *PaginatedListComponent) HasPrevPage() bool {
	return p.Page > 1
}

func (p *PaginatedListComponent) StartIndex() int {
	return (p.Page - 1) * p.PageSize
}

func (p *PaginatedListComponent) EndIndex() int {
	end := p.StartIndex() + p.PageSize
	if end > p.TotalCount {
		return p.TotalCount
	}
	return end
}
```

---

## Summary

- **Properties**: Struct fields with `form` tags for inputs, `json:"-"` for internal state
- **Events**: Methods named `On{EventName}() error` called via `hxc-event` parameter
- **Methods**: Helper functions for logic, formatting, and utilities
- **Computed Fields**: Methods that calculate values from other properties
- **Lifecycle**: BeforeEvent → On{Event} → AfterEvent → Process → Render

This architecture provides a clean, type-safe way to build interactive components with server-side rendering and HTMX.

# Practical HxComponents Examples

Real-world examples demonstrating common use cases with HxComponents.

## Table of Contents

1. [Search with Filters](#search-with-filters)
2. [Shopping Cart](#shopping-cart)
3. [Multi-Step Form](#multi-step-form)
4. [Data Table with Sorting](#data-table-with-sorting)
5. [Comment System](#comment-system)
6. [File Upload](#file-upload)
7. [Infinite Scroll](#infinite-scroll)
8. [Modal Dialog](#modal-dialog)

---

## Search with Filters

A search component with multiple filters, sorting, and real-time results.

### Component

```go
package search

import (
	"context"
	"io"
	"strings"
)

type SearchComponent struct {
	// Search parameters
	Query      string   `form:"query"`
	Category   string   `form:"category"`
	MinPrice   int      `form:"minPrice"`
	MaxPrice   int      `form:"maxPrice"`
	SortBy     string   `form:"sortBy"`
	SortOrder  string   `form:"sortOrder"`
	Page       int      `form:"page"`

	// Results
	Results    []Product `json:"-"`
	TotalCount int       `json:"-"`

	// Dependencies
	productRepo ProductRepository `json:"-"`
}

type Product struct {
	ID          int
	Name        string
	Description string
	Price       int
	Category    string
	ImageURL    string
}

func (s *SearchComponent) BeforeEvent(eventName string) error {
	// Set defaults
	if s.Page == 0 {
		s.Page = 1
	}
	if s.SortBy == "" {
		s.SortBy = "name"
	}
	if s.SortOrder == "" {
		s.SortOrder = "asc"
	}
	return nil
}

func (s *SearchComponent) OnSearch() error {
	// Search is triggered, reset to page 1
	s.Page = 1
	return nil
}

func (s *SearchComponent) OnClearFilters() error {
	s.Query = ""
	s.Category = ""
	s.MinPrice = 0
	s.MaxPrice = 0
	s.Page = 1
	return nil
}

func (s *SearchComponent) OnSort() error {
	// Sort changed, reset to page 1
	s.Page = 1
	return nil
}

func (s *SearchComponent) Process() error {
	// Load search results
	filters := ProductFilters{
		Query:     s.Query,
		Category:  s.Category,
		MinPrice:  s.MinPrice,
		MaxPrice:  s.MaxPrice,
		SortBy:    s.SortBy,
		SortOrder: s.SortOrder,
		Page:      s.Page,
		PageSize:  20,
	}

	var err error
	s.Results, s.TotalCount, err = s.productRepo.Search(filters)
	return err
}

// Computed fields
func (s *SearchComponent) HasResults() bool {
	return len(s.Results) > 0
}

func (s *SearchComponent) HasActiveFilters() bool {
	return s.Query != "" || s.Category != "" ||
	       s.MinPrice > 0 || s.MaxPrice > 0
}

func (s *SearchComponent) ResultCountText() string {
	if s.TotalCount == 0 {
		return "No results"
	}
	if s.TotalCount == 1 {
		return "1 result"
	}
	return fmt.Sprintf("%d results", s.TotalCount)
}

func (s *SearchComponent) Render(ctx context.Context, w io.Writer) error {
	return SearchView(*s).Render(ctx, w)
}
```

### Template

```templ
package search

import "fmt"

templ SearchView(data SearchComponent) {
	<div class="search-component">
		<!-- Search bar -->
		<div class="search-bar">
			<input
				type="text"
				name="query"
				value={ data.Query }
				placeholder="Search products..."
			/>
			<button
				hx-post="/component/search"
				hx-include="closest .search-component input, closest .search-component select"
				hx-vals='{"hxc-event": "search"}'
				hx-target="closest .search-component"
				hx-swap="outerHTML"
			>
				Search
			</button>
		</div>

		<!-- Filters -->
		<div class="filters">
			<select name="category">
				<option value="">All Categories</option>
				<option value="electronics" selected?={ data.Category == "electronics" }>Electronics</option>
				<option value="clothing" selected?={ data.Category == "clothing" }>Clothing</option>
				<option value="books" selected?={ data.Category == "books" }>Books</option>
			</select>

			<input
				type="number"
				name="minPrice"
				value={ fmt.Sprint(data.MinPrice) }
				placeholder="Min Price"
			/>
			<input
				type="number"
				name="maxPrice"
				value={ fmt.Sprint(data.MaxPrice) }
				placeholder="Max Price"
			/>

			if data.HasActiveFilters() {
				<button
					hx-post="/component/search"
					hx-vals='{"hxc-event": "clearFilters"}'
					hx-target="closest .search-component"
					hx-swap="outerHTML"
				>
					Clear Filters
				</button>
			}
		</div>

		<!-- Sort controls -->
		<div class="sort-controls">
			<select name="sortBy">
				<option value="name" selected?={ data.SortBy == "name" }>Name</option>
				<option value="price" selected?={ data.SortBy == "price" }>Price</option>
				<option value="created" selected?={ data.SortBy == "created" }>Newest</option>
			</select>
			<select
				name="sortOrder"
				hx-post="/component/search"
				hx-trigger="change"
				hx-include="closest .search-component input, closest .search-component select"
				hx-vals='{"hxc-event": "sort"}'
				hx-target="closest .search-component"
				hx-swap="outerHTML"
			>
				<option value="asc" selected?={ data.SortOrder == "asc" }>Ascending</option>
				<option value="desc" selected?={ data.SortOrder == "desc" }>Descending</option>
			</select>
		</div>

		<!-- Results count -->
		<div class="results-header">
			<p>{ data.ResultCountText() }</p>
		</div>

		<!-- Results -->
		<div class="results">
			if data.HasResults() {
				for _, product := range data.Results {
					<div class="product-card">
						<img src={ product.ImageURL } alt={ product.Name }/>
						<h3>{ product.Name }</h3>
						<p>{ product.Description }</p>
						<p class="price">${ fmt.Sprintf("%.2f", float64(product.Price)/100) }</p>
					</div>
				}
			} else {
				<p class="no-results">No products found. Try different filters.</p>
			}
		</div>
	</div>
}
```

---

## Shopping Cart

A shopping cart with add/remove items, quantity updates, and totals.

### Component

```go
package cart

import (
	"context"
	"fmt"
	"io"
)

type ShoppingCartComponent struct {
	Items      []CartItem `json:"-"`
	ItemID     int        `form:"itemId"`
	Quantity   int        `form:"quantity"`
	ProductID  int        `form:"productId"`

	// Promo code
	PromoCode  string     `form:"promoCode"`
	Discount   int        `json:"-"` // Percentage (0-100)

	cartService CartService `json:"-"`
}

type CartItem struct {
	ProductID   int
	ProductName string
	Price       int // Cents
	Quantity    int
	ImageURL    string
}

func (c *ShoppingCartComponent) OnAddItem() error {
	if c.ProductID == 0 {
		return fmt.Errorf("product ID required")
	}

	if c.Quantity <= 0 {
		c.Quantity = 1
	}

	// Check if item already exists
	for i := range c.Items {
		if c.Items[i].ProductID == c.ProductID {
			c.Items[i].Quantity += c.Quantity
			return nil
		}
	}

	// Add new item
	product, err := c.cartService.GetProduct(c.ProductID)
	if err != nil {
		return err
	}

	c.Items = append(c.Items, CartItem{
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       product.Price,
		Quantity:    c.Quantity,
		ImageURL:    product.ImageURL,
	})

	return nil
}

func (c *ShoppingCartComponent) OnRemoveItem() error {
	for i, item := range c.Items {
		if item.ProductID == c.ItemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

func (c *ShoppingCartComponent) OnUpdateQuantity() error {
	if c.Quantity <= 0 {
		return c.OnRemoveItem()
	}

	for i := range c.Items {
		if c.Items[i].ProductID == c.ItemID {
			c.Items[i].Quantity = c.Quantity
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

func (c *ShoppingCartComponent) OnApplyPromo() error {
	discount, err := c.cartService.ValidatePromoCode(c.PromoCode)
	if err != nil {
		return err
	}

	c.Discount = discount
	return nil
}

func (c *ShoppingCartComponent) OnClearCart() error {
	c.Items = nil
	c.PromoCode = ""
	c.Discount = 0
	return nil
}

// Computed fields
func (c *ShoppingCartComponent) IsEmpty() bool {
	return len(c.Items) == 0
}

func (c *ShoppingCartComponent) ItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

func (c *ShoppingCartComponent) Subtotal() int {
	total := 0
	for _, item := range c.Items {
		total += item.Price * item.Quantity
	}
	return total
}

func (c *ShoppingCartComponent) DiscountAmount() int {
	return c.Subtotal() * c.Discount / 100
}

func (c *ShoppingCartComponent) Total() int {
	return c.Subtotal() - c.DiscountAmount()
}

func (c *ShoppingCartComponent) SubtotalFormatted() string {
	return formatCents(c.Subtotal())
}

func (c *ShoppingCartComponent) DiscountFormatted() string {
	return formatCents(c.DiscountAmount())
}

func (c *ShoppingCartComponent) TotalFormatted() string {
	return formatCents(c.Total())
}

func formatCents(cents int) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100)
}

func (c *ShoppingCartComponent) Render(ctx context.Context, w io.Writer) error {
	return ShoppingCart(*c).Render(ctx, w)
}
```

### Template

```templ
package cart

import "fmt"

templ ShoppingCart(data ShoppingCartComponent) {
	<div class="shopping-cart">
		<h2>Shopping Cart ({ fmt.Sprint(data.ItemCount()) } items)</h2>

		if data.IsEmpty() {
			<div class="empty-cart">
				<p>Your cart is empty</p>
			</div>
		} else {
			<!-- Cart items -->
			<div class="cart-items">
				for _, item := range data.Items {
					<div class="cart-item">
						<img src={ item.ImageURL } alt={ item.ProductName }/>
						<div class="item-details">
							<h3>{ item.ProductName }</h3>
							<p>{ formatCents(item.Price) }</p>
						</div>
						<div class="item-quantity">
							<input
								type="number"
								name="quantity"
								value={ fmt.Sprint(item.Quantity) }
								min="1"
								hx-post="/component/cart"
								hx-trigger="change"
								hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "updateQuantity"}`, item.ProductID) }
								hx-include="this"
								hx-target="closest .shopping-cart"
								hx-swap="outerHTML"
							/>
						</div>
						<div class="item-total">
							{ formatCents(item.Price * item.Quantity) }
						</div>
						<button
							class="remove-btn"
							hx-post="/component/cart"
							hx-vals={ fmt.Sprintf(`{"itemId": %d, "hxc-event": "removeItem"}`, item.ProductID) }
							hx-target="closest .shopping-cart"
							hx-swap="outerHTML"
							hx-confirm="Remove this item?"
						>
							Remove
						</button>
					</div>
				}
			</div>

			<!-- Promo code -->
			<div class="promo-code">
				<input
					type="text"
					name="promoCode"
					value={ data.PromoCode }
					placeholder="Promo code"
				/>
				<button
					hx-post="/component/cart"
					hx-include="[name='promoCode']"
					hx-vals='{"hxc-event": "applyPromo"}'
					hx-target="closest .shopping-cart"
					hx-swap="outerHTML"
				>
					Apply
				</button>
			</div>

			<!-- Totals -->
			<div class="cart-totals">
				<div class="total-line">
					<span>Subtotal:</span>
					<span>{ data.SubtotalFormatted() }</span>
				</div>
				if data.Discount > 0 {
					<div class="total-line discount">
						<span>Discount ({ fmt.Sprint(data.Discount) }%):</span>
						<span>-{ data.DiscountFormatted() }</span>
					</div>
				}
				<div class="total-line grand-total">
					<span>Total:</span>
					<span>{ data.TotalFormatted() }</span>
				</div>
			</div>

			<!-- Actions -->
			<div class="cart-actions">
				<button
					class="clear-cart"
					hx-post="/component/cart"
					hx-vals='{"hxc-event": "clearCart"}'
					hx-target="closest .shopping-cart"
					hx-swap="outerHTML"
					hx-confirm="Clear entire cart?"
				>
					Clear Cart
				</button>
				<button class="checkout">
					Proceed to Checkout
				</button>
			</div>
		}
	</div>
}
```

---

## Multi-Step Form

A wizard-style form with validation at each step.

### Component

```go
package wizard

import (
	"context"
	"fmt"
	"io"
)

type RegistrationWizard struct {
	// Current step (1-3)
	Step int `form:"step"`

	// Step 1: Account
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirmPassword"`

	// Step 2: Profile
	FirstName string `form:"firstName"`
	LastName  string `form:"lastName"`
	Phone     string `form:"phone"`

	// Step 3: Preferences
	Newsletter    bool   `form:"newsletter"`
	EmailDigest   string `form:"emailDigest"` // daily, weekly, never
	Theme         string `form:"theme"`       // light, dark, auto

	// Validation errors
	Errors map[string]string `json:"-"`
}

func (w *RegistrationWizard) BeforeEvent(eventName string) error {
	// Initialize errors map
	if w.Errors == nil {
		w.Errors = make(map[string]string)
	}

	// Set defaults
	if w.Step == 0 {
		w.Step = 1
	}
	if w.EmailDigest == "" {
		w.EmailDigest = "weekly"
	}
	if w.Theme == "" {
		w.Theme = "auto"
	}

	return nil
}

func (w *RegistrationWizard) OnNext() error {
	// Validate current step
	if !w.validateStep(w.Step) {
		return fmt.Errorf("please fix validation errors")
	}

	// Move to next step
	if w.Step < 3 {
		w.Step++
	}

	return nil
}

func (w *RegistrationWizard) OnPrevious() error {
	if w.Step > 1 {
		w.Step--
	}
	return nil
}

func (w *RegistrationWizard) OnGoToStep() error {
	// Validate all previous steps before allowing jump
	for i := 1; i < w.Step; i++ {
		if !w.validateStep(i) {
			return fmt.Errorf("please complete step %d first", i)
		}
	}
	return nil
}

func (w *RegistrationWizard) OnSubmit() error {
	// Validate all steps
	allValid := true
	for i := 1; i <= 3; i++ {
		if !w.validateStep(i) {
			allValid = false
		}
	}

	if !allValid {
		return fmt.Errorf("please fix all validation errors")
	}

	// Process registration
	// Save to database, send confirmation email, etc.

	return nil
}

func (w *RegistrationWizard) validateStep(step int) bool {
	w.Errors = make(map[string]string)

	switch step {
	case 1:
		if w.Email == "" {
			w.Errors["email"] = "Email is required"
		} else if !strings.Contains(w.Email, "@") {
			w.Errors["email"] = "Invalid email"
		}

		if w.Password == "" {
			w.Errors["password"] = "Password is required"
		} else if len(w.Password) < 8 {
			w.Errors["password"] = "Password must be 8+ characters"
		}

		if w.Password != w.ConfirmPassword {
			w.Errors["confirmPassword"] = "Passwords don't match"
		}

	case 2:
		if w.FirstName == "" {
			w.Errors["firstName"] = "First name is required"
		}
		if w.LastName == "" {
			w.Errors["lastName"] = "Last name is required"
		}

	case 3:
		// All fields optional in step 3
	}

	return len(w.Errors) == 0
}

// Computed fields
func (w *RegistrationWizard) ProgressPercentage() int {
	return (w.Step * 100) / 3
}

func (w *RegistrationWizard) CanGoNext() bool {
	return w.Step < 3 && w.validateStep(w.Step)
}

func (w *RegistrationWizard) CanGoPrevious() bool {
	return w.Step > 1
}

func (w *RegistrationWizard) IsLastStep() bool {
	return w.Step == 3
}

func (w *RegistrationWizard) GetError(field string) string {
	if w.Errors == nil {
		return ""
	}
	return w.Errors[field]
}

func (w *RegistrationWizard) Render(ctx context.Context, w io.Writer) error {
	return RegistrationWizardView(*w).Render(ctx, w)
}
```

### Template

```templ
package wizard

import "fmt"

templ RegistrationWizardView(data RegistrationWizard) {
	<div class="wizard">
		<!-- Progress bar -->
		<div class="progress-bar">
			<div class="progress" style={ fmt.Sprintf("width: %d%%", data.ProgressPercentage()) }></div>
		</div>

		<!-- Step indicator -->
		<div class="steps">
			<div class={ templ.KV("step", true), templ.KV("active", data.Step == 1), templ.KV("complete", data.Step > 1) }>
				1. Account
			</div>
			<div class={ templ.KV("step", true), templ.KV("active", data.Step == 2), templ.KV("complete", data.Step > 2) }>
				2. Profile
			</div>
			<div class={ templ.KV("step", true), templ.KV("active", data.Step == 3) }>
				3. Preferences
			</div>
		</div>

		<!-- Step 1: Account -->
		if data.Step == 1 {
			<div class="step-content">
				<h2>Create Account</h2>

				<input
					type="email"
					name="email"
					value={ data.Email }
					placeholder="Email"
					class={ templ.KV("error", data.GetError("email") != "") }
				/>
				if data.GetError("email") != "" {
					<p class="error">{ data.GetError("email") }</p>
				}

				<input
					type="password"
					name="password"
					value={ data.Password }
					placeholder="Password"
					class={ templ.KV("error", data.GetError("password") != "") }
				/>
				if data.GetError("password") != "" {
					<p class="error">{ data.GetError("password") }</p>
				}

				<input
					type="password"
					name="confirmPassword"
					value={ data.ConfirmPassword }
					placeholder="Confirm Password"
					class={ templ.KV("error", data.GetError("confirmPassword") != "") }
				/>
				if data.GetError("confirmPassword") != "" {
					<p class="error">{ data.GetError("confirmPassword") }</p>
				}
			</div>
		}

		<!-- Step 2: Profile -->
		if data.Step == 2 {
			<div class="step-content">
				<h2>Your Profile</h2>

				<input
					type="text"
					name="firstName"
					value={ data.FirstName }
					placeholder="First Name"
					class={ templ.KV("error", data.GetError("firstName") != "") }
				/>
				if data.GetError("firstName") != "" {
					<p class="error">{ data.GetError("firstName") }</p>
				}

				<input
					type="text"
					name="lastName"
					value={ data.LastName }
					placeholder="Last Name"
					class={ templ.KV("error", data.GetError("lastName") != "") }
				/>
				if data.GetError("lastName") != "" {
					<p class="error">{ data.GetError("lastName") }</p>
				}

				<input
					type="tel"
					name="phone"
					value={ data.Phone }
					placeholder="Phone (optional)"
				/>
			</div>
		}

		<!-- Step 3: Preferences -->
		if data.Step == 3 {
			<div class="step-content">
				<h2>Preferences</h2>

				<label>
					<input
						type="checkbox"
						name="newsletter"
						checked?={ data.Newsletter }
					/>
					Subscribe to newsletter
				</label>

				<label>Email Digest:</label>
				<select name="emailDigest">
					<option value="daily" selected?={ data.EmailDigest == "daily" }>Daily</option>
					<option value="weekly" selected?={ data.EmailDigest == "weekly" }>Weekly</option>
					<option value="never" selected?={ data.EmailDigest == "never" }>Never</option>
				</select>

				<label>Theme:</label>
				<select name="theme">
					<option value="light" selected?={ data.Theme == "light" }>Light</option>
					<option value="dark" selected?={ data.Theme == "dark" }>Dark</option>
					<option value="auto" selected?={ data.Theme == "auto" }>Auto</option>
				</select>
			</div>
		}

		<!-- Navigation buttons -->
		<div class="wizard-actions">
			if data.CanGoPrevious() {
				<button
					hx-post="/component/wizard"
					hx-include="closest .wizard input, closest .wizard select"
					hx-vals='{"hxc-event": "previous"}'
					hx-target="closest .wizard"
					hx-swap="outerHTML"
				>
					Previous
				</button>
			}

			if !data.IsLastStep() {
				<button
					class="primary"
					hx-post="/component/wizard"
					hx-include="closest .wizard input, closest .wizard select"
					hx-vals='{"hxc-event": "next"}'
					hx-target="closest .wizard"
					hx-swap="outerHTML"
				>
					Next
				</button>
			} else {
				<button
					class="primary"
					hx-post="/component/wizard"
					hx-include="closest .wizard input, closest .wizard select"
					hx-vals='{"hxc-event": "submit"}'
					hx-target="closest .wizard"
					hx-swap="outerHTML"
				>
					Complete Registration
				</button>
			}
		</div>
	</div>
}
```

---

## Summary

These examples demonstrate:

1. **Search with Filters**: Complex state management with multiple parameters
2. **Shopping Cart**: Item management with computed totals
3. **Multi-Step Form**: Wizard pattern with validation at each step

Key patterns used:
- Form parameter binding
- Event-driven state changes
- Computed fields for display
- Validation in BeforeEvent
- Conditional rendering based on state
- HTMX for reactive updates

Each example shows how to structure components for real-world applications using the event-driven pattern with properties, events, methods, and computed fields.

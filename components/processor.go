package components

import "context"

// Processor is an optional interface that components can implement to perform
// processing logic before rendering. This is useful for validation, business logic,
// data transformation, database queries, or setting response headers based on processing results.
//
// The Process method is called after form data is decoded and request headers are applied,
// but before response headers are applied and the component is rendered.
//
// The context parameter provides request-scoped values and cancellation signals,
// which is useful for database queries, API calls, and other operations that may
// need to be cancelled or have timeouts.
//
// Example:
//
//	type LoginForm struct {
//	    Username   string `form:"username"`
//	    Password   string `form:"password"`
//	    RedirectTo string `json:"-"`
//	    Error      string `json:"-"`
//	}
//
//	func (f *LoginForm) Process(ctx context.Context) error {
//	    // Can use ctx for database queries, timeouts, etc.
//	    user, err := db.FindUser(ctx, f.Username)
//	    if err != nil {
//	        return fmt.Errorf("database error: %w", err)
//	    }
//
//	    if user != nil && user.CheckPassword(f.Password) {
//	        f.RedirectTo = "/dashboard"
//	        return nil
//	    }
//	    f.Error = "Invalid credentials"
//	    return nil
//	}
//
// Process should return an error only for unexpected failures. Validation errors
// or business logic errors should be stored in the struct fields and rendered
// in the template.
type Processor interface {
	Process(ctx context.Context) error
}

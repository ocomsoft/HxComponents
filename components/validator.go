package components

import "context"

// Validator is an optional interface that components can implement to perform
// validation after form decoding but before processing.
//
// Example:
//
//	type LoginForm struct {
//	    Username string `form:"username"`
//	    Password string `form:"password"`
//	    Errors   []ValidationError `json:"-"`
//	}
//
//	func (f *LoginForm) Validate(ctx context.Context) []ValidationError {
//	    var errs []ValidationError
//	    if f.Username == "" {
//	        errs = append(errs, ValidationError{Field: "username", Message: "Username is required"})
//	    }
//	    if len(f.Password) < 8 {
//	        errs = append(errs, ValidationError{Field: "password", Message: "Password must be at least 8 characters"})
//	    }
//	    f.Errors = errs
//	    return errs
//	}
type Validator interface {
	Validate(ctx context.Context) []ValidationError
}

// ValidationError represents a single validation error for a field.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError.
func (v ValidationError) Error() string {
	return v.Field + ": " + v.Message
}

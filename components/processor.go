package components

// Processor is an optional interface that components can implement to perform
// processing logic before rendering. This is useful for validation, business logic,
// data transformation, or setting response headers based on processing results.
//
// The Process method is called after form data is decoded and request headers are applied,
// but before response headers are applied and the component is rendered.
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
//	func (f *LoginForm) Process() error {
//	    if f.Username == "demo" && f.Password == "password" {
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
	Process() error
}

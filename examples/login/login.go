package login

// LoginForm represents the data for a login component.
type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`

	// Response fields
	RedirectTo string `json:"-"`
	Refresh    bool   `json:"-"`
	Error      string `json:"-"`
}

// Implement response header interfaces

func (f *LoginForm) GetHxRedirect() string {
	return f.RedirectTo
}

func (f *LoginForm) GetHxRefresh() bool {
	return f.Refresh
}

// Process implements the Processor interface to handle login logic.
// This is called automatically by the registry after form decoding
// and before rendering the component.
func (f *LoginForm) Process() error {
	// Simple validation for demo purposes
	if f.Username == "" || f.Password == "" {
		f.Error = "Username and password are required"
		return nil
	}

	// Simulate successful login
	if f.Username == "demo" && f.Password == "password" {
		f.RedirectTo = "/dashboard"
		return nil
	}

	// Invalid credentials
	f.Error = "Invalid credentials"
	return nil
}

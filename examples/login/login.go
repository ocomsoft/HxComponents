package login

// LoginComponent represents the data for a login component.
type LoginComponent struct {
	Username string `form:"username"`
	Password string `form:"password"`

	// Response fields
	RedirectTo string `json:"-"`
	Refresh    bool   `json:"-"`
	Error      string `json:"-"`
}

// Implement response header interfaces

func (c *LoginComponent) GetHxRedirect() string {
	return c.RedirectTo
}

func (c *LoginComponent) GetHxRefresh() bool {
	return c.Refresh
}

// Process implements the Processor interface to handle login logic.
// This is called automatically by the registry after form decoding
// and before rendering the component.
func (c *LoginComponent) Process() error {
	// Simple validation for demo purposes
	if c.Username == "" || c.Password == "" {
		c.Error = "Username and password are required"
		return nil
	}

	// Simulate successful login
	if c.Username == "demo" && c.Password == "password" {
		c.RedirectTo = "/dashboard"
		return nil
	}

	// Invalid credentials
	c.Error = "Invalid credentials"
	return nil
}

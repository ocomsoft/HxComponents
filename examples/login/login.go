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

// ProcessLogin simulates login logic. In a real application, this would
// query a database and validate credentials.
func (f *LoginForm) ProcessLogin() error {
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

package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ocomsoft/HxComponents/components"
	"github.com/ocomsoft/HxComponents/examples/login"
	"github.com/ocomsoft/HxComponents/examples/profile"
	"github.com/ocomsoft/HxComponents/examples/search"
)

func main() {
	// Create the component registry
	registry := components.NewRegistry()

	// Register components
	components.Register(registry, "search", search.SearchComponent)
	components.Register(registry, "login", func(data login.LoginForm) templ.Component {
		// Process login before rendering
		data.ProcessLogin()
		return login.LoginComponent(data)
	})
	components.Register(registry, "profile", func(data profile.UserProfile) templ.Component {
		// Process profile update before rendering
		data.ProcessUpdate()
		return profile.ProfileComponent(data)
	})

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount the component registry
	registry.Mount(router)

	// Serve static files and demo page
	router.Get("/", serveHomePage)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func serveHomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTMX Component Registry Demo</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 8px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #007bff;
            padding-bottom: 10px;
        }
        h2 {
            color: #555;
            margin-top: 0;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #555;
        }
        input[type="text"],
        input[type="email"],
        input[type="password"],
        input[type="number"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
            font-size: 14px;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
        }
        button:hover {
            background-color: #0056b3;
        }
        .result {
            margin-top: 20px;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 4px;
            min-height: 50px;
        }
        .alert {
            padding: 12px 20px;
            border-radius: 4px;
            margin-bottom: 15px;
        }
        .alert-success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .alert-danger {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .alert-info {
            background-color: #d1ecf1;
            border: 1px solid #bee5eb;
            color: #0c5460;
        }
        .alert-warning {
            background-color: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
        }
        .info {
            font-size: 14px;
            color: #666;
            font-style: italic;
        }
        code {
            background-color: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
    </style>
</head>
<body>
    <h1>HTMX Generic Component Registry - Demo</h1>

    <div class="container">
        <p>This page demonstrates the HTMX Generic Component Registry pattern. Each form below uses the same registry infrastructure with type-safe component rendering.</p>
        <p><strong>Try the examples below:</strong></p>
        <ul>
            <li><strong>Search:</strong> Demonstrates request header capture (HX-Boosted, HX-Request, etc.)</li>
            <li><strong>Login:</strong> Demonstrates response headers (HX-Redirect) - Use username: <code>demo</code>, password: <code>password</code></li>
            <li><strong>Profile:</strong> Demonstrates complex form data with arrays</li>
            <li><strong>GET Support:</strong> Components can be loaded via GET requests with query parameters - <a href="#" hx-get="/component/search?q=golang&limit=5" hx-target="#get-demo" style="color: #007bff; text-decoration: underline;">Click here to demo</a></li>
        </ul>
        <div id="get-demo" class="result" style="margin-top: 15px;"></div>
    </div>

    <div class="grid">
        <!-- Search Example -->
        <div class="container">
            <h2>Search Component</h2>
            <form hx-post="/component/search" hx-target="#search-result" hx-boost="true">
                <div class="form-group">
                    <label for="search-query">Search Query:</label>
                    <input type="text" id="search-query" name="q" placeholder="Enter search term" value="htmx components">
                </div>
                <div class="form-group">
                    <label for="search-limit">Result Limit:</label>
                    <input type="number" id="search-limit" name="limit" value="10" min="1" max="100">
                </div>
                <button type="submit">Search</button>
            </form>
            <div id="search-result" class="result"></div>
        </div>

        <!-- Login Example -->
        <div class="container">
            <h2>Login Component</h2>
            <form hx-post="/component/login" hx-target="#login-result">
                <div class="form-group">
                    <label for="username">Username:</label>
                    <input type="text" id="username" name="username" placeholder="demo">
                </div>
                <div class="form-group">
                    <label for="password">Password:</label>
                    <input type="password" id="password" name="password" placeholder="password">
                </div>
                <button type="submit">Login</button>
            </form>
            <div id="login-result" class="result"></div>
            <p class="info">Hint: Use username "demo" and password "password" for successful login.</p>
        </div>

        <!-- Profile Example -->
        <div class="container">
            <h2>Profile Update Component</h2>
            <form hx-post="/component/profile" hx-target="#profile-result">
                <div class="form-group">
                    <label for="name">Name:</label>
                    <input type="text" id="name" name="name" placeholder="John Doe">
                </div>
                <div class="form-group">
                    <label for="email">Email:</label>
                    <input type="email" id="email" name="email" placeholder="john@example.com">
                </div>
                <div class="form-group">
                    <label for="tags">Tags (comma-separated):</label>
                    <input type="text" id="tags" name="tags" placeholder="developer, golang, htmx">
                </div>
                <button type="submit">Update Profile</button>
            </form>
            <div id="profile-result" class="result"></div>
        </div>
    </div>

    <div class="container" style="margin-top: 30px;">
        <h2>How It Works</h2>
        <p>Each component above is registered in the Go component registry with type-safe form binding and HTMX header handling:</p>
        <ol>
            <li>Forms submit to <code>/component/{component_name}</code> via POST or GET</li>
            <li>The registry parses form data (POST) or query parameters (GET) into typed structs</li>
            <li>HTMX request headers are automatically captured (if component implements interfaces)</li>
            <li>Component logic executes (validation, business logic, etc.)</li>
            <li>HTMX response headers are set (redirects, triggers, etc.)</li>
            <li>Template is rendered and returned to the target element</li>
        </ol>
        <p><strong>GET vs POST:</strong></p>
        <ul>
            <li><strong>POST</strong> - Standard HTMX pattern for form submissions with form data in request body</li>
            <li><strong>GET</strong> - Useful for loading components with initial state via query parameters (e.g., <code>/component/search?q=golang&limit=5</code>)</li>
        </ul>
    </div>
</body>
</html>`

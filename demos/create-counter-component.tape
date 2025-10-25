# VHS Script: Creating a Counter Component with HxComponents
# This script demonstrates how to create a simple counter component
# Run with: vhs create-counter-component.tape

Output create-counter-component.gif

Set FontSize 16
Set Width 1400
Set Height 900
Set TypingSpeed 50ms
Set PlaybackSpeed 1.0

Hide
Type "mkdir -p counter-demo && cd counter-demo"
Enter
Sleep 500ms
Type "clear"
Enter
Show

Type "# Let's create a Counter component using HxComponents!"
Sleep 2s
Enter
Enter

Type "# Step 1: Create the Go file for our component"
Sleep 2s
Enter
Type "cat > counter.go << 'EOF'"
Enter

Type "package counter"
Enter
Enter

Type "import ("
Enter
Type "    \"context\""
Enter
Type "    \"io\""
Enter
Type ")"
Enter
Enter

Type "// CounterComponent represents a simple counter that can increment/decrement."
Sleep 1s
Enter
Type "// Key concepts:"
Sleep 1s
Enter
Type "//   1. Use struct fields with 'form' tags to capture form data"
Sleep 1.5s
Enter
Type "//   2. Event handlers are methods prefixed with 'On' + event name"
Sleep 1.5s
Enter
Type "//   3. No Action field needed - events are handled via hxc-event parameter"
Sleep 1.5s
Enter
Type "type CounterComponent struct {"
Enter
Type "    Count int \`form:\"count\"\`  // 'form' tag binds to form/JSON data"
Sleep 1s
Enter
Type "}"
Enter
Enter
Sleep 1s

Type "// OnIncrement is an event handler method."
Sleep 1s
Enter
Type "// Convention: On + EventName (e.g., hxc-event=increment -> OnIncrement)"
Sleep 1.5s
Enter
Type "// The framework automatically calls this when hxc-event=increment is received"
Sleep 1.5s
Enter
Type "func (c *CounterComponent) OnIncrement(ctx context.Context) error {"
Enter
Type "    c.Count++  // Modify component state"
Sleep 1s
Enter
Type "    return nil // Return nil for success, or error if something fails"
Sleep 1s
Enter
Type "}"
Enter
Enter
Sleep 1s

Type "// OnDecrement handles the decrement event"
Sleep 1s
Enter
Type "func (c *CounterComponent) OnDecrement(ctx context.Context) error {"
Enter
Type "    c.Count--"
Enter
Type "    return nil"
Enter
Type "}"
Enter
Enter
Sleep 1s

Type "// Render implements templ.Component interface."
Sleep 1s
Enter
Type "// This allows the component to be used in templ templates"
Sleep 1.5s
Enter
Type "// and as a standalone HTMX component"
Sleep 1.5s
Enter
Type "func (c *CounterComponent) Render(ctx context.Context, w io.Writer) error {"
Enter
Type "    return Counter(*c).Render(ctx, w)"
Sleep 1s
Enter
Type "}"
Enter

Type "EOF"
Enter
Sleep 2s
Enter

Type "# Step 2: Create the templ template file"
Sleep 2s
Enter
Type "cat > counter.templ << 'EOF'"
Enter

Type "package counter"
Enter
Enter
Type "import \"fmt\""
Enter
Enter

Type "// Counter template renders the UI for our component"
Sleep 1s
Enter
Type "// Key HTMX attributes explained:"
Sleep 1s
Enter
Type "templ Counter(data CounterComponent) {"
Enter
Type "    <div class=\"counter-component\">"
Enter
Sleep 1s

Type "        <!-- Decrement button -->"
Enter
Type "        <button"
Enter
Type "            hx-post=\"/component/counter\""
Sleep 1s
Enter
Type "            // hx-post: Send POST request to server"
Sleep 1.5s
Enter
Enter

Type "            hx-vals={ fmt.Sprintf(\`{\"count\": %d, \"hxc-event\": \"decrement\"}\`, data.Count) }"
Sleep 1s
Enter
Type "            // hx-vals: Send current count + event name"
Sleep 1.5s
Enter
Type "            // 'hxc-event': 'decrement' triggers OnDecrement method"
Sleep 1.5s
Enter
Enter

Type "            hx-target=\"closest .counter-component\""
Sleep 1s
Enter
Type "            // hx-target: Replace this div with server response"
Sleep 1.5s
Enter
Enter

Type "            hx-swap=\"outerHTML\""
Sleep 1s
Enter
Type "            // hx-swap: Replace entire element (not just contents)"
Sleep 1.5s
Enter
Type "        >"
Enter
Type "            −"
Enter
Type "        </button>"
Enter
Sleep 1s
Enter

Type "        <!-- Display current count -->"
Enter
Type "        <span>{ fmt.Sprint(data.Count) }</span>"
Enter
Sleep 1s
Enter

Type "        <!-- Increment button (same pattern as decrement) -->"
Enter
Type "        <button"
Enter
Type "            hx-post=\"/component/counter\""
Enter
Type "            hx-vals={ fmt.Sprintf(\`{\"count\": %d, \"hxc-event\": \"increment\"}\`, data.Count) }"
Enter
Type "            // 'hxc-event': 'increment' triggers OnIncrement method"
Sleep 1s
Enter
Type "            hx-target=\"closest .counter-component\""
Enter
Type "            hx-swap=\"outerHTML\""
Enter
Type "        >"
Enter
Type "            +"
Enter
Type "        </button>"
Enter
Type "    </div>"
Enter
Type "}"
Enter

Type "EOF"
Enter
Sleep 2s
Enter

Type "# Step 3: Understanding the component lifecycle"
Sleep 2s
Enter
Type "cat << 'EXPLANATION'"
Enter
Sleep 500ms

Type "Component Request Flow:"
Enter
Type "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Sleep 1s
Enter

Type "1. User clicks + button"
Sleep 1s
Enter

Type "2. HTMX sends POST to /component/counter with:"
Sleep 1s
Enter
Type "   {\"count\": 5, \"hxc-event\": \"increment\"}"
Sleep 1.5s
Enter

Type "3. HxComponents framework:"
Sleep 1s
Enter
Type "   a) Parses form data into CounterComponent struct"
Sleep 1s
Enter
Type "   b) Sees hxc-event=increment"
Sleep 1s
Enter
Type "   c) Calls OnIncrement() method automatically"
Sleep 1.5s
Enter

Type "4. OnIncrement modifies Count: 5 -> 6"
Sleep 1s
Enter

Type "5. Framework calls Render() with updated state"
Sleep 1s
Enter

Type "6. Returns HTML with Count=6 to browser"
Sleep 1s
Enter

Type "7. HTMX swaps the old component with new HTML"
Sleep 1.5s
Enter
Sleep 1s

Enter
Type "Key Benefits:"
Enter
Type "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
Sleep 1s
Enter

Type "✓ No JavaScript required - pure Go + HTMX"
Sleep 1s
Enter

Type "✓ Type-safe component state with Go structs"
Sleep 1s
Enter

Type "✓ Convention-based event handling (On + EventName)"
Sleep 1s
Enter

Type "✓ Automatic state binding with form tags"
Sleep 1s
Enter

Type "✓ Composable with other templ components"
Sleep 1.5s
Enter

Type "EXPLANATION"
Enter
Sleep 3s

Type "# That's it! You've created a complete interactive component!"
Sleep 3s
Enter
Sleep 2s

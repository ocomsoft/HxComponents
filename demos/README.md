# HxComponents Demos

This directory contains VHS (Video Handout System) scripts for creating animated terminal demonstrations of HxComponents features.

## Prerequisites

Install VHS:
```bash
# macOS
brew install vhs

# Linux
go install github.com/charmbracelet/vhs@latest

# Or download from: https://github.com/charmbracelet/vhs
```

## Available Demos

### Create Counter Component (`create-counter-component.tape`)

An interactive tutorial showing how to create a Counter component from scratch, with detailed explanations of:

- Component struct definition with form tags
- Event handler methods (OnIncrement, OnDecrement)
- Templ template creation with HTMX attributes
- Complete request/response flow explanation
- Key benefits of the HxComponents pattern

**Generate the GIF:**
```bash
cd demos
vhs create-counter-component.tape
```

This will create `create-counter-component.gif` which can be embedded in documentation or shared as a tutorial.

## Creating Your Own Demos

VHS scripts use a simple tape format. See the [VHS documentation](https://github.com/charmbracelet/vhs) for more details on available commands.

Basic structure:
```tape
Output my-demo.gif
Set FontSize 16
Set Width 1400
Set Height 900

Type "echo 'Hello World'"
Enter
Sleep 1s
```

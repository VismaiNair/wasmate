# wasmate - Quickstart Guide

A simple CLI tool that streamlines Go WebAssembly development by handling the build process, runtime setup, and local testing.
For thorough installation and uninstallation directions, please take a look at [the install directions](INSTALL.md).

## 5-Minute Setup

### Prerequisites
- Go 1.16+ installed
- Basic familiarity with Go development

### Installation

```bash
go install github.com/vismainair/wasmate@latest
```

## Your First WASM App

Let's create a DOM-manipulating WASM application in 5 steps:

### Step 1: Create Your Project
```bash
mkdir hello-wasm
cd hello-wasm
```

### Step 2: Initialize Go Module
```bash
go mod init wasm-test
go mod tidy
```

### Step 3: Create Your Go Application

Create `hello_wasm.go`:
```go
package main

import (
	"syscall/js"
)

func main() {
	// Get the document object
	document := js.Global().Get("document")
	element := document.Call("getElementById", "hello-world")

	// Set the Inner HTML
	element.Set("innerHTML", "Hello from WASM!")
	
	select {} // Keep the program running (required for WASM)
}
```

### Step 4: Create Your HTML File

Create `index.html`:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Go WASM DOM Example</title>
</head>
<body>
    <h1>Go WASM Demo</h1>
    <p>The Go WASM module will add content below:</p>
    <p id="hello-world"></p>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("hello_wasm.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });
    </script>
</body>
</html>
```

### Step 5: Build and Run
```bash
# Copy the JavaScript runtime
wasmate js

# Build your Go code to WebAssembly
wasmate build

# Start the development server
wasmate run
```

**That's it!** Open your browser to `http://localhost:8080` and see your Go code running in the browser!
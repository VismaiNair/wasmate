# `wasmate`
A CLI tool to simplify Go WebAssembly development.

Are you tired of remembering long Go commands and complex flags just to build a simple WebAssembly project? wasmate is a CLI wrapper that handles all the hassle for you. It simplifies the entire workflow, letting you build, run, and manage your Go Wasm applications with just a few intuitive commands.

## Why Use `wasmate`?
- __No More Flags__: Stop memorizing `GOOS=js GOARCH=wasm go build`. Just run `wasmate build`.

- __Simple Workflow__: Go from Go source code to a running web server in three easy steps.

- __Built for Developers__: Focus on writing your code, not on wrestling with the build system.

## Getting Started
To get up and running, check out our [Quickstart Guide](QUICKSTART.md) for installation and your first project. For detailed installation and uninstallation instructions, take a look at the [Install Guide](INSTALL.md).

## Commands
### `wasmate build`
Compiles your Go source code into a WebAssembly (.wasm) binary.

__Examples:__

- Compile all Go files in the current directory: 
```
wasmate build
```

- Compile a specific file or directory:
```
wasmate build main.go
```
```
wasmate build ./path/to/my/project
```

- Compile to a specific output:
```
wasmate build --output main.wasm
```
```
wasmate build -o main.wasm
```
```
wasmate build ./path/to/my/project -o main.wasm
```

You may add whichever flags you would like to use and use them within the same command, as long as you separate them with spaces.

### `wasmate js`
Copies the required wasm_exec.js file from your Go installation. This file is essential for running WebAssembly in a web browser.

__Example:__

- Copy the file to your current directory:

```
wasmate js
```

### `wasmate run`
Serves your HTML and WebAssembly files over a local web server, making it easy to test your application.

__Examples:__

- Start the server on the default port (8080):
```
wasmate run
```

- Specify a different port:

```
wasmate run --port 8000
```
```
wasmate run -p 8000
```

- Open localhost link in your browser"

```
wasmate run --open
```

```
wasmate run -o
```


You may add whichever flags you would like to use and use them within the same command, as long as you separate them with spaces.
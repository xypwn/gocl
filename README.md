# GoCL: OpenCL bindings for Go
Dependency-free (except CGo) automatic bindings for modern OpenCL (**WIP!!!**).

[![Go Reference](https://pkg.go.dev/badge/github.com/xypwn/gocl.svg)](https://pkg.go.dev/github.com/xypwn/gocl)

### Dependency-free
- This library has the official OpenCL headers and ICD loader built into its source code, meaning no binary dependencies.

### API Design
The primary goal is to create OpenCL bindings that adhere to idiomatic Go API design as much as possible without sacrificing functionality or speed.

> [!WARNING]
> The API is currently unstable, meaning it may undergo changes in future versions!

- Every Go function is named the same as its OpenCL counterpart, but without the `cl` prefix.
- Every OpenCL function that can error returns an `error`, whose underlying type is **always** `*cl.Error` if it is **not** `nil` (meaning it can safely be cast).
- Full support for enum types (including flag types) with `.String()` method.
- Some functions are added to simplify certain common operations (e.g. `CreateBufferSlice`).
- If a function returns its error through a pointer in OpenCL, that error is a return value in the Go function.

### Examples
See the [examples](examples) directory.
- [helloworld.go](examples/helloworld/helloworld.go): Square an input array on the GPU and return it.

### License
All files from the OpenCL Headers and OpenCL ICD Loader are licensed under the [Apache-2.0 License](https://www.apache.org/licenses/LICENSE-2.0) (each file also explicitly states such).

All other files are licensed under the [MIT License](https://opensource.org/license/mit).

### Disclaimer
This project is unofficial and bears no connection to OpenCL, Khronos Group or any associated entitires.
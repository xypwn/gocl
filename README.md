# GoCL: OpenCL bindings for Go
Dependency-free (except CGo) automatic bindings for modern OpenCL (**WIP!!!**).

[![Go Reference](https://pkg.go.dev/badge/github.com/xypwn/gocl.svg)](https://pkg.go.dev/github.com/xypwn/gocl)

### Dependency-free
- This library has the official OpenCL headers and ICD loader built into its source code, meaning no binary dependencies.

### API Design
- **The API is currently unstable, meaning it may undergo changes in future versions!**
- Every Go function is named the same as its OpenCL counterpart, but without the `cl` prefix.
- Every OpenCL function that can error returns an `error`, whose underlying type is **always** `*cl.Error` if it is **not** `nil` (meaning it can safely be cast).
- If a function returns its error through a pointer in OpenCL, that error is returned as another parameter in the Go function.

### License
All files from the OpenCL Headers and OpenCL ICD Loader are licensed under the [Apache-2.0 License](https://www.apache.org/licenses/LICENSE-2.0) (each file also explicitly states such).

All other files are licensed under the [MIT License](https://opensource.org/license/mit).

### Disclaimer
This project is unofficial and bears no connection to OpenCL, Khronos Group or any associated entitires.
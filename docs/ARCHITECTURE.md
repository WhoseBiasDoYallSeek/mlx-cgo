# mlx-cgo — Architecture

## The Idea

Apple built something remarkable. MLX — a machine learning framework designed from the ground up for Apple Silicon — is fast, elegant, and native. They wrapped it in C. But to talk to it from Go, you had to go through Python. A Python script. At build time. With an external dependency.

mlx-cgo eliminates that entirely.

This is a fork of [mlx-c](https://github.com/ml-explore/mlx-c) that rewrites the Python code generation toolchain in Go and adds a first-class CGo bindings generator. One binary. Zero external dependencies. Built for Apple Silicon.

---

## What mlx-cgo Does

Three things, precisely:

1. **Generates the same C bindings** that the original Python toolchain produced — fully compatible with upstream mlx-c
2. **Adds a Go/CGo bindings generator** — idiomatic Go wrappers over those C bindings (the new differentiator)
3. **Eliminates Python from the build process** entirely — no `pip install`, no interpreter, no `cxxheaderparser`

---

## The Generator: `mlxgen`

Everything flows through a single binary built with `go build ./cmd/mlxgen`. No runtime. No interpreter. One file on disk.

```
cmd/mlxgen/main.go           ← entry point: --header / --type / --cgo-from
       │
       ├── internal/parser/
       │       ├── clang.go        ← runs clang -Xclang -ast-dump=json
       │       ├── ast.go          ← Go structs for the JSON AST nodes
       │       └── extract.go      ← filters mlx:: namespace, extracts functions/enums
       │
       ├── internal/types/
       │       ├── ctypes.go       ← C++ → C type registry
       │       └── gotypes.go      ← C → Go/CGo type registry
       │
       ├── internal/variants/
       │       └── variants.go     ← overload resolution tables
       │
       ├── internal/hooks/
       │       ├── hooks.go        ← hardcoded overrides for complex types
       │       └── hooks_test.go   ← 23 unit tests
       │
       ├── internal/cheader/
       │       └── parse.go        ← parses generated .h files → []CFunc
       │
       └── internal/codegen/
               ├── cnamespace.go   ← ToSnakeLetters, CNamespace, FuncCName
               ├── c_header.go     ← generates .h
               ├── c_impl.go       ← generates .cpp
               ├── vector.go       ← mlx/c/vector.h + .cpp
               ├── map.go          ← mlx/c/map.h + .cpp
               ├── closure.go      ← mlx/c/closure.h + .cpp
               ├── private.go      ← mlx/c/private/*.h
               ├── regression_test.go
               └── cgo/
                   ├── cgo.go      ← GenerateGoFile() — idiomatic CGo wrappers
                   └── scaffold.go ← GenerateGoBase() — opaque structs + helpers
```

**Total**: ~5,800 lines of Go. 21 source files. Zero external dependencies.

---

## Repository Layout

```
mlx-cgo/
├── cmd/
│   └── mlxgen/
│       └── main.go                  ← CLI entry point
│
├── internal/
│   ├── model/
│   │   └── types.go                 ← domain structs: Function, Param, Enum
│   ├── parser/
│   │   ├── ast.go                   ← clang JSON AST node structs
│   │   ├── clang.go                 ← executes clang, parses JSON → *Node
│   │   └── extract.go               ← filters mlx::, extracts []Function, []Enum
│   ├── types/
│   │   ├── ctypes.go                ← C++ → C type registry (543 lines)
│   │   └── gotypes.go               ← C → Go/CGo type registry
│   ├── variants/
│   │   └── variants.go              ← overload tables + Resolve()
│   ├── hooks/
│   │   ├── hooks.go                 ← hardcoded C/C++ overrides (825 lines)
│   │   └── hooks_test.go            ← 23 unit tests
│   ├── cheader/
│   │   └── parse.go                 ← parses generated .h → []CFunc
│   └── codegen/
│       ├── cnamespace.go
│       ├── c_header.go
│       ├── c_impl.go
│       ├── vector.go
│       ├── map.go
│       ├── closure.go
│       ├── private.go
│       ├── regression_test.go       ← 10 golden-file regression tests
│       └── cgo/
│           ├── cgo.go               ← GenerateGoFile()
│           └── scaffold.go          ← GenerateGoBase()
│
├── mlx/c/                           ← 29 .h + 26 .cpp — Apple's C bindings
├── scripts/
│   └── regenerate.sh                ← regenerates all static files in one command
├── go.mod
└── docs/
```

---

## How Data Flows

```
                    ┌──────────────────────────────────────────┐
                    │          cmd/mlxgen/main.go               │
                    │  --header     →  C generator mode         │
                    │  --cgo-from   →  CGo generator mode       │
                    └──────────────────┬───────────────────────┘
                                       │
               ┌───────────────────────┴─────────────────────────┐
               │  C generator mode                                │  CGo generator mode
               ▼                                                  ▼
┌──────────────────────────┐                     ┌───────────────────────────┐
│    internal/parser/       │                     │   internal/cheader/        │
│  clang -ast-dump=json     │                     │  parse.go → []CFunc       │
│  JSON → []Function, Enum  │                     └───────────────┬───────────┘
└─────────────┬────────────┘                                     │
              │                                                   ▼
              ▼                                   ┌───────────────────────────┐
┌──────────────────────────┐                     │ internal/codegen/cgo/     │
│   internal/variants/      │                     │ GenerateGoFile()          │
│  resolves overloads        │                     │ → idiomatic Go wrappers   │
└─────────────┬────────────┘                     └───────────────────────────┘
              │
              ▼
┌──────────────────────────┐
│    internal/types/        │
│  C++ type → C type info   │
└─────────────┬────────────┘
              │
       ┌──────┴────────────────────────┐
       │                               │
       ▼                               ▼
┌────────────────────┐     ┌────────────────────────┐
│  codegen/c_header  │     │  codegen/c_impl         │
│  → .h              │     │  → .cpp                 │
└────────────────────┘     └────────────────────────┘

Static generators — no C++ headers required:
  codegen/vector.go  → mlx/c/vector.h + vector.cpp + private/vector.h
  codegen/map.go     → mlx/c/map.h    + map.cpp    + private/map.h
  codegen/closure.go → mlx/c/closure.h + closure.cpp + private/closure.h
  codegen/private.go → mlx/c/private/*.h
```

---

## The Key Insight: Use Clang

We don't write a C++ parser. Clang is already mandatory to compile mlx-c — it's already there. And it has a superpower most people forget about:

```bash
clang -Xclang -ast-dump=json -fsyntax-only -x c++-header -std=c++20 header.h
```

This produces a complete, structured JSON representation of every function, type, and namespace in the header. The compiler parses the code — perfectly — because it's the same parser that will eventually compile it.

> **Apple Clang note**: The correct flag is `-Xclang -ast-dump=json`. Not `-ast-dump-format=json` — that flag does not exist in Apple Clang 21+.

The JSON AST is parsed into Go structs (`internal/parser/ast.go`), then `extract.go` walks the tree, filters to the `mlx::` namespace, and extracts `[]Function` and `[]Enum` for the code generators.

---

## The Domain Model

```go
// Function represents a C++ function extracted from a header.
type Function struct {
    Name       string
    Namespace  string   // e.g. "mlx::core::fft"
    ReturnType string   // clang qualType string
    Params     []Param
}

type Param struct {
    Name       string
    Type       string   // clang qualType string
    HasDefault bool
}

type Enum struct {
    Name      string
    Namespace string
    Values    []string
}
```

---

## The Type System (`internal/types/`)

Every MLX type is described by a `TypeInfo` — a struct of code-generation functions. One lookup, everything needed to emit correct C code.

```go
type TypeInfo struct {
    CName    string
    CPPName  string
    AltNames []string

    CArg           func(name string) string  // "const mlx_array a"
    CReturnArg     func(name string) string  // "mlx_array* res"
    CToCPP         func(name string) string  // "mlx_array_get_(a)"
    CAssignFromCPP func(dest, src string) string
    CNew           func(name string) string  // "auto res = mlx_array_new_()"
    Free           func(name string) string  // "mlx_array_free(a)"
}
```

Supported type categories:

- **Opaque handles**: `mlx_array`, `mlx_stream`, `mlx_device`, all closure, vector, and map variants
- **Primitives**: `int`, `float`, `double`, `bool`, `uint32_t`, `int64_t`, `size_t`
- **Raw vectors**: `std::vector<int>` → `(const int* ptr, size_t num)` — one C++ param becomes two C params
- **Pairs**: `std::pair<A,B>` → `A* res_0, B* res_1`
- **Optionals**: `std::optional<T>` → `mlx_optional_*`
- **Closures**: 6 distinct closure types with full signatures
- **Strings**: `std::string` ↔ `const char*`

---

## The Hooks System (`internal/hooks/`)

Some MLX functions use types the parser cannot map from clang's AST — closures with complex template arguments, optional std::function parameters, module-level constructs. These are handled by `hooks.go`: a registry of handwritten C/C++ snippets that override or supplement the automatic generator.

```go
var registry = map[string]hookFn{
    "mlx_fast_metal_kernel": hookFastMetalKernel,
    "mlx_load_gguf":         hookLoadGGUF,
    "mlx_save_gguf":         hookSaveGGUF,
    "mlx_export_to_dot":     hookExportToDot,
    "mlx_custom_function":   hookCustomFunction,
    "mlx_custom_vjp":        hookCustomVJP,
}

// Module hooks emit code at the module level, independent of parsed functions.
// Used for headers like export.h where all types are too complex to parse.
var moduleRegistry = map[string]func(bool) string{
    "export": hookModuleExport,
}
```

`Apply(name, impl)` is called by both `c_header.go` and `c_impl.go` for every function before attempting normal generation. If a hook handles it, the hook's output is used directly. `ApplyModule(headername, impl)` is called once per module to emit module-level declarations.

---

## The CLI

```
# Dynamic generators — require C++ headers:
mlxgen --header <path.h> [--implementation] [--headername name] [--docstring "..."]

# Static generators — no headers required:
mlxgen --type=vector   [--implementation]
mlxgen --type=map      [--implementation]
mlxgen --type=closure  [--implementation]
mlxgen --type=private  --ctype <mlx_X[;mlx_Y]> --cpptype <cpp_X[;cpp_Y]> \
                       [--no-copy] [--using <alias[;alias]>] [--mlx-include <path>]

# CGo generator:
mlxgen --cgo-from <mlx/c/foo.h> [--package <pkgname>]

# Go scaffolding:
mlxgen --type=go-base [--package <pkgname>]
```

---

## The CGo Generator (`internal/codegen/cgo/`)

This is the new differentiator. The Go community needs idiomatic bindings to MLX — not raw C calls with unsafe pointers. The CGo generator reads the generated `.h` files and produces real Go functions with proper types, error returns, and slice arguments.

### From C to Go — automatically

```go
// C signature:
// int mlx_fft_fft(mlx_array a, int n, int axis, mlx_stream s, mlx_array* res)

// Generated Go:
func FftFft(a Array, n int32, axis int32, s Stream) (Array, error) {
    var res C.mlx_array
    if ret := C.mlx_fft_fft(a.c, C.int(n), C.int(axis), s.c, &res); ret != 0 {
        return Array{}, fmt.Errorf("mlx_fft_fft failed (code %d)", ret)
    }
    return Array{c: res}, nil
}
```

### The scaffold (`scaffold.go`)

Generates the base Go package — the foundation everything else builds on:

- CGo preamble: `#include "mlx/c/mlx.h"`
- Opaque structs: `type Array struct { c C.mlx_array }`
- `Free()` method for every handle
- `DefaultStream()`, `DefaultDevice()`
- `Dtype` constants: Bool, Float32, Float16, BFloat16, etc.
- Error helper: `mlxError(name string, code C.int) error`

```bash
mlxgen --type=go-base --package mlx > mlx/go/mlx.go
```

### C → Go type mappings

| C Type                 | Go Type          | Notes |
|------------------------|------------------|-------|
| `mlx_array`            | `Array`          | Opaque handle |
| `mlx_stream`           | `Stream`         | Opaque handle |
| `mlx_device`           | `Device`         | Opaque handle |
| `mlx_dtype`            | `Dtype`          | Alias for `C.mlx_dtype` |
| `int`                  | `int32`          | |
| `float`                | `float32`        | |
| `double`               | `float64`        | |
| `bool`                 | `bool`           | |
| `size_t`               | `uint`           | |
| `const int*, size_t`   | `[]int32`        | Nil-safe guard on empty slice |
| `const float*, size_t` | `[]float32`      | Nil-safe guard on empty slice |
| `const char*`          | `string`         | `C.CString` + `defer C.free` |
| `void*`                | `unsafe.Pointer` | |

### Empty slice safety

`(*C.int)(unsafe.Pointer(&slice[0]))` panics on an empty slice. The generator always emits a nil-safe guard:

```go
var _ptr_axes *C.int
if len(axes) > 0 {
    _ptr_axes = (*C.int)(unsafe.Pointer(&axes[0]))
}
C.mlx_fft_fftn(&_res, a.c, _ptr_axes, C.size_t(len(axes)), ...)
```

### C string management

`C.CString()` allocates on the C heap with `malloc`. The generator always emits a `defer C.free`:

```go
_cstr_file := C.CString(file)
defer C.free(unsafe.Pointer(_cstr_file))
C.mlx_load(&_res, _cstr_file, s.c)
```

### Raw-vector pair detection

`internal/cheader/parse.go` detects consecutive `(const T*, size_t)` pairs in C function signatures and marks them `IsRawVecPtr` / `IsRawVecLen`. The CGo generator collapses each pair into a single Go `[]T` slice.

---

## What Gets Generated

### Input

```cpp
// mlx/core/ops.h
array abs(const array& a, StreamOrDevice s = {});
```

### C header — `mlx/c/ops.h`

```c
int mlx_abs(mlx_array* res, mlx_array a, mlx_stream s);
```

### C implementation — `mlx/c/ops.cpp`

```cpp
extern "C" int mlx_abs(mlx_array* res, mlx_array a, mlx_stream s) {
  try {
    mlx_array_set_(*res, mlx::core::abs(mlx_array_get_(a), mlx_stream_get_(s)));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
```

### Go/CGo wrapper — new

```go
func Abs(a Array, s Stream) (Array, error) {
    var res C.mlx_array
    if C.mlx_abs(&res, a.c, s.c) != 0 {
        return Array{}, fmt.Errorf("mlx_abs failed")
    }
    return Array{c: res}, nil
}
```

---

## Memory Model: Go as Gatekeeper

`mlx::core::array` is internally a `std::shared_ptr<ArrayDesc>`. C++ manages tensor lifetime via reference counting. `mlx_array_free` decrements the refcount — the actual GPU/Metal buffer survives as long as the computation graph needs it.

Go does **not** use `runtime.SetFinalizer`. It acts as a high-performance gatekeeper: dispatching C calls without GC scan overhead.

```go
// Generated by scaffold.go
type Array struct { c C.mlx_array }

// Memory is owned by the caller. Call Free when done.
// Idiomatic usage: defer result.Free()
func (v Array) Free() { C.mlx_array_free(v.c) }
```

### Idiomatic usage

```go
s := mlx.DefaultStream()
defer s.Free()

result, err := mlx.Abs(a, s)
if err != nil { return err }
defer result.Free()  // frees the 8-byte wrapper; tensor data stays alive in C++
```

### The one risk: forgetting `.Free()`

Each C handle is an 8-byte `{ void* ctx }` struct allocated with `new` in C++. Forgetting `.Free()` leaks those 8 bytes. **Tensor data does not leak** — `shared_ptr<ArrayDesc>` handles that automatically.

| Scenario | Leak per op | Impact |
|---|---|---|
| Inference script (runs and exits) | 8 bytes × N | ✅ Irrelevant — process exits |
| Long-running inference server | 8 bytes × N ops/req | ⚠️ Accumulates over time |
| Training loop (millions of steps) | 8 bytes × steps × ops | 🔴 Can be significant |

```go
// ❌ Leaks 8 bytes per call
for i := 0; i < 1_000_000; i++ {
    result, _ := mlx.Abs(input, stream)
    // forgot result.Free()
}

// ✅ Correct — explicit Free inside the loop
for i := 0; i < 1_000_000; i++ {
    result, _ := mlx.Abs(input, stream)
    result.Free()  // don't use defer here — defer fires at function exit, not loop iteration
}
```

Detection:

```bash
# macOS
MALLOC_STACK_LOGGING=1 leaks --atExit -- ./my-binary

# Linux
valgrind --leak-check=full ./my-binary
```

---

## `scripts/regenerate.sh`

Regenerates all static output files in one command:

```bash
# Static files only (vector, map, closure, private headers):
./scripts/regenerate.sh

# Static + dynamic C bindings from MLX C++ headers:
./scripts/regenerate.sh --mlx-source <build_dir>/_deps/mlx-src/mlx
```

Builds `mlxgen`, generates each file, runs `clang-format --style=file`, writes directly to `mlx/c/`.

> **macOS**: `/bin/bash` on macOS is bash 3.2 (pre-GPLv3). The script uses indexed arrays with `IFS='|' read -r` — no `declare -A`, no bash 4 requirement.

---

## CMake Integration

```cmake
option(MLX_C_REGENERATE_BINDINGS "Regenerate C bindings using mlxgen (requires Go)" OFF)
```

When `ON`:

| Target | What it does |
|--------|-------------|
| `mlxgen_build` | Compiles `mlxgen` via `go build` |
| `mlxgen_static` | Runs `regenerate.sh` → static files |
| `mlxgen_dynamic` | Runs `regenerate.sh --mlx-source` → `ops.h`, `fft.h`, etc. |
| `mlxgen_all` | Both static + dynamic |

`add_dependencies(mlxc mlxgen_all)` ensures generation always precedes compilation.

Auto-detected: Go via `find_program(GO_EXECUTABLE go)`, clang-format via `xcrun --find clang-format` on macOS.

CMake delegates all generation logic to `regenerate.sh` instead of inline `foreach` — avoids CMake treating `;` in C++ type names like `mlx::core::array` as list separators.

```bash
# Normal build — uses committed files, no regeneration:
cmake -B build -DMLX_C_BUILD_EXAMPLES=OFF
cmake --build build

# Build with regeneration:
cmake -B build -DMLX_C_REGENERATE_BINDINGS=ON
cmake --build build --target mlxgen_static
cmake --build build
```

---

## Test Suite

### Regression tests

10 tests compare generator output byte-for-byte against the committed golden files in `mlx/c/`. Whitespace is normalized by running `clang-format --style=file` on both sides before diffing.

| Test | Golden File | Status |
|------|-------------|--------|
| `TestRegressionVectorHeader`   | `mlx/c/vector.h`          | ✅ PASS |
| `TestRegressionVectorImpl`     | `mlx/c/vector.cpp`        | ✅ PASS |
| `TestRegressionMapHeader`      | `mlx/c/map.h`             | ✅ PASS |
| `TestRegressionMapImpl`        | `mlx/c/map.cpp`           | ✅ PASS |
| `TestRegressionClosureHeader`  | `mlx/c/closure.h`         | ✅ PASS |
| `TestRegressionClosureImpl`    | `mlx/c/closure.cpp`       | ✅ PASS |
| `TestRegressionPrivateArray`   | `mlx/c/private/array.h`   | ✅ PASS |
| `TestRegressionPrivateStream`  | `mlx/c/private/stream.h`  | ✅ PASS |
| `TestRegressionPrivateDevice`  | `mlx/c/private/device.h`  | ✅ PASS |
| `TestRegressionPrivateClosure` | `mlx/c/private/closure.h` | ✅ PASS |

### Unit tests

#### `internal/parser/parser_test.go` (10 tests)
- `TestExtract_BasicFunction` — basic function extraction with parameters
- `TestExtract_Overloads` — multiple overloads of the same function
- `TestExtract_StdinFilter` — filters functions from transitive includes via `<stdin>`
- `TestNormalizeType_BuiltinTypes` — normalizes primitive types (int, float, bool)
- `TestNormalizeType_Qualifiers` — removes `const`, `&`, `&&` recursively
- `TestNormalizeFunctionType_Closures` — normalizes `std::function<T(Args)>` with closure types
- `TestReturnTypeFromQual` — extracts return type from clang qualType
- `TestJoinNS` — namespace concatenation
- `TestExtract_Enum` — extraction of enums and their values
- `TestExtract_DefaultParameters` — marks parameters with default values

#### `internal/types/types_registry_test.go` (10 tests)
- `TestFindByCPP_SimpleTypes`, `_VectorTypes`, `_OptionalTypes`, `_PairTypes`
- `TestFindByC_Lookup` — reverse lookup
- `TestAltNames_ShortFormLookup`
- `TestTypeInfo_CArg`, `_CReturnArg`, `_CToCPP`, `_Free`

#### `internal/variants/variants_test.go` (11 tests)
- Resolve (suffix application, drops, mismatches, unknown functions)
- ResolveDetail (allowlist, rejection)
- NSKey conversion, empty input handling, table validation

#### `internal/hooks/hooks_test.go` (23 tests)
- `TestApply_UnknownFunction_NotHandled`
- `TestApply_KnownFunction_IsHandled` — all 6 registered hooks
- `TestApply_AllHooksReturnNonEmptyCode` — all hooks × header + impl
- `TestHookLoadGGUF_Header` / `_Impl`
- `TestHookSaveGGUF_Header` / `_Impl`
- `TestHookExportToDot_Header` / `_Impl`
- `TestHookCustomFunction_Header` / `_Impl`
- `TestHookCustomVJP_Header` / `_Impl`
- `TestHookFastMetalKernel_Header` — asserts no `_custom_kernel_config` in output
- `TestHookFastMetalKernel_Impl` — calls `mlx::core::fast::metal_kernel`
- `TestApplyModule_UnknownModule_ReturnsEmpty`
- `TestApplyModule_Export_Header` / `_Impl`
- `TestApplyModule_Export_HeaderHasNoExternC`
- `TestAllHooks_HeaderHasNoExternC` — contract: headers never emit `extern "C"`
- `TestAllHooks_ImplHasExternC` — contract: impls always emit `extern "C"`
- `TestRegistry_HasExpectedEntries`
- `TestModuleRegistry_HasExportEntry`

#### `internal/codegen/cgo/edge_cases_test.go` (7 tests + 2 benchmarks)
- `TestEmptySlicePanic` — nil-safe slice guard
- `TestCStringDefer` — `C.CString` + `defer C.free` pattern
- `TestNilPointerCheck` — nil pointer handling from C
- `TestLargeArrayAllocation` — 1KB / 1MB / 10MB allocation safety
- `TestConcurrentMemoryAccess` — 100 goroutines × 100 iterations, race detector
- `TestGoToCGo_StandalonePointerParams` — `goToCGo` uses `_ptr_NAME` for standalone pointer params
- `BenchmarkEmptySliceGuard`, `BenchmarkNilCheck`

### Summary

```bash
go test ./...

# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/codegen      (10 regression)
# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/codegen/cgo  (7 unit + 2 bench)
# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/hooks        (23 unit)
# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/parser       (10 unit)
# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/types        (10 unit)
# ok  github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/variants     (11 unit)
#
# Total: 71 tests, 100% passing, race detector clean
```

---

## Design Decisions

| Decision | Choice | Why |
|---|---|---|
| C++ parsing | `clang -Xclang -ast-dump=json` | Clang is mandatory; no new dependencies |
| Language | Go | Single binary, zero runtime, zero interpreter |
| Output compatibility | Identical to upstream mlx-c | Validates correctness; enables incremental migration |
| External libraries | None | `go build` — no network, no vendor, no `go.sum` |
| Overload resolution | Table in `variants.go` | Proven logic from Python original |
| CGo generator reads `.h` | Generated `.h`, not raw C++ | Separation of concerns; `.h` is the stable interface |
| Platform | Apple Silicon (macOS/arm64) | MLX is Apple's framework; Metal is the target |
| No `SetFinalizer` | Explicit `.Free()` | High-performance gatekeeper; no GC overhead |

---

## Coverage Analysis

| Package | Lines | Risk | Tests | Coverage |
|---|---|---|---|---|
| `internal/parser/` | ~196 | High — errors cascade | 10 unit | ~85% lines |
| `internal/types/` | ~573 | High — all generators depend on it | 10 unit | ~80% lines |
| `internal/variants/` | ~196 | Medium | 11 unit | ~90% lines |
| `internal/hooks/` | ~825 | Medium — manual code, changes with MLX | 23 unit | ~75% lines |
| `internal/codegen/` | ~1,500+ | Medium — regression-tested | 10 regression | ~60% lines |
| `internal/cheader/` | ~217 | Low | None (indirect) | ~40% |

### Open gaps

- `cheader/parse.go` — no dedicated unit tests; covered indirectly via `--cgo-from`
- Full `--cgo-from` pipeline has no end-to-end integration test

---

## Validated Headers

Validated against `mlx/c/` golden files using MLX v0.31.2:

| Header | Functions | Status |
|---|---|---|
| `ops.h` + `einsum.h` | 239 | ✅ |
| `fft.h` | 16 | ✅ |
| `linalg.h` + `linalg/ops.h` | 19 | ✅ |
| `random.h` | 19 | ✅ |
| `compile.h` + `compile_impl.h` | 7 | ✅ |
| `io.h` | 10 | ✅ |
| `distributed/ops.h` | 8 | ✅ |
| `transforms.h` | 8 | ✅ |
| `fast.h` | 23 | ✅ |
| `graph_utils.h` | 5 | ✅ |
| `export.h` | Full API | ✅ — via module hook system |
| `metal.h` | 3 | ⚠️ Not regression-tested — requires MLX compiled with Metal backend (`brew install mlx`). No separate Metal SDK needed; Metal is part of Xcode/CLT, which is already a build prerequisite. | Don't Worry... We Are Testing On Our Own Project...

---

## How to Run

```bash
# Build
go build ./cmd/mlxgen

# Static generators
./mlxgen --type=vector
./mlxgen --type=vector --implementation
./mlxgen --type=map
./mlxgen --type=closure
./mlxgen --type=private --ctype=mlx_array --cpptype='mlx::core::array'

# Dynamic generator — requires MLX source
./mlxgen --header path/to/mlx/fft.h --headername fft
./mlxgen --header path/to/mlx/fft.h --headername fft --implementation

# CGo wrappers
./mlxgen --cgo-from mlx/c/fft.h --package mlx

# Go base scaffolding
./mlxgen --type=go-base --package mlx > mlx/go/mlx.go

# Regenerate all static files
./scripts/regenerate.sh

# Via CMake
cmake -B build -DMLX_C_REGENERATE_BINDINGS=ON
cmake --build build --target mlxgen_static
cmake --build build

# Tests
go test ./...
go test ./... -race
```

# mlx-cgo

MLX is Apple's machine learning framework for Apple Silicon. Fast, unified, native. They gave it a C API. We gave it Go.

mlx-cgo is a fork of [mlx-c](https://github.com/ml-explore/mlx-c) that does two things:

1. **Replaces the Python code-generator** with a single Go binary — no Python, no `pip install`, no external dependencies
2. **Adds a first-class CGo bindings generator** — idiomatic Go wrappers over the full C API, built for Apple Silicon

---

## Requirements

- **Apple Silicon** Mac (M-series chip)
- **Xcode** or Command Line Tools — `xcode-select --install`
- **CMake** — `brew install cmake`
- **Go** — `brew install go`

---

## Build

```bash
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j
```

That's it. The committed `mlx/c/` files are already generated — no Python, no regeneration step needed for a normal build.

---

## The Generator: `mlxgen`

The code-generator lives at `cmd/mlxgen`. One binary. Every generation mode you need.

```bash
go build ./cmd/mlxgen
```

### Generate C bindings

```bash
# .h
./mlxgen --header path/to/mlx/fft.h --headername fft

# .cpp
./mlxgen --header path/to/mlx/fft.h --headername fft --implementation

# Static generators — vector, map, closure, private headers
./mlxgen --type=vector
./mlxgen --type=map
./mlxgen --type=closure
./mlxgen --type=private --ctype=mlx_array --cpptype='mlx::core::array'

# Regenerate everything at once
./scripts/regenerate.sh
```

### Generate Go/CGo wrappers

```bash
./mlxgen --cgo-from mlx/c/fft.h --package mlx
```

Produces real Go functions with proper types and error returns:

```go
func FftFft(a Array, n int32, axis int32, s Stream) (Array, error) {
    var res C.mlx_array
    if ret := C.mlx_fft_fft(a.c, C.int(n), C.int(axis), s.c, &res); ret != 0 {
        return Array{}, fmt.Errorf("mlx_fft_fft failed (code %d)", ret)
    }
    return Array{c: res}, nil
}
```

### Go base scaffolding

```bash
./mlxgen --type=go-base --package mlx > mlx/go/mlx.go
```

Generates the foundation: opaque structs, `Free()` methods, `DefaultStream()`, `DefaultDevice()`, `Dtype` constants.

---

## Tests

```bash
go test ./...
go test ./... -race
```

71 tests. 100% passing. Race detector clean.

---

## Usage

```go
import "github.com/WhoseBiasDoYallSeek/mlx-cgo/mlx"

s := mlx.DefaultStream()
defer s.Free()

a, err := mlx.Full([]int32{3, 3}, float32(1.0), s)
if err != nil {
    log.Fatal(err)
}
defer a.Free()

result, err := mlx.Abs(a, s)
if err != nil {
    log.Fatal(err)
}
defer result.Free()
```

Memory is explicit: call `.Free()` on every returned handle. In hot loops, call it directly — `defer` fires at function exit, not loop iteration.

---

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full technical documentation: data flow, type system, CGo generator, memory model, CMake integration, and test coverage.

---

## License

MIT. See [LICENSE](LICENSE).

© 2023 Apple Inc. (original mlx-c)  
© 2026 WhoseBiasDoYallSeek (mlx-cgo fork)

See [ATTRIBUTION.md](ATTRIBUTION.md) for the full credits.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). See [ACKNOWLEDGMENTS.md](ACKNOWLEDGMENTS.md) for contributors.

---

## FAQ

**Why not Rust?**

R: 

We at SRARS don't fight the compiler, we build the compiler. We don't follow Go Proverbs like a Dogma so we prefer to "defer" to Eclesiastes. 

"Vanity of vanities, said Koheleth; vanity of vanities, all is vanity."

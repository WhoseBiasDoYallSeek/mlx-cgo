# mlx-cgo: Troubleshooting Guide

## Quick Diagnostics

```bash
# Check your environment (Apple Silicon macOS only)
uname -m                # Must be: arm64
uname -s                # Must be: Darwin
clang --version         # Need 14+
go version              # Need 1.19+
cmake --version         # Need 3.24+
```

If not `arm64` on Darwin:
- ✅ mlx-cgo is not for your platform (intentional)
- Use other inference engines for that hardware

---

## Problem: Go Code Panics on Empty Arrays

### Symptom
```
panic: runtime error: index out of range [0] with length 0
goroutine 1 [running]:
  github.com/WhoseBiasDoYallSeek/mlx-cgo/generated.mlxArraySlice(...)
```

### Cause
Old mlx-cgo (<v1.0) generates unsafe `&slice[0]` without length check.

### Solution
**Upgrade to mlx-cgo v1.0+**:
```bash
git pull origin main
go build ./cmd/mlxgen
./scripts/regenerate.sh
cmake --build build
```

### Workaround (if stuck on old version)
```go
// Bad:
result := mlxArraySlice(array, []int32{})  // Crashes

// Good:
if len(axes) > 0 {
    result := mlxArraySlice(array, axes)
}
```

---

## Problem: Memory Grows Over Time (Memory Leak)

### Symptom
```
App uses 50MB after startup, 200MB after 1 hour
No goroutines leaking, no Go memory issues
```

### Cause
mlx-cgo <v1.0 doesn't `free()` C strings. Each call leaks ~100 bytes.

### Solution
**Upgrade to mlx-cgo v1.0+**:
```bash
git pull origin main
```

Generated code now emits `defer C.free(unsafe.Pointer(...))`.

### Verify
```bash
go test ./... -v  # Should show "defer C.free" pattern
grep -r "defer C.free" mlx/c/  # Confirm freed strings
```

---

## Problem: Build Fails - "ast-dump=json not available"

### Symptom
```bash
$ cmake --build build
Error: unknown argument: '-ast-dump=json'
```

### Cause
Your Clang is <14. Apple Clang 13 (macOS 12) doesn't support JSON AST output.

### Solution
Upgrade Clang:
```bash
# Option 1: Upgrade Xcode
xcode-select --install

# Option 2: Use Homebrew
brew install llvm@16
cmake -B build -DCMAKE_C_COMPILER=/usr/local/opt/llvm@16/bin/clang

# Option 3: Check version
clang --version  # Should be >= 14
```

### Verify
```bash
clang -Xclang -ast-dump=json -fsyntax-only /dev/null
# If OK: <TranslationUnitDecl...
# If bad: error: unknown argument
```

---

## Problem: Generators Don't Find MLX Functions

### Symptom
```bash
$ ./mlxgen --header mlx/include/mlx/c/ops.h --headername ops
Error: couldn't find "mlx_array_cumsum" in header
```

### Cause
MLX version mismatch. Your MLX is too old (v0.30) or too new (v0.32+).

### Solution
Check MLX version:
```bash
cd build/_deps/mlx-src
git describe --tags  # Should be v0.31.2
```

If not v0.31.2:
```bash
# Update to v0.31.2
git fetch --all
git checkout v0.31.2
cd ../..
rm -rf build
cmake -B build
cmake --build build -j
```

### Verify
```bash
./mlxgen --header mlx/include/mlx/c/ops.h --headername ops 2>&1 | grep -i error
# Should be silent (no errors)
```

---

## Problem: Build Fails on Linux or x86_64

### Not Applicable ✅

**mlx-cgo is intentionally Apple Silicon + macOS only.**

- MLX is an Apple framework (Metal GPU, optimized for arm64)
- Narrow focus = better quality

**This is by design, not a limitation.**

If you need cross-platform inference:
- Use ONNX Runtime for x86_64/Linux
- Use PyTorch or TensorFlow for other hardware

**mlx-cgo is the macOS backend. It excels at that.**

---

## Problem: CGo Crashes Under Concurrency

### Symptom
```
Occasional segfault: "memory access violation"
Only happens with 50+ goroutines
```

### Cause
Unsynchronized access to C memory. Race condition.

### Solution
Serialize C calls:
```go
var mu sync.Mutex

go func() {
    mu.Lock()
    result := mlxArrayLoad("model.safetensors")
    mu.Unlock()
    
    // Safe from here (loaded in C)
    _ = mlxArrayShape(result)
    
    mu.Lock()
    mlxArrayFree(result)
    mu.Unlock()
}()
```

### Verify
```bash
go test ./... -race  # Run with race detector
# If clean: "race detector enabled"
# If bad: "WARNING: DATA RACE"
```

---

## Problem: "Type mismatch" Compile Error

### Symptom
```
cannot use axes (type []int64) as type []int32 in argument
```

### Cause
CGo generates strict C types (int32, float32). Go requires exact match.

### Solution
Convert types:
```go
// Bad:
axes := []int64{0, 1}
mlxArraySlice(array, axes)  // Compile error

// Good:
axes64 := []int64{0, 1}
axes32 := make([]int32, len(axes64))
for i, v := range axes64 {
    axes32[i] = int32(v)
}
mlxArraySlice(array, axes32)
```

Or define type aliases:
```go
type MLXInt = int32
type MLXFloat = float32

axes := []MLXInt{0, 1}
mlxArraySlice(array, axes)
```

---

## Problem: App Leaks Memory (Go Side)

### Symptom
```
go test -run TestLoadModels
Memory grows: 10MB → 100MB → OOM
```

### Cause
Forgetting `defer mlxArrayFree()` or similar.

### Solution
Always defer cleanup:
```go
// Bad:
func LoadModel(path string) {
    array := mlxArrayLoad(path)
    // Forgot to free!
}

// Good:
func LoadModel(path string) *mlxArray {
    array := mlxArrayLoad(path)
    runtime.SetFinalizer(array, func(a *mlxArray) {
        mlxArrayFree(a)
    })
    return array
}
```

Or manual:
```go
result := mlxArrayLoad(path)
defer mlxArrayFree(result)
// Use result...
```

### Verify
```bash
# Check for leaks with Go's leak detector
go test -race -run TestLoadModels
# Or use valgrind on macOS:
leaks -atExit -- ./app
```

---

## Problem: Changes to `.h` Files Not Regenerated

### Symptom
```bash
$ git diff mlx/c/private/ops.h
# Shows old code (100+ lines different)
# But no actual edits were made
```

### Cause
Someone edited MLX upstream, but didn't regenerate bindings.

### Solution
Regenerate:
```bash
./scripts/regenerate.sh
# Or manually:
go build ./cmd/mlxgen
./mlxgen --all
```

Commit changes:
```bash
git add mlx/c/private/*.h
git add mlx/c/*.cpp
git commit -m "Regenerate bindings for MLX v0.31.2"
```

---

## Problem: CMake Can't Find Clang

### Symptom
```bash
$ cmake -B build
Error: CMAKE_C_COMPILER not found
```

### Cause
Clang not in PATH.

### Solution
```bash
# Option 1: Install via Xcode
xcode-select --install

# Option 2: Install via Homebrew
brew install llvm@16

# Option 3: Specify path
cmake -B build -DCMAKE_C_COMPILER=/usr/local/opt/llvm@16/bin/clang
```

### Verify
```bash
which clang
clang --version  # Should be >= 14
```

---

## Problem: Go Modules Error

### Symptom
```
error: package xxx requires Go x.xx but current is x.xx
```

### Cause
Go version too old.

### Solution
Upgrade Go:
```bash
brew install go@1.21
# Or download from golang.org
```

### Verify
```bash
go version  # Should be >= 1.19
```

---

## Problem: CMake "Can't Find MLX"

### Symptom
```bash
$ cmake --build build
Error: MLX not found at /path/to/mlx
```

### Cause
MLX submodule not initialized or updated.

### Solution
```bash
git submodule update --init --recursive
cmake -B build
cmake --build build -j
```

### Verify
```bash
ls -la build/_deps/mlx-src
# Should have libmlx.a
```

---

## Common Fixes Checklist

Before reporting an issue, verify:

- [ ] Using mlx-cgo v1.0+ (`git log --oneline | head -1`)
- [ ] Clang 14+ (`clang --version`)
- [ ] Go 1.19+ (`go version`)
- [ ] MLX v0.31.2 (`git -C build/_deps/mlx-src describe --tags`)
- [ ] CMake 3.24+ (`cmake --version`)
- [ ] Running on arm64 (`uname -m`)
- [ ] All tests pass (`go test ./... -v`)
- [ ] Race detector clean (`go test ./... -race`)
- [ ] No stale generated code (`./scripts/regenerate.sh`)

---

## Getting Help

1. **Check this guide first** (you're reading it!)
2. **Check ARCHITECTURE.md** (detailed system design)
3. **Check MAINTENANCE.md** (upstream sync strategy)
4. **Search GitHub Issues** (might be known problem)
5. **File a new issue** with:
   - Your environment (OS, Go version, Clang version)
   - Full error message
   - Steps to reproduce
   - Output of `go test ./... -v -race`

---

## Performance Tips

If mlx-cgo is slow:

1. **Check compilation flags**:
   ```bash
   cmake -B build -DCMAKE_BUILD_TYPE=Release  # Not Debug
   cmake --build build -j
   ```

2. **Avoid serialization bottlenecks**:
   ```go
   // Slow: one mutex for all calls
   mu.Lock()
   result1 := mlxArrayLoad(...)
   result2 := mlxArrayLoad(...)
   mu.Unlock()
   
   // Better: batch operations
   batch := []string{...}
   results := make([]*mlxArray, len(batch))
   for i, path := range batch {
       results[i] = mlxArrayLoad(path)  // Serialize only startup
   }
   // Then use results in parallel
   ```

3. **Use streaming for large models**:
   ```go
   // Instead of loading entire model:
   array := mlxArrayLoad("model.safetensors")  // All at once
   
   // Consider streaming (if MLX supports it):
   stream := mlxStreamCreate()
   chunk1 := mlxArrayLoad("model_part1.safetensors", stream)
   chunk2 := mlxArrayLoad("model_part2.safetensors", stream)
   ```

---

## Still Stuck?

Add this debug output:
```go
import "C"

func init() {
    println("mlx-cgo build info:")
    println("  MLX version:", C.MLX_VERSION_STRING)
    println("  Go version:", runtime.Version())
    println("  Platform:", runtime.GOOS, runtime.GOARCH)
}
```

Then file a GitHub issue with full output.

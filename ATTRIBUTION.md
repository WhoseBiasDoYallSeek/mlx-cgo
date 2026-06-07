# Attribution & Licensing

## License

**mlx-cgo** is licensed under the **MIT License**. See [LICENSE](LICENSE) for full text.

This license is compatible with and derived from the original mlx-c project, which is also MIT-licensed by ml-explore.

### Dual Copyright

- **Original mlx-c (2023):** © 2023 ml-explore — [github.com/ml-explore/mlx-c](https://github.com/ml-explore/mlx-c)
- **mlx-cgo fork (2026):** © 2026 WhoseBiasDoYallSeek

Both projects are MIT-licensed. All rights are retained per the MIT license terms.

## Code Origin

### Ported from Python (ml-explore)

The following modules in mlx-cgo are direct ports or heavily derived from the original Python codebase in mlx-c:

| Original (Python) | Ported to (Go) | Copyright |
|---|---|---|
| `mlxtypes.py` (643 LOC) | `internal/types/ctypes.go` | © 2023 ml-explore |
| `mlxvariants.py` (143 LOC) | `internal/variants/variants.go` | © 2023 ml-explore |
| `mlxhooks.py` (508 LOC) | `internal/hooks/hooks.go` | © 2023 ml-explore |
| `vector_generator.py` (338 LOC) | `internal/codegen/vector.go` | © 2023 ml-explore |
| `map_generator.py` (342 LOC) | `internal/codegen/map.go` | © 2023 ml-explore |
| `closure_generator.py` (395 LOC) | `internal/codegen/closure.go` | © 2023 ml-explore |
| `type_private_generator.py` (123 LOC) | `internal/codegen/private.go` | © 2023 ml-explore |
| `c.py` + `generator.py` (394 LOC) | `internal/codegen/c_header.go` + `c_impl.go` | © 2023 ml-explore |

These are algorithmic ports — functionality preserved, implementation language changed from Python to Go.

### New in mlx-cgo (WhoseBiasDoYallSeek)

The following modules are original to mlx-cgo and not derived from mlx-c:

| Module | Lines | Purpose |
|--------|-------|---------|
| `internal/parser/` | ~400 | Clang AST parsing (replaces cxxheaderparser) |
| `internal/codegen/cgo/` | ~350 | CGo wrapper generation (new feature) |
| `internal/codegen/cgo/scaffold.go` | ~200 | Go base package scaffolding (new) |
| `internal/cheader/` | ~220 | Header parsing for CGo (new) |
| `scripts/regenerate.sh` | ~130 | Build orchestration (new) |
| `cmd/mlxgen/main.go` | ~150 | CLI interface (replaces generator.py) |

---

## MIT License Text

```
MIT License

Copyright (c) 2023 ml-explore
Copyright (c) 2026 WhoseBiasDoYallSeek

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## Acknowledgments

- **Apple MLX team** — for the original ml-explore/mlx project
- **ml-explore** — for the original mlx-c C bindings generator in Python
- **Contributors** — see git history and CONTRIBUTORS file (if any)

---

## How to Attribute

If you use mlx-cgo in your project, we ask that you:

1. Include the LICENSE file (or equivalent notice)
2. Mention both projects if giving credit:
   - mlx-cgo (WhoseBiasDoYallSeek)
   - mlx-c (ml-explore)
3. Specify you're using the Go version to avoid confusion

Example:

```
This project uses mlx-cgo, a Go rewrite of mlx-c (ml-explore),
licensed under MIT.
```

---

**Questions?** Open an issue on the mlx-cgo repository.

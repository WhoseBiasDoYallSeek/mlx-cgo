# Contributing to MLX-CGO

We want to make contributing to this project as easy and transparent as
possible.

## Pull Requests

1. Fork and submit pull requests to the repo. 
2. If you've added code that should be tested, add tests.
3. If a change is likely to impact efficiency, run benchmarks before/after. See `internal/benchmark_test.go`.
4. If you've changed APIs, update the documentation.
5. Every PR should have passing tests and at least one review.
6. For code formatting:
   - **Go**: `gofmt -w file.go`, `goimports -w file.go`
   - **C/C++**: `clang-format -i file.cpp file.c`
   - Run `go vet ./...` and `go test ./...` before submitting.
   
   Or use `make fmt` if available.

## Issues

We use GitHub issues to track public bugs. Please ensure your description is
clear and has sufficient instructions to be able to reproduce the issue.

## License

By contributing to MLX-CGO, you agree that your contributions will be
licensed under the LICENSE file in the root directory of this source tree.

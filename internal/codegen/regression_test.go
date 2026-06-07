// Package codegen_test provides regression tests that verify generator output
// matches the committed C binding files in mlx/c/.
//
// Run from repository root:
//
//	go test ./internal/codegen/ -run TestRegression
package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/codegen"
)

// repoRoot returns the absolute path to the repository root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine caller path")
	}
	// internal/codegen/ → two levels up
	return filepath.Join(filepath.Dir(file), "..", "..")
}

// clangFormat runs clang-format on src with the repo's .clang-format config
// and the given assumed filename (for main-include heuristic in include sorting).
func clangFormat(t *testing.T, src []byte, assumeFilename, repoDir string) []byte {
	t.Helper()
	cfmt, err := exec.LookPath("clang-format")
	if err != nil {
		// Try xcrun on macOS
		out, xerr := exec.Command("xcrun", "--find", "clang-format").Output()
		if xerr != nil {
			t.Skip("clang-format not found; skipping regression test")
		}
		cfmt = strings.TrimSpace(string(out))
	}

	cmd := exec.Command(cfmt, "--style=file", "--assume-filename="+assumeFilename)
	cmd.Dir = repoDir
	cmd.Stdin = bytes.NewReader(src)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("clang-format: %v", err)
	}
	return out
}

// runRegressionTest generates output from fn, formats it with clang-format,
// reads the golden file from mlx/c/<name>, and compares them.
func runRegressionTest(t *testing.T, name string, fn func(*bytes.Buffer)) {
	t.Helper()
	root := repoRoot(t)

	var got bytes.Buffer
	fn(&got)

	formatted := clangFormat(t, got.Bytes(), name, root)

	golden := filepath.Join(root, "mlx", "c", name)
	wantBytes, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden %q: %v", golden, err)
	}

	normalize := func(b []byte) string {
		var sb strings.Builder
		for _, line := range strings.Split(string(b), "\n") {
			sb.WriteString(strings.TrimRight(line, " \t"))
			sb.WriteString("\n")
		}
		return strings.TrimSpace(sb.String())
	}

	gotNorm := normalize(formatted)
	wantNorm := normalize(wantBytes)

	if gotNorm != wantNorm {
		gotLines := strings.Split(gotNorm, "\n")
		wantLines := strings.Split(wantNorm, "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Errorf("%s: first diff at line %d:\n  got:  %q\n  want: %q",
					name, i+1, gotLines[i], wantLines[i])
				break
			}
		}
		if len(gotLines) != len(wantLines) {
			t.Errorf("%s: line count: got %d, want %d", name, len(gotLines), len(wantLines))
		}
	}
}

// runPrivateRegressionTest generates a private header using GeneratePrivateHeader
// and compares it against the golden file in mlx/c/private/<shortname>.h.
func runPrivateRegressionTest(
	t *testing.T,
	shortname, ctype, cpptype string,
	noCopy bool,
	usingDecl string,
) {
	t.Helper()
	name := "private/" + shortname + ".h"
	runRegressionTest(t, name, func(b *bytes.Buffer) {
		codegen.GeneratePrivateHeader(b, ctype, cpptype, noCopy, shortname, "", usingDecl)
	})
}

func TestRegressionPrivateArray(t *testing.T) {
	runPrivateRegressionTest(t, "array", "mlx_array", "mlx::core::array", false, "")
}

func TestRegressionPrivateStream(t *testing.T) {
	runPrivateRegressionTest(t, "stream", "mlx_stream", "mlx::core::Stream", false, "")
}

func TestRegressionPrivateDevice(t *testing.T) {
	const (
		ctypes   = "mlx_device;mlx_device_info"
		cpptypes = "mlx::core::Device;std::unordered_map<std::string, std::variant<std::string, size_t>>"
		usings   = ";mlx_device_info_cpp"
	)
	runPrivateRegressionTest(t, "device", ctypes, cpptypes, false, usings)
}

func TestRegressionPrivateClosure(t *testing.T) {
	runPrivateRegressionTest(t, "closure",
		"mlx_closure",
		"std::function<std::vector<array>(std::vector<array>)>",
		false, "")
}

func TestRegressionVectorHeader(t *testing.T) {
	runRegressionTest(t, "vector.h", func(b *bytes.Buffer) {
		codegen.GenerateVectorHeader(b)
	})
}

func TestRegressionVectorImpl(t *testing.T) {
	runRegressionTest(t, "vector.cpp", func(b *bytes.Buffer) {
		codegen.GenerateVectorImpl(b)
	})
}

func TestRegressionMapHeader(t *testing.T) {
	runRegressionTest(t, "map.h", func(b *bytes.Buffer) {
		codegen.GenerateMapHeader(b)
	})
}

func TestRegressionMapImpl(t *testing.T) {
	runRegressionTest(t, "map.cpp", func(b *bytes.Buffer) {
		codegen.GenerateMapImpl(b)
	})
}

func TestRegressionClosureHeader(t *testing.T) {
	runRegressionTest(t, "closure.h", func(b *bytes.Buffer) {
		codegen.GenerateClosureHeader(b)
	})
}

func TestRegressionClosureImpl(t *testing.T) {
	runRegressionTest(t, "closure.cpp", func(b *bytes.Buffer) {
		codegen.GenerateClosureImpl(b)
	})
}

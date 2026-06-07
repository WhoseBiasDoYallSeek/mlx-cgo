package parser

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ParseHeader runs clang on the given C++ header file and returns the root
// AST node. Multiple headers can be provided separated by semicolons, which
// mirrors the behaviour of the original Python generator.py.
func ParseHeader(headers string, extraArgs []string) (*Node, error) {
	// Build a tiny translation unit that includes all requested headers so
	// clang sees them in a single AST pass.
	var sb strings.Builder
	for _, h := range strings.Split(headers, ";") {
		h = strings.TrimSpace(h)
		if h != "" {
			fmt.Fprintf(&sb, "#include \"%s\"\n", h)
		}
	}
	tu := sb.String()

	args := []string{
		"-Xclang", "-ast-dump=json",
		"-fsyntax-only",
		"-x", "c++-header",
		"-std=c++20",
	}
	args = append(args, extraArgs...)
	// Read source from stdin so we don't need a temp file.
	args = append(args, "-")

	cmd := exec.Command("clang", args...)
	cmd.Stdin = strings.NewReader(tu)

	out, err := cmd.Output()
	if err != nil {
		// clang returns exit code 1 even on success when there are warnings;
		// only treat it as an error when we got no output at all.
		if len(out) == 0 {
			return nil, fmt.Errorf("clang failed: %w", err)
		}
	}

	var root Node
	if err := json.Unmarshal(out, &root); err != nil {
		return nil, fmt.Errorf("parsing clang JSON: %w", err)
	}
	return &root, nil
}

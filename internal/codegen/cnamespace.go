// Package codegen contains all code generation backends (C header, C impl,
// Go/CGo, vector, map, closure, private).
package codegen

import (
	"strings"
	"unicode"
)

// ToSnakeLetters converts a CamelCase identifier to snake_case.
// Mirrors the to_snake_letters function from python/c.py.
//
//	"MyName"   → "my_name"
//	"FFTNorm"  → "fft_norm"
//	"getValue" → "get_value"
func ToSnakeLetters(name string) string {
	runes := []rune(name)
	var out []rune
	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Insert underscore before an upper-case letter when:
			//  1. Not the first character.
			//  2. Previous char is lower-case or digit, OR
			//     next char exists and is lower-case (handles "FFTNorm" → "fft_norm").
			if i > 0 {
				prev := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) ||
					(unicode.IsUpper(prev) && nextIsLower) {
					out = append(out, '_')
				}
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// CNamespace converts a C++ namespace string into a C prefix.
// Mirrors c_namespace() in python/c.py.
//
//	"mlx::core"           → "mlx"
//	"mlx::core::random"   → "mlx_random"
func CNamespace(namespace string) string {
	parts := strings.Split(namespace, "::")
	if len(parts) >= 2 && parts[0] == "mlx" && parts[1] == "core" {
		parts = append(parts[:1], parts[2:]...) // drop "core"
	}
	return strings.Join(parts, "_")
}

// FuncCName builds the full C function name for a given function, combining
// the namespace prefix and the function name (plus optional variant suffix).
//
//	namespace="mlx::core", name="abs", variant=""    → "mlx_abs"
//	namespace="mlx::core", name="squeeze", variant="axes" → "mlx_squeeze_axes"
func FuncCName(namespace, name, variant string) string {
	prefix := CNamespace(namespace)
	cname := prefix + "_" + name
	if variant != "" {
		cname += "_" + variant
	}
	return cname
}

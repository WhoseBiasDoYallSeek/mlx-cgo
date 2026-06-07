// Package cheader parses the generated C header files (.h) to extract
// function declarations, which are then used by the CGo generator to
// produce idiomatic Go wrappers.
//
// The C headers follow a strict pattern produced by our own generators:
//
//	int mlx_<namespace>_<func>(
//	    <type>* res,          // output param (may have multiple)
//	    const <type> arg1,    // input params
//	    ...);
//
// This parser handles multi-line and single-line declarations.
package cheader

import (
	"bufio"
	"bytes"
	"strings"
)

// CParam represents a single parameter in a C function declaration.
type CParam struct {
	// Type is the C type string (e.g. "mlx_array*", "const mlx_array", "int").
	Type string
	// Name is the parameter name (may be empty for unnamed params).
	Name string
	// IsReturn is true when this is the out-param (pointer to the return value).
	IsReturn bool
	// IsRawVecPtr is true when this param is the pointer half of a (T*, size_t N) pair.
	IsRawVecPtr bool
	// IsRawVecLen is true when this param is the size_t half of a (T*, size_t N) pair.
	IsRawVecLen bool
	// RawVecElem is the C element type for raw vector params (e.g. "int", "float").
	RawVecElem string
}

// CFunc is a parsed C function declaration from a .h file.
type CFunc struct {
	// Name is the full C name (e.g. "mlx_fft_fft").
	Name string
	// Params is the ordered list of parameters.
	Params []CParam
}

// ParseHeader parses a C header file content and returns all mlx_ function
// declarations found. It handles both single-line and multi-line declarations.
func ParseHeader(src []byte) []CFunc {
	var funcs []CFunc

	scanner := bufio.NewScanner(bytes.NewReader(src))
	var accum strings.Builder
	inFunc := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines, comments, preprocessor
		if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "typedef") ||
			strings.HasPrefix(trimmed, "static") || strings.HasPrefix(trimmed, "extern") {
			continue
		}

		// Start of a function: "int mlx_"
		if !inFunc && strings.HasPrefix(trimmed, "int mlx_") {
			accum.Reset()
			accum.WriteString(trimmed)
			inFunc = true
			// Single-line if it ends with ";"
			if strings.HasSuffix(trimmed, ";") {
				if f, ok := parseDecl(accum.String()); ok {
					funcs = append(funcs, f)
				}
				inFunc = false
			}
			continue
		}

		if inFunc {
			accum.WriteString(" ")
			accum.WriteString(trimmed)
			if strings.HasSuffix(trimmed, ";") {
				if f, ok := parseDecl(accum.String()); ok {
					funcs = append(funcs, f)
				}
				inFunc = false
			}
		}
	}
	return funcs
}

// parseDecl parses a single accumulated declaration string like:çƒ
// "int mlx_fft_fft( mlx_array* res, const mlx_array a, int n, ... );"
func parseDecl(s string) (CFunc, bool) {
	s = strings.TrimSuffix(strings.TrimSpace(s), ";")
	// Remove leading "int "
	s = strings.TrimPrefix(s, "int ")
	paren := strings.Index(s, "(")
	if paren < 0 {
		return CFunc{}, false
	}
	name := strings.TrimSpace(s[:paren])
	inner := s[paren+1:]
	// Remove trailing ")"
	if idx := strings.LastIndex(inner, ")"); idx >= 0 {
		inner = inner[:idx]
	}

	var params []CParam
	for _, raw := range splitParams(inner) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		p := parseParam(raw)
		params = append(params, p)
	}

	// First pointer param (non-func-ptr) is the return value
	for i := range params {
		if strings.Contains(params[i].Type, "*") && !isFuncPtr(params[i].Type) {
			params[i].IsReturn = true
			break
		}
	}

	// Mark (const T*, size_t T_num) pairs as raw vector
	markRawVecPairs(params)

	return CFunc{Name: name, Params: params}, true
}

// isFuncPtr returns true for function-pointer types like "int (*fun)(...)".
func isFuncPtr(t string) bool {
	return strings.Contains(t, "(*")
}

// parseParam splits a raw param string like "const mlx_array a" into type+name.
func parseParam(s string) CParam {
	// Handle function pointers: "int (*fun)(args)"
	if isFuncPtr(s) {
		return CParam{Type: s, Name: ""}
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return CParam{Type: s}
	}

	last := parts[len(parts)-1]
	// If the last part is a type component, no name
	if strings.HasSuffix(last, "*") || last == "const" || last == "unsigned" {
		return CParam{Type: s}
	}

	name := last
	typePart := strings.Join(parts[:len(parts)-1], " ")

	// Move "*" that sticks to name into type
	if strings.HasPrefix(name, "*") {
		typePart += "*"
		name = strings.TrimPrefix(name, "*")
	}

	return CParam{Type: typePart, Name: name}
}

// markRawVecPairs detects consecutive (const T*, size_t T_num) pairs and
// marks them so the CGo generator can fold them into []T.
func markRawVecPairs(params []CParam) {
	rawPtrs := map[string]string{
		"const int*":    "int",
		"const float*":  "float",
		"const double*": "double",
		"const bool*":   "bool",
	}
	for i := 0; i < len(params)-1; i++ {
		t := strings.TrimSpace(params[i].Type)
		elem, ok := rawPtrs[t]
		if !ok {
			continue
		}
		// Look for the following size_t param named "<name>_num"
		next := params[i+1]
		nt := strings.TrimSpace(next.Type)
		if nt == "size_t" && strings.HasSuffix(next.Name, "_num") {
			params[i].IsRawVecPtr = true
			params[i].RawVecElem = elem
			params[i+1].IsRawVecLen = true
		}
	}
}

// splitParams splits a parameter list string by top-level commas.
func splitParams(s string) []string {
	var result []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '(', '<':
			depth++
		case ')', '>':
			depth--
		case ',':
			if depth == 0 {
				result = append(result, s[start:i])
				start = i + 1
			}
		}
	}
	result = append(result, s[start:])
	return result
}

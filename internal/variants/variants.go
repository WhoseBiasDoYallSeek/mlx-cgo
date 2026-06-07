// Package variants resolves function overload naming for the C binding layer.
//
// This is a direct port of python/mlxvariants.py from the original mlx-c project.
// When multiple C++ overloads of the same function exist, each overload receives
// a suffix (its "variant") to produce a unique C function name.
//
// A nil suffix means the overload should be dropped entirely.
// An empty string "" means no suffix is appended (the function keeps its base name).
package variants

import "github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/model"

// variantTable maps function name → ordered list of variant suffixes.
// Each entry in the list corresponds to one overload (sorted by decreasing
// number of parameters, matching what generator.py did).
// nil   → drop this overload entirely.
// ""    → keep the overload with no suffix.
// "foo" → append "_foo" to the C function name.
type variantTable map[string][]*string

// ptr helpers to keep the table readable.
func s(v string) *string { return &v }
func drop() *string      { return nil }
func keep() *string      { return s("") }

// Resolve applies the variant table for the given namespace to the list of
// overloads of a single function name. It returns the (possibly reduced) list
// with Variant set on each remaining Function.
//
// nsKey is the dot-separated namespace key used to select the right table,
// e.g. "mlx_core", "mlx_core_random".
func Resolve(nsKey, funcName string, defs []model.Function) []model.Function {
	table, ok := tables[nsKey]
	if !ok {
		// No table for this namespace: keep only the first overload.
		if len(defs) > 0 {
			return defs[:1]
		}
		return nil
	}

	suffixes, ok := table[funcName]
	if !ok {
		// Function not in table: keep only the first overload.
		if len(defs) > 0 {
			return defs[:1]
		}
		return nil
	}

	if len(suffixes) != len(defs) {
		// Mismatch — log to stderr in a real run; here we just keep first.
		if len(defs) > 0 {
			return defs[:1]
		}
		return nil
	}

	var out []model.Function
	for i, suf := range suffixes {
		if suf == nil {
			continue // drop this overload
		}
		f := defs[i]
		f.Variant = *suf
		out = append(out, f)
	}
	return out
}

// NSKey converts a namespace string like "mlx::core::random" into the table
// lookup key "mlx_core_random".
func NSKey(namespace string) string {
	key := ""
	for i, seg := range splitNS(namespace) {
		if i > 0 {
			key += "_"
		}
		key += seg
	}
	return key
}

func splitNS(ns string) []string {
	var segs []string
	cur := ""
	for _, c := range ns {
		if c == ':' {
			if cur != "" {
				segs = append(segs, cur)
				cur = ""
			}
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		segs = append(segs, cur)
	}
	return segs
}

// tables is the registry of all namespace variant tables.
var tables = map[string]variantTable{
	"mlx_core": {
		"arange":           []*string{keep(), drop(), drop(), drop(), drop(), drop(), drop(), drop(), drop()},
		"eye":              []*string{keep(), drop(), drop(), drop(), drop()},
		"tri":              []*string{keep(), drop()},
		"flatten":          []*string{keep(), drop()},
		"squeeze":          []*string{s("axes"), s("axis"), keep()},
		"expand_dims":      []*string{s("axes"), keep()},
		"slice":            []*string{keep(), drop(), s("dynamic"), drop()},
		"slice_update":     []*string{keep(), drop(), s("dynamic")},
		"split":            []*string{keep(), s("sections"), drop(), drop()},
		"concatenate":      []*string{s("axis"), keep()},
		"stack":            []*string{s("axis"), keep()},
		"repeat":           []*string{s("axis"), keep()},
		"transpose":        []*string{s("axes"), drop(), keep()},
		"all":              []*string{s("axes"), s("axis"), keep(), drop()},
		"any":              []*string{s("axes"), s("axis"), keep(), drop()},
		"sum":              []*string{s("axes"), s("axis"), keep(), drop()},
		"mean":             []*string{s("axes"), s("axis"), keep(), drop()},
		"var":              []*string{s("axes"), s("axis"), keep(), drop()},
		"std":              []*string{s("axes"), s("axis"), keep(), drop()},
		"prod":             []*string{s("axes"), s("axis"), keep(), drop()},
		"max":              []*string{s("axes"), s("axis"), keep(), drop()},
		"min":              []*string{s("axes"), s("axis"), keep(), drop()},
		"argmax":           []*string{s("axis"), keep(), drop()},
		"argmin":           []*string{s("axis"), keep(), drop()},
		"load":             []*string{s("reader"), keep()},
		"load_safetensors": []*string{s("reader"), keep()},
		"pad":              []*string{keep(), drop(), drop(), s("symmetric")},
		"save":             []*string{s("writer"), keep()},
		"save_safetensors": []*string{s("writer"), keep()},
		"gather":           []*string{keep(), s("single")},
		"scatter":          []*string{keep(), s("single")},
		"scatter_add":      []*string{keep(), s("single")},
		"scatter_min":      []*string{keep(), s("single")},
		"scatter_prod":     []*string{keep(), s("single")},
		"scatter_max":      []*string{keep(), s("single")},
		"argpartition":     []*string{s("axis"), keep()},
		"partition":        []*string{s("axis"), keep()},
		"argsort":          []*string{s("axis"), keep()},
		"sort":             []*string{s("axis"), keep()},
		"topk":             []*string{s("axis"), keep()},
		"take":             []*string{s("axis"), drop(), keep(), drop()},
		"roll":             []*string{drop(), drop(), s("axis"), s("axes"), drop(), keep()},
		"logsumexp":        []*string{s("axes"), s("axis"), keep(), drop()},
		"softmax":          []*string{s("axes"), s("axis"), keep()},
		"tensordot":        []*string{keep(), s("axis")},
		"array_equal":      []*string{keep(), drop()},
		"round":            []*string{keep(), drop()},
		"trace":            []*string{keep(), drop(), drop()},
		"export_function":  []*string{drop(), keep(), s("kwargs")},
	},

	"mlx_core_linalg": {
		"norm": []*string{keep(), drop(), s("matrix"), drop(), s("l2"), drop()},
	},

	"mlx_core_random": {
		"categorical": []*string{s("shape"), s("num_samples"), keep()},
		"permutation": []*string{keep(), s("arange")},
		"split":       []*string{s("num"), keep()},
		"uniform":     []*string{keep(), drop(), drop(), drop()},
		"normal":      []*string{s("broadcast"), keep(), drop(), drop(), drop()},
	},

	"mlx_core_fast": {
		"rope": []*string{keep(), s("dynamic")},
	},

	// mlx_core_detail: only these function names are kept; all others are dropped.
	// Handled via the special-case logic in ResolveDetail below.
}

// allowedDetail is the set of function names that pass through mlx::core::detail.
var allowedDetail = map[string]bool{
	"compile":             true,
	"compile_clear_cache": true,
	"compile_erase":       true,
	"vmap_replace":        true,
	"vmap_trace":          true,
}

// ResolveDetail implements the mlx_core_detail namespace logic: only a small
// allow-list of functions are kept; all others return nil.
func ResolveDetail(funcName string, defs []model.Function) []model.Function {
	if !allowedDetail[funcName] {
		return nil
	}
	if len(defs) > 0 {
		return defs[:1]
	}
	return nil
}

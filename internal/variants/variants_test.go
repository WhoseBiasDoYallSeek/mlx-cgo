package variants

import (
	"testing"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/model"
)

// TestResolve_NoTableKeepsFirstOverload verifies default behavior without variant table.
func TestResolve_NoTableKeepsFirstOverload(t *testing.T) {
	funcs := []model.Function{
		{Name: "abs", Namespace: "mlx::core", Params: nil},
		{Name: "abs", Namespace: "mlx::core", Params: nil},
		{Name: "abs", Namespace: "mlx::core", Params: nil},
	}

	result := Resolve("mlx_core", "abs", funcs)

	if len(result) != 1 {
		t.Errorf("Resolve with no variants: expected 1 result, got %d", len(result))
	}
	if result[0].Name != "abs" {
		t.Errorf("Resolve with no variants: expected name 'abs', got %q", result[0].Name)
	}
}

// TestResolve_AppliesSuffixes verifies suffix application for variant overloads.
func TestResolve_AppliesSuffixes(t *testing.T) {
	// squeeze has variants: [s("axes"), s("axis"), keep()]
	// So if we pass 3 overloads, they should get those suffixes
	funcs := []model.Function{
		{Name: "squeeze", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "squeeze", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "squeeze", Namespace: "mlx::core", Params: nil, Variant: ""},
	}

	result := Resolve("mlx_core", "squeeze", funcs)

	// Expected: 3 results with Variant set to "axes", "axis", ""
	if len(result) != 3 {
		t.Errorf("Resolve squeeze: expected 3 results, got %d", len(result))
		return
	}

	expectedVariants := []string{"axes", "axis", ""}
	for i, expected := range expectedVariants {
		if result[i].Variant != expected {
			t.Errorf("squeeze variant[%d]: got %q, expected %q", i, result[i].Variant, expected)
		}
	}
}

// TestResolve_DropsMissingOverloads verifies that overloads marked as nil are dropped.
func TestResolve_DropsMissingOverloads(t *testing.T) {
	// eye has variants: [keep(), drop(), drop(), drop(), drop()]
	// So only the first overload should be kept
	funcs := []model.Function{
		{Name: "eye", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "eye", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "eye", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "eye", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "eye", Namespace: "mlx::core", Params: nil, Variant: ""},
	}

	result := Resolve("mlx_core", "eye", funcs)

	if len(result) != 1 {
		t.Errorf("Resolve eye: expected 1 result (drops 4), got %d", len(result))
	}
}

// TestResolve_LengthMismatch handles overload count mismatch.
func TestResolve_LengthMismatch(t *testing.T) {
	// expand_dims expects [s("axes"), keep()] (2 overloads)
	// But we provide 3, so it should fall back to keeping just the first
	funcs := []model.Function{
		{Name: "expand_dims", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "expand_dims", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "expand_dims", Namespace: "mlx::core", Params: nil, Variant: ""},
	}

	result := Resolve("mlx_core", "expand_dims", funcs)

	// On mismatch, keep only first
	if len(result) != 1 {
		t.Errorf("Resolve on length mismatch: expected 1 result, got %d", len(result))
	}
}

// TestResolve_UnknownFunction returns first only for unknown functions.
func TestResolve_UnknownFunction(t *testing.T) {
	funcs := []model.Function{
		{Name: "unknown_func", Namespace: "mlx::core", Params: nil, Variant: ""},
		{Name: "unknown_func", Namespace: "mlx::core", Params: nil, Variant: ""},
	}

	result := Resolve("mlx_core", "unknown_func", funcs)

	// Unknown function keeps only first
	if len(result) != 1 {
		t.Errorf("Resolve unknown function: expected 1 result, got %d", len(result))
	}
}

// TestResolveDetail_AllowList verifies that only allowlisted functions have detail variants.
func TestResolveDetail_AllowList(t *testing.T) {
	// These functions should be in the allowlist
	allowedFunctions := []string{
		"compile",
		"compile_clear_cache",
		"compile_erase",
		"vmap_replace",
		"vmap_trace",
	}

	for _, fname := range allowedFunctions {
		funcs := []model.Function{
			{Name: fname, Namespace: "mlx::core", Params: nil, Variant: ""},
			{Name: fname, Namespace: "mlx::core", Params: nil, Variant: ""},
		}

		result := ResolveDetail(fname, funcs)
		if result == nil || len(result) != 1 {
			t.Errorf("ResolveDetail(%q): expected 1 result, got %v", fname, result)
		}
	}
}

// TestResolveDetail_NotAllowed verifies that non-allowlisted functions return nil.
func TestResolveDetail_NotAllowed(t *testing.T) {
	// These functions should NOT be in the allowlist
	notAllowedFunctions := []string{
		"add",
		"multiply",
		"abs",
	}

	for _, fname := range notAllowedFunctions {
		funcs := []model.Function{
			{Name: fname, Namespace: "mlx::core", Params: nil, Variant: ""},
			{Name: fname, Namespace: "mlx::core", Params: nil, Variant: ""},
		}

		result := ResolveDetail(fname, funcs)
		if result != nil {
			t.Errorf("ResolveDetail(%q): expected nil, got %v", fname, result)
		}
	}
}

// TestNSKey_Conversion verifies namespace → variant key conversion.
func TestNSKey_Conversion(t *testing.T) {
	testCases := []struct {
		namespace string
		expected  string
	}{
		{"mlx::core", "mlx_core"},
		{"mlx::linalg", "mlx_linalg"},
		{"mlx::random", "mlx_random"},
		{"mlx::fft", "mlx_fft"},
		{"mlx::io", "mlx_io"},
		{"mlx::distributed", "mlx_distributed"},
		{"mlx::transforms", "mlx_transforms"},
		{"mlx::fast", "mlx_fast"},
	}

	for _, tc := range testCases {
		got := NSKey(tc.namespace)
		if got != tc.expected {
			t.Errorf("NSKey(%q): got %q, want %q", tc.namespace, got, tc.expected)
		}
	}
}

// TestVariantTable_CoredHasEntries verifies that mlx_core has variant entries.
func TestVariantTable_CoreHasEntries(t *testing.T) {
	// Test a few known functions that should have variants in mlx_core
	testFunctions := map[string]int{
		"squeeze":     3, // axes, axis, no-arg variants
		"expand_dims": 2, // axes, no-arg variants
		"eye":         5, // multiple overloads
		"arange":      9, // many overloads
	}

	for fname, expectedCount := range testFunctions {
		funcs := make([]model.Function, expectedCount)
		for i := 0; i < expectedCount; i++ {
			funcs[i] = model.Function{Name: fname, Namespace: "mlx::core", Params: nil, Variant: ""}
		}

		result := Resolve("mlx_core", fname, funcs)
		if len(result) == 0 {
			t.Errorf("Resolve %q: expected non-empty result, got empty", fname)
		}
	}
}

// TestResolveDetail_EmptyInput handles empty function list.
func TestResolveDetail_EmptyInput(t *testing.T) {
	result := ResolveDetail("matmul", []model.Function{})
	if result != nil {
		t.Errorf("ResolveDetail with empty input: expected nil, got %v", result)
	}
}

// TestResolve_EmptyInput handles empty function list.
func TestResolve_EmptyInput(t *testing.T) {
	result := Resolve("mlx_core", "add", []model.Function{})
	if result != nil {
		t.Errorf("Resolve with empty input: expected nil, got %v", result)
	}
}

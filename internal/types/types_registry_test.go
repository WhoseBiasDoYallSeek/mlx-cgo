package types

import (
	"testing"
)

// TestFindByCPP_SimpleTypes verifies lookup of basic C++ types.
func TestFindByCPP_SimpleTypes(t *testing.T) {
	tests := []struct {
		cpp       string
		wantCName string
		wantFound bool
	}{
		{"mlx::core::array", "mlx_array", true},
		{"array", "mlx_array", true}, // alt name
		{"bool", "bool", true},
		{"int", "int", true},
		{"float", "float", true},
		{"double", "double", true},
		{"size_t", "size_t", true},
		{"uint64_t", "uint64_t", true},
		{"mlx::core::Stream", "mlx_stream", true},
		{"StreamOrDevice", "mlx_stream", true}, // alt name
		{"std::string", "", true},              // no C name
		{"mlx::core::Dtype", "mlx_dtype", true},
		{"Dtype", "mlx_dtype", true}, // alt name
		{"UnknownType", "", false},
		{"mlx::core::Device", "", false}, // doesn't exist
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpp)
		if found != test.wantFound {
			t.Errorf("FindByCPP(%q): found=%v, want=%v", test.cpp, found, test.wantFound)
		}
		if found && test.wantCName != "" && ti.CName != test.wantCName {
			t.Errorf("FindByCPP(%q): CName=%q, want=%q", test.cpp, ti.CName, test.wantCName)
		}
	}
}

// TestFindByCPP_VectorTypes verifies opaque vector type lookup.
func TestFindByCPP_VectorTypes(t *testing.T) {
	tests := []struct {
		cpp       string
		wantCName string
	}{
		{"std::vector<mlx::core::array>", "mlx_vector_array"},
		{"std::vector<array>", "mlx_vector_array"}, // alt name
		{"std::vector<std::string>", "mlx_vector_string"},
		{"@std::vector<int>", "mlx_vector_int"}, // raw type marker
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpp)
		if !found {
			t.Errorf("FindByCPP(%q): not found", test.cpp)
			continue
		}
		if ti.CName != test.wantCName {
			t.Errorf("FindByCPP(%q): CName=%q, want=%q", test.cpp, ti.CName, test.wantCName)
		}
	}
}

// TestFindByCPP_OptionalTypes verifies optional type unwrapping.
func TestFindByCPP_OptionalTypes(t *testing.T) {
	// OptionalPrimitiveType wraps optional<T> with CName for output
	tests := []struct {
		cpp       string
		wantCName string
	}{
		{"std::optional<mlx::core::array>", "mlx_array"},
		{"std::optional<int>", "mlx_optional_int"},
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpp)
		if !found {
			t.Errorf("FindByCPP(%q): not found", test.cpp)
			continue
		}
		if ti.CName != test.wantCName {
			t.Errorf("FindByCPP(%q): CName=%q, want=%q", test.cpp, ti.CName, test.wantCName)
		}
	}
}

// TestFindByCPP_PairTypes verifies pair type lookup.
func TestFindByCPP_PairTypes(t *testing.T) {
	tests := []struct {
		cpp       string
		wantFound bool
	}{
		{"std::pair<mlx::core::array, mlx::core::array>", true},
		{"std::pair<array, array>", true}, // alt name
	}

	for _, test := range tests {
		_, found := FindByCPP(test.cpp)
		if found != test.wantFound {
			t.Errorf("FindByCPP(%q): found=%v, want=%v", test.cpp, found, test.wantFound)
		}
	}
}

// TestFindByC_Lookup verifies C → C++ reverse lookup.
func TestFindByC_Lookup(t *testing.T) {
	tests := []struct {
		cname     string
		wantFound bool
	}{
		{"mlx_array", true},
		{"mlx_stream", true},
		{"int", true},
		{"mlx_vector_array", true},
		{"mlx_dtype", true},
		{"unknown_type", false},
	}

	for _, test := range tests {
		_, found := FindByC(test.cname)
		if found != test.wantFound {
			t.Errorf("FindByC(%q): found=%v, want=%v", test.cname, found, test.wantFound)
		}
	}
}

// TestAltNames_ShortFormLookup verifies that AltNames enable short-form type lookup.
func TestAltNames_ShortFormLookup(t *testing.T) {
	// "array" is a short form for "mlx::core::array" via AltNames
	ti1, found1 := FindByCPP("array")
	ti2, found2 := FindByCPP("mlx::core::array")

	if !found1 {
		t.Error("Short form 'array' not found")
	}
	if !found2 {
		t.Error("Full form 'mlx::core::array' not found")
	}

	if found1 && found2 && ti1.CName != ti2.CName {
		t.Errorf("Short form 'array' and full form 'mlx::core::array' resolve to different CNames: %q vs %q", ti1.CName, ti2.CName)
	}
}

// TestTypeInfo_CArg verifies C argument generation.
func TestTypeInfo_CArg(t *testing.T) {
	tests := []struct {
		cpptype  string
		varname  string
		expected string
	}{
		{"mlx::core::array", "a", "const mlx_array a"},
		{"int", "x", "int x"},
		{"std::vector<mlx::core::array>", "outputs", "const mlx_vector_array outputs"},
		{"mlx::core::Dtype", "dt", "mlx_dtype dt"},
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpptype)
		if !found {
			t.Errorf("TestTypeInfo_CArg: type %q not found", test.cpptype)
			continue
		}
		if ti.CArg == nil {
			t.Errorf("TestTypeInfo_CArg: type %q has nil CArg", test.cpptype)
			continue
		}
		got := ti.CArg(test.varname)
		if got != test.expected {
			t.Errorf("CArg(%q): got %q, want %q", test.cpptype, got, test.expected)
		}
	}
}

// TestTypeInfo_CReturnArg verifies C return argument generation.
func TestTypeInfo_CReturnArg(t *testing.T) {
	tests := []struct {
		cpptype  string
		varname  string
		expected string
		mayBeNil bool
	}{
		{"mlx::core::array", "res", "mlx_array* res", false},
		{"int", "res", "int* res", false},
		{"bool", "res", "bool* res", false},
		{"std::vector<mlx::core::array>", "res", "mlx_vector_array* res", false},
		{"mlx::core::Dtype", "res", "mlx_dtype* res", false},
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpptype)
		if !found {
			t.Errorf("TestTypeInfo_CReturnArg: type %q not found", test.cpptype)
			continue
		}
		if ti.CReturnArg == nil {
			if !test.mayBeNil {
				t.Errorf("TestTypeInfo_CReturnArg: type %q has nil CReturnArg", test.cpptype)
			}
			continue
		}
		got := ti.CReturnArg(test.varname)
		if got != test.expected {
			t.Errorf("CReturnArg(%q): got %q, want %q", test.cpptype, got, test.expected)
		}
	}
}

// TestTypeInfo_CToCPP verifies C to C++ conversion.
func TestTypeInfo_CToCPP(t *testing.T) {
	tests := []struct {
		cpptype  string
		varname  string
		expected string
	}{
		{"mlx::core::array", "a", "mlx_array_get_(a)"},
		{"int", "x", "x"},
		{"bool", "flag", "flag"},
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpptype)
		if !found {
			t.Errorf("type %q not found", test.cpptype)
			continue
		}
		if ti.CToCPP == nil {
			t.Errorf("type %q has nil CToCPP", test.cpptype)
			continue
		}
		got := ti.CToCPP(test.varname)
		if got != test.expected {
			t.Errorf("CToCPP(%q): got %q, want %q", test.cpptype, got, test.expected)
		}
	}
}

// TestTypeInfo_Free verifies free/cleanup generation.
func TestTypeInfo_Free(t *testing.T) {
	tests := []struct {
		cpptype  string
		varname  string
		expected string
	}{
		{"mlx::core::array", "a", "mlx_array_free(a)"},
		{"int", "x", ""},
		{"mlx::core::Dtype", "dt", ""},
		{"std::vector<mlx::core::array>", "v", "mlx_vector_array_free(v)"},
	}

	for _, test := range tests {
		ti, found := FindByCPP(test.cpptype)
		if !found {
			t.Errorf("type %q not found", test.cpptype)
			continue
		}
		if ti.Free == nil {
			t.Errorf("type %q has nil Free", test.cpptype)
			continue
		}
		got := ti.Free(test.varname)
		if got != test.expected {
			t.Errorf("Free(%q): got %q, want %q", test.cpptype, got, test.expected)
		}
	}
}

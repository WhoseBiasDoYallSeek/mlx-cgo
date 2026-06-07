package parser

import (
	"testing"
)

// TestExtract_BasicFunction verifies basic function extraction from AST.
func TestExtract_BasicFunction(t *testing.T) {
	root := &Node{
		Inner: []Node{
			{
				Kind: "NamespaceDecl",
				Name: "mlx",
				Inner: []Node{
					{
						Kind: "NamespaceDecl",
						Name: "core",
						Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
						Inner: []Node{
							{
								Kind: "FunctionDecl",
								Name: "array",
								Type: astType{QualType: "mlx::core::array (int size)"},
								Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{
									{
										Kind: "ParmVarDecl",
										Name: "size",
										Type: astType{QualType: "int"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	funcs, enums := Extract(root)

	if len(enums) != 0 {
		t.Errorf("expected 0 enums, got %d", len(enums))
	}

	key := "mlx::core::array"
	if len(funcs[key]) != 1 {
		t.Errorf("expected 1 function mlx::core::array, got %d", len(funcs[key]))
	}

	f := funcs[key][0]
	if f.Name != "array" {
		t.Errorf("expected name 'array', got %q", f.Name)
	}
	if f.Namespace != "mlx::core" {
		t.Errorf("expected namespace 'mlx::core', got %q", f.Namespace)
	}
	if len(f.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(f.Params))
	}
	if f.Params[0].Name != "size" || f.Params[0].Type != "int" {
		t.Errorf("expected param (size, int), got (%s, %s)", f.Params[0].Name, f.Params[0].Type)
	}
}

// TestExtract_Overloads verifies that multiple overloads are extracted.
func TestExtract_Overloads(t *testing.T) {
	root := &Node{
		Inner: []Node{
			{
				Kind: "NamespaceDecl",
				Name: "mlx",
				Inner: []Node{
					{
						Kind: "NamespaceDecl",
						Name: "core",
						Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
						Inner: []Node{
							{
								Kind: "FunctionDecl",
								Name: "abs",
								Type: astType{QualType: "mlx::core::array (const mlx::core::array &a)"},
								Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{
									{Kind: "ParmVarDecl", Name: "a", Type: astType{QualType: "const mlx::core::array &"}},
								},
							},
							{
								Kind: "FunctionDecl",
								Name: "abs",
								Type: astType{QualType: "mlx::core::array (const mlx::core::array &a, mlx::core::Stream s)"},
								Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{
									{Kind: "ParmVarDecl", Name: "a", Type: astType{QualType: "const mlx::core::array &"}},
									{Kind: "ParmVarDecl", Name: "s", Type: astType{QualType: "mlx::core::Stream"}},
								},
							},
						},
					},
				},
			},
		},
	}

	funcs, _ := Extract(root)

	key := "mlx::core::abs"
	if len(funcs[key]) != 2 {
		t.Errorf("expected 2 overloads of abs, got %d", len(funcs[key]))
	}

	if len(funcs[key][0].Params) != 1 {
		t.Errorf("expected first overload to have 1 param, got %d", len(funcs[key][0].Params))
	}
	if len(funcs[key][1].Params) != 2 {
		t.Errorf("expected second overload to have 2 params, got %d", len(funcs[key][1].Params))
	}
}

// TestExtract_StdinFilter verifies that functions from transitive includes are filtered.
func TestExtract_StdinFilter(t *testing.T) {
	root := &Node{
		Inner: []Node{
			{
				Kind: "NamespaceDecl",
				Name: "mlx",
				Inner: []Node{
					{
						Kind: "NamespaceDecl",
						Name: "core",
						Inner: []Node{
							{
								Kind:  "FunctionDecl",
								Name:  "from_stdin",
								Type:  astType{QualType: "int ()"},
								Loc:   astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{},
							},
							{
								Kind:  "FunctionDecl",
								Name:  "from_include",
								Type:  astType{QualType: "int ()"},
								Loc:   astLoc{IncludedFrom: &astIncludedFrom{File: "/usr/include/vector.h"}},
								Inner: []Node{},
							},
						},
					},
				},
			},
		},
	}

	funcs, _ := Extract(root)

	if len(funcs["mlx::core::from_stdin"]) != 1 {
		t.Error("expected from_stdin to be extracted")
	}
	if len(funcs["mlx::core::from_include"]) != 0 {
		t.Error("expected from_include to be filtered out")
	}
}

// TestNormalizeType_BuiltinTypes tests normalization of basic types.
func TestNormalizeType_BuiltinTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int", "int"},
		{"float", "float"},
		{"bool", "bool"},
		{"_Bool", "bool"},
		{"size_t", "size_t"},
	}

	for _, test := range tests {
		got := normalizeType(test.input)
		if got != test.expected {
			t.Errorf("normalizeType(%q) = %q, expected %q", test.input, got, test.expected)
		}
	}
}

// TestNormalizeType_Qualifiers tests removal of const and references.
func TestNormalizeType_Qualifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"const int", "int"},
		{"const int &", "int"},
		{"const int &&", "int"},
		{"int &", "int"},
		{"int &&", "int"},
		{"const mlx::core::array &", "mlx::core::array"},
		{"const std::vector<int> &&", "std::vector<int>"},
	}

	for _, test := range tests {
		got := normalizeType(test.input)
		if got != test.expected {
			t.Errorf("normalizeType(%q) = %q, expected %q", test.input, got, test.expected)
		}
	}
}

// TestNormalizeFunctionType_Closures tests normalization of std::function types.
func TestNormalizeFunctionType_Closures(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"std::function<std::vector<array> (const std::vector<array> &)>",
			"std::function<std::vector<array>(std::vector<array>)>",
		},
		{
			"const std::function<int (const int &, const float &)> &",
			"std::function<int(int, float)>",
		},
		{
			"std::optional<std::function<std::pair<array, array> (const array &)>>",
			"std::optional<std::function<std::pair<array, array>(array)>>",
		},
	}

	for _, test := range tests {
		got := normalizeType(test.input)
		if got != test.expected {
			t.Errorf("normalizeType(%q) = %q, expected %q", test.input, got, test.expected)
		}
	}
}

// TestReturnTypeFromQual tests extraction of return type from function qualType.
func TestReturnTypeFromQual(t *testing.T) {
	tests := []struct {
		qualType string
		expected string
	}{
		{"int (int, float)", "int"},
		{"mlx::core::array (const mlx::core::array &)", "mlx::core::array"},
		{"std::vector<int> (const std::vector<int> &)", "std::vector<int>"},
		{
			"std::pair<mlx::core::array, int> (const mlx::core::array &, int idx)",
			"std::pair<mlx::core::array, int>",
		},
	}

	for _, test := range tests {
		got := returnTypeFromQual(test.qualType)
		if got != test.expected {
			t.Errorf("returnTypeFromQual(%q) = %q, expected %q", test.qualType, got, test.expected)
		}
	}
}

// TestJoinNS tests namespace joining.
func TestJoinNS(t *testing.T) {
	tests := []struct {
		ns       string
		name     string
		expected string
	}{
		{"", "array", "array"},
		{"mlx", "core", "mlx::core"},
		{"mlx::core", "array", "mlx::core::array"},
		{"mlx::core::detail", "compile", "mlx::core::detail::compile"},
	}

	for _, test := range tests {
		got := joinNS(test.ns, test.name)
		if got != test.expected {
			t.Errorf("joinNS(%q, %q) = %q, expected %q", test.ns, test.name, got, test.expected)
		}
	}
}

// TestExtract_Enum verifies enum extraction.
func TestExtract_Enum(t *testing.T) {
	root := &Node{
		Inner: []Node{
			{
				Kind: "NamespaceDecl",
				Name: "mlx",
				Inner: []Node{
					{
						Kind: "NamespaceDecl",
						Name: "core",
						Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
						Inner: []Node{
							{
								Kind: "EnumDecl",
								Name: "Dtype",
								Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{
									{Kind: "EnumConstantDecl", Name: "float32"},
									{Kind: "EnumConstantDecl", Name: "float64"},
									{Kind: "EnumConstantDecl", Name: "int32"},
								},
							},
						},
					},
				},
			},
		},
	}

	_, enums := Extract(root)

	key := "mlx::core::Dtype"
	if len(enums) != 1 {
		t.Errorf("expected 1 enum, got %d", len(enums))
	}

	dtype := enums[key]
	if dtype.Name != "Dtype" {
		t.Errorf("expected name 'Dtype', got %q", dtype.Name)
	}
	if len(dtype.Values) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(dtype.Values))
	}
	expected := []string{"float32", "float64", "int32"}
	for i, v := range dtype.Values {
		if i < len(expected) && v != expected[i] {
			t.Errorf("enum value %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

// TestExtract_DefaultParameters verifies handling of default parameters.
func TestExtract_DefaultParameters(t *testing.T) {
	root := &Node{
		Inner: []Node{
			{
				Kind: "NamespaceDecl",
				Name: "mlx",
				Inner: []Node{
					{
						Kind: "NamespaceDecl",
						Name: "core",
						Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
						Inner: []Node{
							{
								Kind: "FunctionDecl",
								Name: "someFunc",
								Type: astType{QualType: "int (int x, int y)"},
								Loc:  astLoc{IncludedFrom: &astIncludedFrom{File: "<stdin>"}},
								Inner: []Node{
									{Kind: "ParmVarDecl", Name: "x", Type: astType{QualType: "int"}},
									{Kind: "ParmVarDecl", Name: "y", Type: astType{QualType: "int"}, Init: "5"},
								},
							},
						},
					},
				},
			},
		},
	}

	funcs, _ := Extract(root)

	f := funcs["mlx::core::someFunc"][0]
	if f.Defaults[0] != "" {
		t.Errorf("expected x to have no default, got %q", f.Defaults[0])
	}
	if f.Defaults[1] != "<default>" {
		t.Errorf("expected y to have default marker, got %q", f.Defaults[1])
	}
}

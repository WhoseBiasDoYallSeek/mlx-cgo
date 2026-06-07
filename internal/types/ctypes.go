// Package types provides the type mapping registry that translates between
// C++ MLX types, C binding types, and (later) Go/CGo types.
//
// This is a direct port of python/mlxtypes.py from the original mlx-c project.
package types

// TypeInfo describes how a single C++ type is represented at the C boundary.
//
// Each field is a function that receives a variable name and returns a
// snippet of C/C++ source code. This mirrors the lambda-based design of the
// original Python implementation.
type TypeInfo struct {
	// CName is the C typedef name (e.g. "mlx_array"). May be empty for
	// types that have no standalone C name (e.g. raw-vector in-params).
	CName string
	// CPPName is the canonical fully-qualified C++ type string used as the
	// registry key (e.g. "mlx::core::array").
	CPPName string
	// AltNames are additional C++ spellings that map to this TypeInfo
	// (e.g. "array", "StreamOrDevice"). May be nil.
	AltNames []string

	// CArg returns the C declaration for a function parameter.
	//   e.g. CArg("a") → "const mlx_array a"
	CArg func(name string) string

	// CReturnArg returns the C declaration for the output (return) pointer param.
	//   e.g. CReturnArg("res") → "mlx_array* res"
	// Returns "" for void return types.
	CReturnArg func(name string) string

	// CToCPP returns a C++ expression that converts a C value to its C++ form.
	//   e.g. CToCPP("a") → "mlx_array_get_(a)"
	CToCPP func(name string) string

	// CAssignFromCPP returns C code that stores a C++ expression result into
	// a C out-param.
	//   e.g. CAssignFromCPP("res", "mlx::core::abs(...)") →
	//        "mlx_array_set_(*res, mlx::core::abs(...))"
	CAssignFromCPP func(dest, src string) string

	// CNew returns a declaration for a local C variable used in the
	// return-value pattern.
	//   e.g. CNew("res") → "auto res = mlx_array_new_()"
	// May be nil for types that are always passed by pointer out-param.
	CNew func(name string) string

	// Free returns a statement that frees a C value.
	// Returns "" for types that need no explicit free.
	Free func(name string) string
}

// opaqueType builds a TypeInfo for mlx opaque handle types (mlx_array,
// mlx_stream, etc.) that use the ctx-pointer pattern.
func opaqueType(cname, cppname string, alts ...string) TypeInfo {
	return TypeInfo{
		CName:      cname,
		CPPName:    cppname,
		AltNames:   alts,
		Free:       func(s string) string { return cname + "_free(" + s + ")" },
		CToCPP:     func(s string) string { return cname + "_get_(" + s + ")" },
		CArg:       func(s string) string { return "const " + cname + " " + s },
		CReturnArg: func(s string) string { return cname + "* " + s },
		CAssignFromCPP: func(d, src string) string {
			return cname + "_set_(*" + d + ", " + src + ")"
		},
		CNew: func(s string) string { return "auto " + s + " = " + cname + "_new_()" },
	}
}

// primitiveType builds a TypeInfo for plain C primitive types (int, float,
// bool, etc.) that are passed by value and returned via pointer out-param.
func primitiveType(name string, alts ...string) TypeInfo {
	return TypeInfo{
		CName:      name,
		CPPName:    name,
		AltNames:   alts,
		Free:       func(string) string { return "" },
		CToCPP:     func(s string) string { return s },
		CArg:       func(s string) string { return name + " " + s },
		CReturnArg: func(s string) string { return name + "* " + s },
		CAssignFromCPP: func(d, src string) string {
			return "*" + d + " = " + src
		},
		CNew: func(s string) string { return name + " " + s },
	}
}

// rawVectorType builds a TypeInfo for std::vector<T> types that are passed
// as a (const T* ptr, size_t num) pair at the C boundary.
func rawVectorType(elemCPP string) TypeInfo {
	cpptype := "std::vector<" + elemCPP + ">"
	return TypeInfo{
		CPPName: cpptype,
		Free:    func(string) string { return "" },
		CToCPP: func(s string) string {
			return "std::vector<" + elemCPP + ">(" + s + ", " + s + " + " + s + "_num)"
		},
		CArg: func(s string) string {
			return "const " + elemCPP + "* " + s + ", size_t " + s + "_num"
		},
		CNew: func(s string) string {
			return "const " + elemCPP + "* " + s + " = nullptr; size_t " + s + "_num = 0"
		},
		CAssignFromCPP: func(d, src string) string {
			return d + " = " + src + ".data(); " + d + "_num = " + src + ".size()"
		},
	}
}

// smallVectorType builds a TypeInfo for named small-vector types like
// mlx::core::Shape (which is really a std::vector<int> underneath) passed
// the same way as rawVectorType.
func smallVectorType(elemC, cpptype string, alts ...string) TypeInfo {
	t := rawVectorType(elemC)
	t.CPPName = cpptype
	t.AltNames = alts
	t.CToCPP = func(s string) string {
		return cpptype + "(" + s + ", " + s + " + " + s + "_num)"
	}
	t.CArg = func(s string) string {
		return "const " + elemC + "* " + s + ", size_t " + s + "_num"
	}
	return t
}

// optionalRawVectorType builds a TypeInfo for std::optional<std::vector<T>>.
func optionalRawVectorType(elemCPP string) TypeInfo {
	cpp := "std::optional<std::vector<" + elemCPP + ">>"
	return TypeInfo{
		CPPName: cpp,
		Free:    func(string) string { return "" },
		CToCPP: func(s string) string {
			return "(" + s + " ? std::make_optional(std::vector<" + elemCPP +
				">(" + s + ", " + s + " + " + s + "_num)) : std::nullopt)"
		},
		CAssignFromCPP: func(d, src string) string {
			return "if(" + src + ".has_value()) { " +
				d + " = " + src + ".value().data(); " +
				d + "_num = " + src + ".value().size(); " +
				"} else { " + d + " = nullptr; " + d + "_num = 0; }"
		},
		CArg: func(s string) string {
			return "const " + elemCPP + "* " + s + " /* may be null */, size_t " + s + "_num"
		},
		CNew: func(s string) string {
			return "const " + elemCPP + "* " + s + " = nullptr; size_t " + s + "_num = 0"
		},
	}
}

// pairType builds a TypeInfo for std::pair<A,B> return types that are split
// into two separate C out-params.
func pairType(cppA, cppB, cA, cB string) TypeInfo {
	cpp := "std::pair<" + cppA + ", " + cppB + ">"
	return TypeInfo{
		CPPName: cpp,
		Free: func(s string) string {
			return cA + "_free(" + s + "_0);\n" + cB + "_free(" + s + "_1);"
		},
		CReturnArg: func(s string) string {
			if s == "" {
				return cA + "*, " + cB + "*"
			}
			return cA + "* " + s + "_0, " + cB + "* " + s + "_1"
		},
		CToCPP: func(s string) string {
			return "std::make_pair(" + cA + "_get_(" + s + "_0), " + cB + "_get_(" + s + "_1))"
		},
		CNew: func(s string) string {
			return "auto " + s + "_0 = " + cA + "_new_();\nauto " + s + "_1 = " + cB + "_new_()"
		},
		CAssignFromCPP: func(d, src string) string {
			return "{ auto [tpl_0, tpl_1] = " + src + "; " +
				cA + "_set_(*" + d + "_0, tpl_0); " +
				cB + "_set_(*" + d + "_1, tpl_1); }"
		},
	}
}

// tripleType builds a TypeInfo for std::tuple<A,B,C> return types.
func tripleType(cppA, cppB, cppC, cA, cB, cC string) TypeInfo {
	cpp := "std::tuple<" + cppA + ", " + cppB + ", " + cppC + ">"
	return TypeInfo{
		CPPName: cpp,
		Free: func(s string) string {
			return cA + "_free(" + s + "_0);\n" + cB + "_free(" + s + "_1);\n" + cC + "_free(" + s + "_2);"
		},
		CReturnArg: func(s string) string {
			if s == "" {
				return cA + "*, " + cB + "*, " + cC + "*"
			}
			return cA + "* " + s + "_0, " + cB + "* " + s + "_1, " + cC + "* " + s + "_2"
		},
		CNew: func(s string) string {
			return "auto " + s + "_0 = " + cA + "_new_();\nauto " + s + "_1 = " + cB + "_new_();\nauto " + s + "_2 = " + cC + "_new_()"
		},
		CAssignFromCPP: func(d, src string) string {
			return "{ auto [tpl_0, tpl_1, tpl_2] = " + src + "; " +
				cA + "_set_(*" + d + "_0, tpl_0); " +
				cB + "_set_(*" + d + "_1, tpl_1); " +
				cC + "_set_(*" + d + "_2, tpl_2); }"
		},
	}
}

// optionalOpaqueType wraps an opaque TypeInfo in std::optional<T>.
func optionalOpaqueType(base TypeInfo) TypeInfo {
	opt := TypeInfo{
		CName:    base.CName,
		CPPName:  "std::optional<" + base.CPPName + ">",
		AltNames: nil,
		Free:     base.Free,
		CArg: func(s string) string {
			return base.CArg(s) + " /* may be null */"
		},
		CToCPP: func(s string) string {
			return "(" + s + ".ctx ? std::make_optional(" + base.CToCPP(s) + ") : std::nullopt)"
		},
		CAssignFromCPP: func(d, src string) string {
			return "(" + src + ".has_value() ? " + src + ".value() : nullptr)"
		},
		CReturnArg: base.CReturnArg,
		CNew:       base.CNew,
	}
	if len(base.AltNames) > 0 {
		opt.AltNames = append(opt.AltNames, "std::optional<"+base.AltNames[0]+">")
	}
	return opt
}

// optionalPrimitiveType wraps a primitive TypeInfo in std::optional<T>.
func optionalPrimitiveType(base TypeInfo) TypeInfo {
	optCName := "mlx_optional_" + stripMLXPrefix(base.CName)
	optCPP := "std::optional<" + base.CPPName + ">"
	toCPP := base.CToCPP
	// Build AltNames: "std::optional<shortAlias>" for each AltName of the base.
	var altNames []string
	for _, alt := range base.AltNames {
		altNames = append(altNames, "std::optional<"+alt+">")
	}
	return TypeInfo{
		CName:    optCName,
		CPPName:  optCPP,
		AltNames: altNames,
		Free:     func(string) string { return "" },
		CArg:     func(s string) string { return optCName + " " + s },
		CToCPP: func(s string) string {
			return "(" + s + ".has_value ? std::make_optional<" + base.CPPName +
				">(" + toCPP(s+".value") + ") : std::nullopt)"
		},
	}
}

func stripMLXPrefix(s string) string {
	if len(s) > 4 && s[:4] == "mlx_" {
		return s[4:]
	}
	return s
}

// All registered types. Order mirrors mlxtypes.py.
var registry []TypeInfo

func init() {
	// --- Opaque MLX handle types ---
	registry = append(registry,
		opaqueType("mlx_array", "mlx::core::array", "array"),
		opaqueType("mlx_stream", "mlx::core::Stream", "StreamOrDevice"),
		opaqueType("mlx_distributed_group", "mlx::core::distributed::Group", "Group"),
		opaqueType("mlx_node_namer", "mlx::core::NodeNamer", "NodeNamer"),
	)

	// mlx_stream without alt (second entry for the bare Stream match)
	registry = append(registry, TypeInfo{
		CName:      "mlx_stream",
		CPPName:    "mlx::core::Stream",
		Free:       func(s string) string { return "mlx_stream_free(" + s + ")" },
		CToCPP:     func(s string) string { return "mlx_stream_get_(" + s + ")" },
		CArg:       func(s string) string { return "const mlx_stream " + s },
		CReturnArg: func(s string) string { return "mlx_stream* " + s },
		CAssignFromCPP: func(d, src string) string {
			return "mlx_stream_set_(*" + d + ", " + src + ")"
		},
		CNew: func(s string) string { return "auto " + s + " = mlx_stream_new_()" },
	})

	// --- Opaque closure/function types ---
	registry = append(registry,
		opaqueType("mlx_closure",
			"std::function<std::vector<array>(std::vector<array>)>"),
		opaqueType("mlx_closure_value_and_grad",
			"std::function<std::pair<std::vector<array>, std::vector<array>>(const std::vector<array>&)>",
			"ValueAndGradFn"),
		opaqueType("mlx_closure_custom",
			"std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<mlx::core::array>)>",
			"std::function<std::vector<array>(std::vector<array>,std::vector<array>,std::vector<array>)>"),
		opaqueType("mlx_closure_custom_jvp",
			"std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<int>)>",
			"std::function<std::vector<array>(std::vector<array>,std::vector<array>,std::vector<int>)>"),
		opaqueType("mlx_closure_custom_vmap",
			"std::function<std::pair<std::vector<mlx::core::array>, std::vector<int>>(std::vector<mlx::core::array>,std::vector<int>)>",
			"std::function<std::pair<std::vector<array>, std::vector<int>>(std::vector<array>,std::vector<int>)>"),
	)

	// --- Opaque vector types (returned as MLX handles) ---
	registry = append(registry,
		opaqueType("mlx_vector_int", "@std::vector<int>", "@std::vector<int>"),
		opaqueType("mlx_vector_string", "std::vector<std::string>", "std::vector<std::string>"),
		opaqueType("mlx_vector_array", "std::vector<mlx::core::array>", "std::vector<array>"),
	)

	// --- Map types ---
	registry = append(registry,
		opaqueType("mlx_map_string_to_array",
			"std::unordered_map<std::string, mlx::core::array>",
			"std::unordered_map<std::string, array>"),
		opaqueType("mlx_map_string_to_string",
			"std::unordered_map<std::string, std::string>",
			"std::unordered_map<std::string, std::string>"),
	)

	// --- Small-vector (in-param) types ---
	registry = append(registry,
		smallVectorType("int", "mlx::core::Shape", "Shape"),
		smallVectorType("int64_t", "mlx::core::Strides", "Strides"),
	)

	// --- Raw in-param vector types ---
	registry = append(registry,
		rawVectorType("int"),
		rawVectorType("size_t"),
		rawVectorType("uint64_t"),
	)

	// --- Optional raw vector ---
	registry = append(registry, optionalRawVectorType("int"))

	// --- Pair / tuple return types ---
	p1 := pairType("mlx::core::array", "mlx::core::array", "mlx_array", "mlx_array")
	p1.AltNames = []string{"std::pair<array, array>"}
	registry = append(registry, p1)

	triple1 := tripleType("mlx::core::array", "mlx::core::array", "mlx::core::array",
		"mlx_array", "mlx_array", "mlx_array")
	triple1.AltNames = []string{"std::tuple<array, array, array>"}
	registry = append(registry, triple1)

	p2 := pairType("std::vector<mlx::core::array>", "std::vector<mlx::core::array>",
		"mlx_vector_array", "mlx_vector_array")
	p2.AltNames = []string{"std::pair<std::vector<array>, std::vector<array>>"}
	registry = append(registry, p2)

	p3 := pairType("std::vector<mlx::core::array>", "@std::vector<int>",
		"mlx_vector_array", "mlx_vector_int")
	p3.AltNames = []string{"std::pair<std::vector<array>, std::vector<int>>"}
	registry = append(registry, p3)

	// SafetensorsLoad pair
	st := pairType(
		"std::unordered_map<std::string, mlx::core::array>",
		"std::unordered_map<std::string, std::string>",
		"mlx_map_string_to_array",
		"mlx_map_string_to_string",
	)
	st.AltNames = []string{"SafetensorsLoad"}
	registry = append(registry, st)

	// --- void ---
	registry = append(registry, TypeInfo{
		CPPName:        "void",
		CReturnArg:     func(string) string { return "" },
		CAssignFromCPP: func(_, src string) string { return src },
	})

	// --- mlx::core::Dtype ---
	registry = append(registry, TypeInfo{
		CName:      "mlx_dtype",
		CPPName:    "mlx::core::Dtype",
		AltNames:   []string{"Dtype"},
		Free:       func(string) string { return "" },
		CToCPP:     func(s string) string { return "mlx_dtype_to_cpp(" + s + ")" },
		CArg:       func(s string) string { return "mlx_dtype " + s },
		CReturnArg: func(s string) string { return "mlx_dtype* " + s },
		CNew:       func(s string) string { return "mlx_dtype " + s },
		CAssignFromCPP: func(d, src string) string {
			return d + " = mlx_dtype_to_c((int)((" + src + ").val))"
		},
	})

	// --- mlx::core::CompileMode ---
	registry = append(registry, TypeInfo{
		CPPName:    "mlx::core::CompileMode",
		AltNames:   []string{"CompileMode"},
		Free:       func(string) string { return "" },
		CToCPP:     func(s string) string { return "mlx_compile_mode_to_cpp(" + s + ")" },
		CArg:       func(s string) string { return "mlx_compile_mode " + s },
		CReturnArg: func(s string) string { return "mlx_compile_mode* " + s },
		CNew:       func(s string) string { return "mlx_compile_mode " + s },
		CAssignFromCPP: func(d, src string) string {
			return d + " = mlx_compile_mode_to_c((int)((" + src + ").val))"
		},
	})

	// --- mlx::core::fft::FFTNorm ---
	registry = append(registry, TypeInfo{
		CName:      "mlx_fft_norm",
		CPPName:    "mlx::core::fft::FFTNorm",
		AltNames:   []string{"FFTNorm"},
		Free:       func(string) string { return "" },
		CToCPP:     func(s string) string { return "mlx_fft_norm_to_cpp(" + s + ")" },
		CArg:       func(s string) string { return "mlx_fft_norm " + s },
		CReturnArg: func(s string) string { return "mlx_fft_norm* " + s },
		CNew:       func(s string) string { return "mlx_fft_norm " + s },
		CAssignFromCPP: func(d, src string) string {
			return d + " = mlx_fft_norm_to_c(" + src + ")"
		},
	})

	// --- std::string (const char* at C boundary) ---
	registry = append(registry, TypeInfo{
		CPPName:    "std::string",
		AltNames:   []string{"std::string"},
		CToCPP:     func(s string) string { return "std::string(" + s + ")" },
		CArg:       func(s string) string { return "const char* " + s },
		CReturnArg: func(s string) string { return "char** " + s },
		CAssignFromCPP: func(d, src string) string {
			return d + " = " + src + ".c_str()"
		},
	})

	// --- IO reader / writer ---
	registry = append(registry, TypeInfo{
		CPPName: "std::shared_ptr<io::Reader>",
		CToCPP:  func(s string) string { return "mlx_io_reader_get_(" + s + ")" },
		CArg:    func(s string) string { return "mlx_io_reader " + s },
	})
	registry = append(registry, TypeInfo{
		CPPName: "std::shared_ptr<io::Writer>",
		CToCPP:  func(s string) string { return "mlx_io_writer_get_(" + s + ")" },
		CArg:    func(s string) string { return "mlx_io_writer " + s },
	})

	// --- std::ostream (FILE* at C boundary) ---
	registry = append(registry, TypeInfo{
		CPPName: "std::ostream",
		CToCPP: func(s string) string {
			return "CFileOutputStream::as_lvalue(CFileOutputStream(" + s + "))"
		},
		CArg: func(s string) string { return "FILE* " + s },
	})

	// --- Primitive types ---
	for _, name := range []string{"int", "size_t", "float", "double", "bool", "uint64_t", "uintptr_t"} {
		registry = append(registry, primitiveType(name))
	}
	// uintptr_t has an alt spelling
	registry[len(registry)-1].AltNames = []string{"std::uintptr_t"}

	// --- std::pair<int,int> ---
	registry = append(registry, TypeInfo{
		CPPName:  "std::pair<int, int>",
		AltNames: []string{"std::pair<int, int>"},
		CToCPP: func(s string) string {
			return "std::make_pair(" + s + "_0, " + s + "_1)"
		},
		CArg:       func(s string) string { return "int " + s + "_0, int " + s + "_1" },
		CReturnArg: func(s string) string { return "int* " + s + "_0, int* " + s + "_1" },
		CAssignFromCPP: func(d, src string) string {
			return "std::tie(" + d + "_0, " + d + "_1) = " + src
		},
	})

	// --- std::tuple<int,int,int> ---
	registry = append(registry, TypeInfo{
		CPPName:  "std::tuple<int, int, int>",
		AltNames: []string{"std::tuple<int, int, int>"},
		CToCPP: func(s string) string {
			return "std::make_tuple(" + s + "_0, " + s + "_1, " + s + "_2)"
		},
		CArg: func(s string) string {
			return "int " + s + "_0, int " + s + "_1, int " + s + "_2"
		},
		CReturnArg: func(s string) string {
			return "int* " + s + "_0, int* " + s + "_1, int* " + s + "_2"
		},
		CAssignFromCPP: func(d, src string) string {
			return "std::tie(" + d + "_0, " + d + "_1, " + d + "_2) = " + src
		},
	})

	// --- Optional wrappers ---
	arrayT, _ := FindByCPP("mlx::core::array")
	registry = append(registry, optionalOpaqueType(arrayT))

	groupT, _ := FindByCPP("mlx::core::distributed::Group")
	registry = append(registry, optionalOpaqueType(groupT))

	// optional primitive: float, int, mlx::core::Dtype
	for _, cpp := range []string{"float", "int", "mlx::core::Dtype"} {
		t, ok := FindByCPP(cpp)
		if ok {
			registry = append(registry, optionalPrimitiveType(t))
		}
	}

	// optional closures
	for _, cpp := range []string{
		"std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<mlx::core::array>)>",
		"std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<int>)>",
		"std::function<std::pair<std::vector<mlx::core::array>, std::vector<int>>(std::vector<mlx::core::array>,std::vector<int>)>",
	} {
		t, ok := FindByCPP(cpp)
		if ok {
			registry = append(registry, optionalOpaqueType(t))
		}
	}
}

// FindByCPP returns the TypeInfo whose CPPName or AltName matches cpptype.
func FindByCPP(cpptype string) (TypeInfo, bool) {
	for _, t := range registry {
		if t.CPPName == cpptype {
			return t, true
		}
		for _, alt := range t.AltNames {
			if alt == cpptype {
				return t, true
			}
		}
	}
	return TypeInfo{}, false
}

// FindByC returns the TypeInfo whose CName matches cname.
func FindByC(cname string) (TypeInfo, bool) {
	for _, t := range registry {
		if t.CName == cname {
			return t, true
		}
	}
	return TypeInfo{}, false
}

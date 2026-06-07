package types

import "unicode"

// GoTypeInfo describes how to map a C boundary type to idiomatic Go.
type GoTypeInfo struct {
	// GoType is the idiomatic Go type (e.g. "Array", "int32", "[]Array").
	GoType string
	// CGo is the raw CGo type as it appears in import "C" scope
	// (e.g. "C.mlx_array", "C.int").
	CGo string
	// ToGo converts a CGo expression to its Go form.
	//   e.g. ToGo("cv") → "newArray(cv)"
	ToGo func(expr string) string
	// ToCGo converts a Go expression to its CGo form.
	//   e.g. ToCGo("a") → "a.c()"
	ToCGo func(expr string) string
	// ReturnArg: if set, the function uses an out-param pattern and this is
	// how the CGo out-var is declared, e.g. "var _res C.mlx_array".
	ReturnVar func(varName string) string
	// FreeExpr: expression to free a CGo return var after use (may be "").
	FreeExpr func(varName string) string
}

// GoRegistry maps C type names (as produced by TypeInfo.CName) to GoTypeInfo.
var GoRegistry = buildGoRegistry()

func buildGoRegistry() map[string]GoTypeInfo {
	r := map[string]GoTypeInfo{}

	// ---- opaque handle helpers -----------------------------------------------
	opaque := func(cname, goType string) GoTypeInfo {
		return GoTypeInfo{
			GoType: goType,
			CGo:    "C." + cname,
			ToGo: func(e string) string {
				return goType + "{" + e + "}"
			},
			ToCGo: func(e string) string {
				return e + ".c"
			},
			ReturnVar: func(v string) string {
				return "var " + v + " C." + cname
			},
			FreeExpr: func(v string) string {
				return "C." + cname + "_free(" + v + ")"
			},
		}
	}

	// opaque MLX handle types
	handles := []struct{ c, go_ string }{
		{"mlx_array", "Array"},
		{"mlx_stream", "Stream"},
		{"mlx_device", "Device"},
		{"mlx_closure", "Closure"},
		{"mlx_closure_kwargs", "ClosureKwargs"},
		{"mlx_closure_value_and_grad", "ClosureValueAndGrad"},
		{"mlx_closure_custom", "ClosureCustom"},
		{"mlx_closure_custom_vjp", "ClosureCustomVjp"},
		{"mlx_closure_custom_jvp", "ClosureCustomJvp"},
		{"mlx_vector_array", "VectorArray"},
		{"mlx_map_string_to_array", "MapStringToArray"},
		{"mlx_map_string_to_string", "MapStringToString"},
		{"mlx_optional_array_dtype", "OptionalDtype"},
		{"mlx_distributed_group", "DistributedGroup"},
		{"mlx_compile_fun", "CompileFun"},
		{"mlx_string", "String"},
		{"mlx_io_key_value_iterator", "IOKeyValueIterator"},
	}
	for _, h := range handles {
		r[h.c] = opaque(h.c, h.go_)
	}

	// ---- primitive types -----------------------------------------------------
	prim := func(cname, goType, cgoType string) GoTypeInfo {
		return GoTypeInfo{
			GoType: goType,
			CGo:    cgoType,
			ToGo: func(e string) string {
				return goType + "(" + e + ")"
			},
			ToCGo: func(e string) string {
				return cgoType + "(" + e + ")"
			},
			ReturnVar: func(v string) string {
				return "var " + v + " " + cgoType
			},
			FreeExpr: func(string) string { return "" },
		}
	}

	r["int"] = prim("int", "int", "C.int")
	r["float"] = prim("float", "float32", "C.float")
	r["double"] = prim("double", "float64", "C.double")
	r["bool"] = prim("bool", "bool", "C.bool")
	r["uint32_t"] = prim("uint32_t", "uint32", "C.uint32_t")
	r["uint64_t"] = prim("uint64_t", "uint64", "C.uint64_t")
	r["int32_t"] = prim("int32_t", "int32", "C.int32_t")
	r["int64_t"] = prim("int64_t", "int64", "C.int64_t")
	r["size_t"] = prim("size_t", "uint64", "C.size_t")
	r["uintptr_t"] = prim("uintptr_t", "uintptr", "C.uintptr_t")
	r["mlx_dtype"] = prim("mlx_dtype", "Dtype", "C.mlx_dtype")

	// string (char*) — Go string ↔ C string via C.CString / C.GoString
	r["const char*"] = GoTypeInfo{
		GoType: "string",
		CGo:    "*C.char",
		ToGo: func(e string) string {
			return "C.GoString(" + e + ")"
		},
		ToCGo: func(e string) string {
			return "C.CString(" + e + ")"
		},
		ReturnVar: func(v string) string {
			return "var " + v + " *C.char"
		},
		FreeExpr: func(v string) string {
			return "C.free(unsafe.Pointer(" + v + "))"
		},
	}

	// raw vector (ptr+len) — []T pairs
	rawVec := func(cname, elemGo, elemCGo string) GoTypeInfo {
		return GoTypeInfo{
			GoType: "[]" + elemGo,
			CGo:    "*" + elemCGo,
			ToGo: func(e string) string {
				// returned as C array — convert via unsafe.Slice
				return "cSliceTo" + titleCase(elemGo) + "(" + e + ")"
			},
			ToCGo: func(e string) string {
				return "&" + e + "[0]"
			},
			ReturnVar: func(v string) string {
				return "var " + v + " *" + elemCGo + "; var " + v + "Num C.size_t"
			},
			FreeExpr: func(string) string { return "" },
		}
	}
	r["const int*"] = rawVec("int*", "int32", "C.int")
	r["const float*"] = rawVec("float*", "float32", "C.float")
	r["const double*"] = rawVec("double*", "float64", "C.double")
	r["const bool*"] = rawVec("bool*", "bool", "C.bool")

	return r
}

// FindGoType looks up a GoTypeInfo by CName.
func FindGoType(cname string) (GoTypeInfo, bool) {
	g, ok := GoRegistry[cname]
	return g, ok
}

// titleCase uppercases the first rune of s. Replaces deprecated strings.Title.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

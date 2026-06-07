package codegen

import (
	"fmt"
	"io"
	"strings"
)

// VectorTypeSpec describes a single vector type to be generated.
type VectorTypeSpec struct {
	// CPPType is the C++ element type, e.g. "mlx::core::array"
	CPPType string
	// CType is the C element type used in the public API, e.g. "const mlx_array"
	CType string
	// ShortName is the suffix used in the generated names, e.g. "array"
	ShortName string
	// ReturnCType is the out-param type for the get() function, e.g. "mlx_array*"
	ReturnCType string
	// CToCPP converts a C name to its C++ equivalent for use in push_back etc.
	CToCPP func(name string) string
	// CAssign assigns a C++ value back to a C out-param.
	CAssign func(dest, src string) string
}

// defaultVectorTypes returns the canonical set of vector types matching what
// the Python vector_generator.py generates.
func defaultVectorTypes() []VectorTypeSpec {
	return []VectorTypeSpec{
		{
			CPPType:     "mlx::core::array",
			CType:       "const mlx_array",
			ShortName:   "array",
			ReturnCType: "mlx_array*",
			CToCPP:      func(s string) string { return "mlx_array_get_(" + s + ")" },
			CAssign:     func(d, s string) string { return "mlx_array_set_(*" + d + ", " + s + ")" },
		},
		{
			CPPType:     "std::vector<mlx::core::array>",
			CType:       "const mlx_vector_array",
			ShortName:   "vector_array",
			ReturnCType: "mlx_vector_array*",
			CToCPP:      func(s string) string { return "mlx_vector_array_get_(" + s + ")" },
			CAssign:     func(d, s string) string { return "mlx_vector_array_set_(*" + d + ", " + s + ")" },
		},
		{
			CPPType:     "int",
			CType:       "int",
			ShortName:   "int",
			ReturnCType: "int*",
			CToCPP:      func(s string) string { return s },
			CAssign:     func(d, s string) string { return "*" + d + " = " + s },
		},
		{
			CPPType:     "std::string",
			CType:       "const char*",
			ShortName:   "string",
			ReturnCType: "char**",
			CToCPP:      func(s string) string { return s },
			CAssign:     func(d, s string) string { return "*" + d + " = " + s + ".data()" },
		},
	}
}

// GenerateVectorHeader writes the mlx/c/vector.h file.
func GenerateVectorHeader(w io.Writer) {
	fmt.Fprint(w, vectorDeclBegin)
	for _, t := range defaultVectorTypes() {
		fmt.Fprint(w, generateVectorDecl(t))
	}
	fmt.Fprint(w, vectorDeclEnd)
}

// GenerateVectorImpl writes the mlx/c/vector.cpp file.
func GenerateVectorImpl(w io.Writer) {
	fmt.Fprint(w, vectorImplBegin)
	for _, t := range defaultVectorTypes() {
		fmt.Fprint(w, generateVectorImpl(t))
	}
	fmt.Fprint(w, vectorImplEnd)
}

// GenerateVectorPrivate writes the mlx/c/private/vector.h file.
func GenerateVectorPrivate(w io.Writer) {
	fmt.Fprint(w, vectorPrivBegin)
	for _, t := range defaultVectorTypes() {
		cname := "mlx_vector_" + t.ShortName
		cpptype := "std::vector<" + t.CPPType + ">"
		fmt.Fprint(w, privateCode(cname, cpptype, false, ""))
	}
	fmt.Fprint(w, vectorPrivEnd)
}

func generateVectorDecl(t VectorTypeSpec) string {
	n := "mlx_vector_" + t.ShortName
	ctype := t.CType
	rctype := t.ReturnCType
	var sb strings.Builder
	p := func(f string, args ...any) { fmt.Fprintf(&sb, f+"\n", args...) }

	p(``)
	p(`/**`)
	p(` * A vector of %s.`, t.ShortName)
	p(` */`)
	p(`typedef struct %s_ {`, n)
	p(`  void* ctx;`)
	p(`} %s;`, n)
	p(`%s %s_new(void);`, n, n)
	p(`int %s_set(%s* vec, const %s src);`, n, n, n)
	p(`int %s_free(%s vec);`, n, n)
	p(`%s %s_new_data(%s* data, size_t size);`, n, n, ctype)
	p(`%s %s_new_value(%s val);`, n, n, ctype)
	p(`int %s_set_data(`, n)
	p(`    %s* vec,`, n)
	p(`    %s* data,`, ctype)
	p(`    size_t size);`)
	p(`int %s_set_value(%s* vec, %s val);`, n, n, ctype)
	p(`int %s_append_data(`, n)
	p(`    %s vec,`, n)
	p(`    %s* data,`, ctype)
	p(`    size_t size);`)
	p(`int %s_append_value(%s vec, %s val);`, n, n, ctype)
	p(`size_t %s_size(%s vec);`, n, n)
	p(`int %s_get(`, n)
	p(`    %s res,`, rctype)
	p(`    const %s vec,`, n)
	p(`    size_t idx);`)
	return sb.String()
}

func generateVectorImpl(t VectorTypeSpec) string {
	n := "mlx_vector_" + t.ShortName
	ctype := t.CType
	rctype := t.ReturnCType
	cpp := t.CPPType
	ctocpp := t.CToCPP
	assign := t.CAssign

	var sb strings.Builder
	p := func(f string, args ...any) { fmt.Fprintf(&sb, f+"\n", args...) }

	// new
	p(``)
	p(`extern "C" %s %s_new(void) {`, n, n)
	p(`  try {`)
	p(`    return %s_new_({});`, n)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return %s_new_();`, n)
	p(`  }`)
	p(`}`)
	// set
	p(``)
	p(`extern "C" int %s_set(`, n)
	p(`    %s* vec,`, n)
	p(`    const %s src) {`, n)
	p(`  try {`)
	p(`    %s_set_(*vec, %s_get_(src));`, n, n)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// free
	p(``)
	p(`extern "C" int %s_free(%s vec) {`, n, n)
	p(`  try {`)
	p(`    %s_free_(vec);`, n)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// new_data
	p(``)
	p(`extern "C" %s %s_new_data(`, n, n)
	p(`    %s* data,`, ctype)
	p(`    size_t size) {`)
	p(`  try {`)
	p(`    auto vec = %s_new();`, n)
	p(`    for (size_t i = 0; i < size; i++) {`)
	p(`      %s_get_(vec).push_back(%s);`, n, ctocpp("data[i]"))
	p(`    }`)
	p(`    return vec;`)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return %s_new_();`, n)
	p(`  }`)
	p(`}`)
	// new_value
	p(``)
	p(`extern "C" %s %s_new_value(%s val) {`, n, n, ctype)
	p(`  try {`)
	p(`    return %s_new_({%s});`, n, ctocpp("val"))
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return %s_new_();`, n)
	p(`  }`)
	p(`}`)
	// set_data
	p(``)
	p(`extern "C" int`)
	p(`%s_set_data(%s* vec_, %s* data, size_t size) {`, n, n, ctype)
	p(`  try {`)
	p(`    std::vector<%s> cpp_arrs;`, cpp)
	p(`    for (size_t i = 0; i < size; i++) {`)
	p(`      cpp_arrs.push_back(%s);`, ctocpp("data[i]"))
	p(`    }`)
	p(`    %s_set_(*vec_, cpp_arrs);`, n)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// set_value
	p(``)
	p(`extern "C" int %s_set_value(%s* vec_, %s val) {`, n, n, ctype)
	p(`  try {`)
	p(`    %s_set_(*vec_, std::vector<%s>({%s}));`, n, cpp, ctocpp("val"))
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// append_data
	p(``)
	p(`extern "C" int`)
	p(`%s_append_data(%s vec, %s* data, size_t size) {`, n, n, ctype)
	p(`  try {`)
	p(`    for (size_t i = 0; i < size; i++) {`)
	p(`      %s_get_(vec).push_back(%s);`, n, ctocpp("data[i]"))
	p(`    }`)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// append_value
	p(``)
	p(`extern "C" int %s_append_value(`, n)
	p(`    %s vec,`, n)
	p(`    %s value) {`, ctype)
	p(`  try {`)
	p(`    %s_get_(vec).push_back(%s);`, n, ctocpp("value"))
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// get
	p(``)
	p(`extern "C" int %s_get(`, n)
	p(`    %s res,`, rctype)
	p(`    const %s vec,`, n)
	p(`    size_t index) {`)
	p(`  try {`)
	p(`    %s;`, assign("res", n+"_get_(vec).at(index)"))
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 1;`)
	p(`  }`)
	p(`  return 0;`)
	p(`}`)
	// size
	p(``)
	p(`extern "C" size_t %s_size(%s vec) {`, n, n)
	p(`  try {`)
	p(`    return %s_get_(vec).size();`, n)
	p(`  } catch (std::exception& e) {`)
	p(`    mlx_error(e.what());`)
	p(`    return 0;`)
	p(`  }`)
	p(`}`)

	return sb.String()
}

// ── static boilerplate ───────────────────────────────────────────────────────

const vectorDeclBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_VECTOR_H
#define MLX_VECTOR_H

#include "mlx/c/array.h"
#include "mlx/c/string.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * \defgroup mlx_vector Vectors
 * MLX vector objects.
 */
/**@{*/
`

const vectorDeclEnd = `
/**@}*/

#ifdef __cplusplus
}
#endif

#endif
`

const vectorImplBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#include "mlx/c/vector.h"
#include "mlx/c/error.h"
#include "mlx/c/private/mlx.h"
`

const vectorImplEnd = "\n"

const vectorPrivBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_VECTOR_PRIVATE_H
#define MLX_VECTOR_PRIVATE_H

#include "mlx/c/vector.h"
#include "mlx/mlx.h"
`

const vectorPrivEnd = `
#endif
`

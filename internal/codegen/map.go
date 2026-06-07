package codegen

import (
	"fmt"
	"io"
)

// MapTypeSpec describes one side (key or value) of a map type.
type MapTypeSpec struct {
	C       string // C type, e.g. "const mlx_array"
	CPP     string // C++ type, e.g. "mlx::core::array"
	Nick    string // short name used in identifiers, e.g. "array"
	CReturn string // out-param type, e.g. "mlx_array*"
	CToCPP  func(name string) string
	CAssign func(dest, src string) string
}

var mapArrayType = MapTypeSpec{
	C:       "const mlx_array",
	CPP:     "mlx::core::array",
	Nick:    "array",
	CReturn: "mlx_array*",
	CToCPP:  func(s string) string { return "mlx_array_get_(" + s + ")" },
	CAssign: func(d, s string) string { return "mlx_array_set_(*" + d + ", " + s + ")" },
}

var mapStringType = MapTypeSpec{
	C:       "const char*",
	CPP:     "std::string",
	Nick:    "string",
	CReturn: "const char**",
	CToCPP:  func(s string) string { return "std::string(" + s + ")" },
	CAssign: func(d, s string) string { return "*" + d + " = " + s + ".data()" },
}

func defaultMapPairs() [][2]MapTypeSpec {
	return [][2]MapTypeSpec{
		{mapStringType, mapArrayType},
		{mapStringType, mapStringType},
	}
}

// GenerateMapHeader writes the mlx/c/map.h file.
func GenerateMapHeader(w io.Writer) {
	fmt.Fprint(w, mapDeclBegin)
	for _, pair := range defaultMapPairs() {
		fmt.Fprint(w, generateMapDecl(pair[0], pair[1]))
	}
	fmt.Fprint(w, mapDeclEnd)
}

// GenerateMapImpl writes the mlx/c/map.cpp file.
func GenerateMapImpl(w io.Writer) {
	fmt.Fprint(w, mapImplBegin)
	for _, pair := range defaultMapPairs() {
		fmt.Fprint(w, generateMapImpl(pair[0], pair[1]))
	}
	fmt.Fprint(w, mapImplEnd)
}

// GenerateMapPrivate writes the mlx/c/private/map.h file.
func GenerateMapPrivate(w io.Writer) {
	fmt.Fprint(w, mapPrivBegin)
	for _, pair := range defaultMapPairs() {
		k, v := pair[0], pair[1]
		mapCType := "mlx_map_" + k.Nick + "_to_" + v.Nick
		mapCPPType := "std::unordered_map<" + k.CPP + ", " + v.CPP + ">"
		iterCType := mapCType + "_iterator"
		iterCPPType := mapCPPType + "::iterator"

		fmt.Fprint(w, privateCode(mapCType, mapCPPType, false, ""))
		fmt.Fprint(w, privateCode(iterCType, iterCPPType, false, ""))
		// Extra helper for map iterator
		fmt.Fprintf(w, `
inline %s& %s_get_map_(%s d) {
  return *static_cast<%s*>(d.map_ctx);
}
`, mapCPPType, iterCType, iterCType, mapCPPType)
	}
	fmt.Fprint(w, mapPrivEnd)
}

func generateMapDecl(k, v MapTypeSpec) string {
	n := "mlx_map_" + k.Nick + "_to_" + v.Nick
	return fmt.Sprintf(`
/**
 * A %s-to-%s map
 */
typedef struct %s_ {
  void* ctx;
} %s;

/**
 * Returns a new empty %s-to-%s map.
 */
%s %s_new(void);
/**
 * Set map to provided src map.
 */
int %s_set(
    %s* map,
    const %s src);
/**
 * Free a %s-to-%s map.
 */
int %s_free(%s map);
/**
 * Insert a new `+"`"+`value`+"`"+` at the specified `+"`"+`key`+"`"+` in the map.
 */
int %s_insert(
    %s map,
    %s key,
    %s value);
/**
 * Returns the value indexed at the specified `+"`"+`key`+"`"+` in the map.
 */
int %s_get(
    %s value,
    const %s map,
    %s key);

/**
 * An iterator over a %s-to-%s map.
 */
typedef struct %s_iterator_ {
  void* ctx;
  void* map_ctx;
} %s_iterator;
/**
 * Returns a new iterator over the given map.
 */
%s_iterator %s_iterator_new(
    %s map);
/**
 * Free iterator.
 */
int %s_iterator_free(
    %s_iterator it);
/**
 * Increment iterator.
 */
int %s_iterator_next(
    %s key,
    %s value,
    %s_iterator it);
`,
		k.Nick, v.Nick,
		n, n,
		k.Nick, v.Nick,
		n, n,
		n, n, n,
		k.Nick, v.Nick,
		n, n,
		n, n, k.C, v.C,
		n, v.CReturn, n, k.C,
		k.Nick, v.Nick,
		n, n,
		n, n, n,
		n, n,
		n, k.CReturn, v.CReturn, n,
	)
}

func generateMapImpl(k, v MapTypeSpec) string {
	n := "mlx_map_" + k.Nick + "_to_" + v.Nick
	cppmap := "std::unordered_map<" + k.CPP + ", " + v.CPP + ">"

	return fmt.Sprintf(`
extern "C" %s %s_new(void) {
  try {
    return %s_new_({});
  } catch (std::exception& e) {
    mlx_error(e.what());
    return %s_new_();
  }
}

extern "C" int %s_set(
    %s* map,
    const %s src) {
  try {
    %s_set_(*map, %s_get_(src));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" int %s_free(%s map) {
  try {
    %s_free_(map);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" int %s_insert(
    %s map,
    %s key,
    %s value) {
  try {
    %s_get_(map).insert_or_assign(%s, %s);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" int %s_get(
    %s value,
    const %s map,
    %s key) {
  try {
    auto search = %s_get_(map).find(%s);
    if (search == %s_get_(map).end()) {
      return 2;
    } else {
      %s;
      return 0;
    }
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" %s_iterator
%s_iterator_new(%s map) {
  auto& cpp_map = %s_get_(map);
  try {
    return %s_iterator{
        new %s::iterator(cpp_map.begin()),
        &cpp_map};
  } catch (std::exception& e) {
    mlx_error(e.what());
    return %s_iterator{0};
  }
}

extern "C" int %s_iterator_next(
    %s key,
    %s value,
    %s_iterator it) {
  try {
    if (%s_iterator_get_(it) ==
        %s_iterator_get_map_(it).end()) {
      return 2;
    } else {
      %s;
      %s;
      %s_iterator_get_(it)++;
      return 0;
    }
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" int %s_iterator_free(
    %s_iterator it) {
  try {
    %s_iterator_free_(it);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
`,
		// new
		n, n, n, n,
		// set
		n, n, n, n, n,
		// free
		n, n, n,
		// insert
		n, n, k.C, v.C, n, k.CToCPP("key"), v.CToCPP("value"),
		// get
		n, v.CReturn, n, k.C, n, k.CToCPP("key"), n,
		v.CAssign("value", "search->second"),
		// iterator_new
		n, n, n, n, n, cppmap, n,
		// iterator_next
		n, k.CReturn, v.CReturn, n,
		n, n,
		k.CAssign("key", n+"_iterator_get_(it)->first"),
		v.CAssign("value", n+"_iterator_get_(it)->second"),
		n,
		// iterator_free
		n, n, n,
	)
}

// ── boilerplate ───────────────────────────────────────────────────────────────

const mapDeclBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_MAP_H
#define MLX_MAP_H

#include "mlx/c/array.h"
#include "mlx/c/string.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * \defgroup mlx_map Maps
 * MLX map objects.
 */
/**@{*/
`

const mapDeclEnd = `
/**@}*/

#ifdef __cplusplus
}
#endif

#endif
`

const mapImplBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#include "mlx/c/map.h"
#include "mlx/c/error.h"
#include "mlx/c/private/mlx.h"
`

const mapImplEnd = "\n"

const mapPrivBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_MAP_PRIVATE_H
#define MLX_MAP_PRIVATE_H

#include "mlx/c/map.h"
#include "mlx/mlx.h"
`

const mapPrivEnd = `
#endif
`

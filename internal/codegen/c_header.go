package codegen

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/hooks"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/model"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/types"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/variants"
)

// GenerateCHeader writes a C header file (.h) to w.
// headers is the original header path(s) used in the #include guard name.
// headername is the base name (without extension) used in the include guard.
// docstring is an optional Doxygen docstring for the module group.
func GenerateCHeader(
	w io.Writer,
	funcs map[string][]model.Function,
	enums map[string]model.Enum,
	headername string,
	docstring string,
) {
	p := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	pr := func(s string) { fmt.Fprint(w, s) }

	pr(cFileHeader)

	guard := "MLX_" + strings.ToUpper(headername) + "_H"
	p("#ifndef %s", guard)
	p("#define %s", guard)
	pr(cStdIncludes)
	pr(cMLXIncludes)
	p(`#ifdef __cplusplus
extern "C" {
#endif
`)
	if docstring != "" {
		docstring = strings.ReplaceAll(docstring, "\n", "\n * ")
		p("/**")
		p(" * \\defgroup %s %s", headername, docstring)
		p(" */")
		p("/**@{*/")
		p("")
	}

	// Enums
	for _, enum := range sortedEnums(enums) {
		cTypeName := "mlx_" + ToSnakeLetters(enum.Name)
		var vals []string
		for _, v := range enum.Values {
			vals = append(vals, "  MLX_"+strings.ToUpper(ToSnakeLetters(enum.Name))+"_"+strings.ToUpper(v))
		}
		p("typedef enum %s_ {", cTypeName)
		p("%s", strings.Join(vals, ",\n"))
		p("} %s;", cTypeName)
		p("")
	}

	// Functions
	for _, f := range sortedFunctions(funcs) {
		cname := FuncCName(f.Namespace, f.Name, f.Variant)

		if r := hooks.Apply(cname, false); r.Handled {
			if r.Code != "" {
				pr(r.Code)
				p("")
			}
			continue
		}

		sig, ok := buildCSignature(f)
		if !ok {
			continue
		}
		p("%s;", sig)
	}

	// Module-level hook (e.g. export) — emits content regardless of parsed functions.
	if code := hooks.ApplyModule(headername, false); code != "" {
		pr(code)
	}

	if docstring != "" {
		p("")
		p("/**@}*/")
	}

	pr(`
#ifdef __cplusplus
}
#endif

#endif
`)
}

// buildCSignature returns the "int funcname(args)" signature string.
// Returns (sig, true) on success or ("", false) if any type is unsupported.
func buildCSignature(f model.Function) (string, bool) {
	retT, ok := types.FindByCPP(f.ReturnType)
	if !ok {
		return "", false
	}

	cname := FuncCName(f.Namespace, f.Name, f.Variant)
	var args []string

	// Return value as first out-param.
	if retT.CReturnArg != nil {
		if ra := retT.CReturnArg("res"); ra != "" {
			args = append(args, ra)
		}
	}

	// Input parameters.
	for i, param := range f.Params {
		if f.UseDefaults && f.Defaults[i] != "" {
			continue
		}
		pT, ok := types.FindByCPP(param.Type)
		if !ok {
			return "", false
		}
		if pT.CArg == nil {
			continue
		}
		args = append(args, pT.CArg(param.Name))
	}

	cargs := "void"
	if len(args) > 0 {
		cargs = strings.Join(args, ", ")
	}

	sig := fmt.Sprintf("int %s(%s)", cname, cargs)
	// Clean up spaces around parens (matches Python output style).
	sig = strings.ReplaceAll(sig, "( ", "(")
	sig = strings.ReplaceAll(sig, " )", ")")
	return sig, true
}

// ── sorting helpers ──────────────────────────────────────────────────────────

func sortedEnums(enums map[string]model.Enum) []model.Enum {
	keys := make([]string, 0, len(enums))
	for k := range enums {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]model.Enum, 0, len(enums))
	for _, k := range keys {
		out = append(out, enums[k])
	}
	return out
}

// sortedFunctions returns a flat, deduplicated, variant-resolved list of
// functions sorted by name — exactly as Python's generate() does.
func sortedFunctions(funcs map[string][]model.Function) []model.Function {
	// Collect map keys and sort them so iteration is deterministic.
	keys := make([]string, 0, len(funcs))
	for k := range funcs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []model.Function
	for _, fqname := range keys {
		defs := funcs[fqname]
		parts := strings.Split(fqname, "::")
		ns := strings.Join(parts[:len(parts)-1], "::")
		name := parts[len(parts)-1]

		// Sort overloads by descending parameter count (matches Python).
		// Stable sort preserves declaration order for same param count.
		sort.SliceStable(defs, func(i, j int) bool {
			return len(defs[i].Params) > len(defs[j].Params)
		})

		// Apply namespace-specific variant resolution.
		nsKey := variants.NSKey(ns)
		var resolved []model.Function
		if ns == "mlx::core::detail" {
			resolved = variants.ResolveDetail(name, defs)
		} else {
			resolved = variants.Resolve(nsKey, name, defs)
		}

		// Deduplicate by variant suffix.
		seen := map[string]bool{}
		for _, f := range resolved {
			if !seen[f.Variant] {
				seen[f.Variant] = true
				out = append(out, f)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		// Primary: C++ function base name alphabetically
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		// Secondary: most params first (matches Python overload ordering)
		if len(out[i].Params) != len(out[j].Params) {
			return len(out[i].Params) > len(out[j].Params)
		}
		// Equal: preserve original (declaration) order via stable sort
		return false
	})
	return out
}

// ── static string constants ──────────────────────────────────────────────────

const cFileHeader = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

`

const cStdIncludes = `
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>

`

const cMLXIncludes = `#include "mlx/c/array.h"
#include "mlx/c/closure.h"
#include "mlx/c/distributed_group.h"
#include "mlx/c/io_types.h"
#include "mlx/c/map.h"
#include "mlx/c/stream.h"
#include "mlx/c/string.h"
#include "mlx/c/vector.h"

`

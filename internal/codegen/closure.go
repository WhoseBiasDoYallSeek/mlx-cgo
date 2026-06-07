package codegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/types"
)

// ClosureSpec describes a single closure type to generate.
type ClosureSpec struct {
	Name      string   // e.g. "mlx_closure"
	ReturnCPP string   // C++ return type key in the type registry
	InputCPPs []string // C++ input types
}

func defaultClosureSpecs() []ClosureSpec {
	va := "std::vector<mlx::core::array>"
	vi := "std::vector<int>"
	pva := "std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>"
	pvi := "std::pair<std::vector<mlx::core::array>, @std::vector<int>>"
	sm := "std::unordered_map<std::string, mlx::core::array>"

	return []ClosureSpec{
		{Name: "mlx_closure", ReturnCPP: va, InputCPPs: []string{va}},
		{Name: "mlx_closure_kwargs", ReturnCPP: va, InputCPPs: []string{va, sm}},
		{Name: "mlx_closure_value_and_grad", ReturnCPP: pva, InputCPPs: []string{va}},
		{Name: "mlx_closure_custom", ReturnCPP: va, InputCPPs: []string{va, va, va}},
		{Name: "mlx_closure_custom_jvp", ReturnCPP: va, InputCPPs: []string{va, va, vi}},
		{Name: "mlx_closure_custom_vmap", ReturnCPP: pvi, InputCPPs: []string{va, vi}},
	}
}

// GenerateClosureHeader writes the mlx/c/closure.h file.
func GenerateClosureHeader(w io.Writer) {
	fmt.Fprint(w, closureDeclBegin)
	for _, spec := range defaultClosureSpecs() {
		fmt.Fprint(w, generateClosureDecl(spec))
		if spec.Name == "mlx_closure" {
			fmt.Fprintln(w, "\nmlx_closure mlx_closure_new_unary(int (*fun)(mlx_array*, const mlx_array));")
		}
	}
	fmt.Fprint(w, closureDeclEnd)
}

// GenerateClosureImpl writes the mlx/c/closure.cpp file.
func GenerateClosureImpl(w io.Writer) {
	fmt.Fprint(w, closureImplBegin)
	for _, spec := range defaultClosureSpecs() {
		fmt.Fprint(w, generateClosureImpl(spec))
		if spec.Name == "mlx_closure" {
			fmt.Fprint(w, closureUnaryImpl)
		}
	}
	fmt.Fprint(w, closureImplEnd)
}

// GenerateClosurePrivate writes the mlx/c/private/closure.h file.
func GenerateClosurePrivate(w io.Writer) {
	fmt.Fprint(w, closurePrivBegin)
	for _, spec := range defaultClosureSpecs() {
		// Reconstruct the std::function<> type for the private header.
		retT, _ := types.FindByCPP(spec.ReturnCPP)
		retCPP := strings.ReplaceAll(retT.CPPName, "@", "")

		var inputCPPs []string
		for _, inp := range spec.InputCPPs {
			inT, _ := types.FindByCPP(inp)
			cpp := strings.ReplaceAll(inT.CPPName, "@", "")
			inputCPPs = append(inputCPPs, cpp)
		}
		fnType := "std::function<" + retCPP + "(" + strings.Join(inputCPPs, ", ") + ")>"
		fmt.Fprint(w, privateCode(spec.Name, fnType, false, ""))
	}
	fmt.Fprint(w, closurePrivEnd)
}

// ── declaration generator ────────────────────────────────────────────────────

func generateClosureDecl(spec ClosureSpec) string {
	name := spec.Name
	retT, ok := types.FindByCPP(spec.ReturnCPP)
	if !ok {
		return ""
	}

	rcArg := ""
	if retT.CReturnArg != nil {
		rcArg = retT.CReturnArg("res")
	}
	rcArgUnnamed := ""
	if retT.CReturnArg != nil {
		rcArgUnnamed = strings.TrimSpace(retT.CReturnArg(""))
	}

	var cArgs, cArgsUnnamed []string
	for i, inp := range spec.InputCPPs {
		t, ok := types.FindByCPP(inp)
		if !ok {
			continue
		}
		suffix := ""
		if len(spec.InputCPPs) > 1 {
			suffix = fmt.Sprintf("_%d", i)
		}
		name2 := "input" + suffix
		if t.CArg != nil {
			cArgs = append(cArgs, strings.TrimSpace(t.CArg(name2)))
			cArgsUnnamed = append(cArgsUnnamed, strings.TrimSpace(t.CArg("")))
		}
	}

	allArgs := strings.Join(cArgs, ", ")
	unnamedArgs := strings.Join(cArgsUnnamed, ", ")
	rcArgs := rcArg
	if rcArgs != "" && allArgs != "" {
		allArgs = rcArgs + ", " + allArgs
		unnamedArgs = rcArgUnnamed + ", " + unnamedArgs
	} else if rcArgs != "" {
		allArgs = rcArgs
		unnamedArgs = rcArgUnnamed
	}

	return fmt.Sprintf(`
typedef struct %s_ {
  void* ctx;
} %s;
%s %s_new(void);
int %s_free(%s cls);
%s %s_new_func(int (*fun)(%s));
%s %s_new_func_payload(
    int (*fun)(%s, void*),
    void* payload,
    void (*dtor)(void*));
int %s_set(%s *cls, const %s src);
int %s_apply(%s, %s cls, %s);
`,
		name, name,
		name, name,
		name, name,
		name, name, unnamedArgs,
		name, name, unnamedArgs,
		name, name, name,
		name, rcArg, name, strings.Join(cArgs, ", "),
	)
}

// ── implementation generator ─────────────────────────────────────────────────

func generateClosureImpl(spec ClosureSpec) string {
	name := spec.Name
	retT, ok := types.FindByCPP(spec.ReturnCPP)
	if !ok {
		return ""
	}

	retCPP := strings.ReplaceAll(retT.CPPName, "@", "")
	_ = retCPP
	rcArg := ""
	if retT.CReturnArg != nil {
		rcArg = retT.CReturnArg("res")
	}
	rcArgUnnamed := ""
	if retT.CReturnArg != nil {
		rcArgUnnamed = strings.TrimSpace(retT.CReturnArg(""))
	}
	rcArgUntyped := ""
	if retT.CReturnArg != nil {
		// Extract just the variable names from the typed declaration and prefix with &.
		// e.g. "mlx_vector_array* res" → "&res"
		// e.g. "mlx_vector_array* res_0, mlx_vector_array* res_1" → "&res_0, &res_1"
		typed := retT.CReturnArg("res")
		var refs []string
		for _, part := range strings.Split(typed, ",") {
			part = strings.TrimSpace(part)
			fields := strings.Fields(part)
			if len(fields) > 0 {
				varName := fields[len(fields)-1]
				varName = strings.TrimPrefix(varName, "*")
				refs = append(refs, "&"+varName)
			}
		}
		rcArgUntyped = strings.Join(refs, ", ")
	}

	// c_new / free / to_cpp for return type
	rcNew := ""
	if retT.CNew != nil {
		cn := retT.CNew("res")
		// Multi-line CNew (e.g. pair types) needs a trailing standalone ";"
		// to match the Python generator's output pattern.
		if strings.Contains(cn, "\n") {
			rcNew = cn + ";\n;"
		} else {
			rcNew = cn + ";"
		}
	}
	rcFree := ""
	if retT.Free != nil {
		rcFree = retT.Free("res") + ";"
	}
	rcToCPP := ""
	if retT.CToCPP != nil {
		rcToCPP = "auto cpp_res = " + retT.CToCPP("res") + ";"
	}

	// assign from C++ back to C return param
	assignCls := ""
	if retT.CAssignFromCPP != nil {
		assignCls = retT.CAssignFromCPP("res", name+"_get_(cls)("+buildCPPArgs(spec)+")") + ";"
	}

	// input parameters
	var cppArgsTypeName, cppArgsToCArgs, cArgsFree, cArgsUntyped, cArgsDecl, cArgsUnnamed []string
	for i, inp := range spec.InputCPPs {
		t, ok2 := types.FindByCPP(inp)
		if !ok2 {
			continue
		}
		suffix := ""
		if len(spec.InputCPPs) > 1 {
			suffix = fmt.Sprintf("_%d", i)
		}
		argName := "input" + suffix
		cppName := "cpp_input" + suffix

		inpCPP := strings.ReplaceAll(t.CPPName, "@", "")
		cppArgsTypeName = append(cppArgsTypeName, "const "+inpCPP+"& "+cppName)

		if t.CNew != nil {
			cppArgsToCArgs = append(cppArgsToCArgs, t.CNew(argName)+";")
		}
		if t.CAssignFromCPP != nil {
			// Local variable assignment: CAssignFromCPP dereferences for out-params
			// (e.g. *res), but here argName is a local var, not a pointer.
			// Strip the dereference by replacing "cname_set_(*argName," with "cname_set_(argName,".
			assign := t.CAssignFromCPP(argName, cppName)
			assign = strings.ReplaceAll(assign, "_set_(*"+argName+",", "_set_("+argName+",")
			cppArgsToCArgs = append(cppArgsToCArgs, assign+";")
		}
		if t.Free != nil {
			// Always append the free statement (even if empty) so raw-vector types
			// generate a placeholder ";" that matches the Python generator output.
			cArgsFree = append(cArgsFree, t.Free(argName)+";")
		}
		if t.CArg != nil {
			decl := strings.TrimSpace(t.CArg(argName))
			unnamed := strings.TrimSpace(t.CArg(""))
			cArgsDecl = append(cArgsDecl, decl)
			cArgsUnnamed = append(cArgsUnnamed, unnamed)
		}
		// untyped (no type name, just var name); raw-vector types need both ptr+num.
		cArgsUntyped = append(cArgsUntyped, argName)
		if t.CArg != nil && strings.Contains(t.CArg(argName), ", size_t ") {
			cArgsUntyped = append(cArgsUntyped, argName+"_num")
		}
	}

	unnamedArgs := strings.Join(cArgsUnnamed, ", ")
	unnamedFull := unnamedArgs
	if rcArgUnnamed != "" && unnamedFull != "" {
		unnamedFull = rcArgUnnamed + ", " + unnamedFull
	} else if rcArgUnnamed != "" {
		unnamedFull = rcArgUnnamed
	}

	allDeclArgs := strings.Join(cArgsDecl, ", ")
	allApplyArgs := rcArg
	if allApplyArgs != "" && allDeclArgs != "" {
		allApplyArgs += ", " + allDeclArgs
	} else if allDeclArgs != "" {
		allApplyArgs = allDeclArgs
	}

	cppArgsToCArgsStr := strings.Join(cppArgsToCArgs, "\n      ")
	cArgsFreeStr := strings.Join(cArgsFree, "\n      ")

	return fmt.Sprintf(`
extern "C" %s %s_new(void) {
  try {
    return %s_new_();
  } catch (std::exception& e) {
    mlx_error(e.what());
    return %s_new_();
  }
}

extern "C" int %s_set(%s *cls, const %s src) {
  try {
    %s_set_(*cls, %s_get_(src));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" int %s_free(%s cls) {
  try {
    %s_free_(cls);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" %s %s_new_func(int (*fun)(%s)) {
  try {
    auto cpp_closure = [fun](%s) {
      %s
      %s
      auto status = fun(%s, %s);
      %s
      if(status) {
        %s
        throw std::runtime_error("%s returned a non-zero value");
      }
      %s
      %s
      return cpp_res;
    };
    return %s_new_(cpp_closure);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return %s_new_();
  }
}

extern "C" %s %s_new_func_payload(
    int (*fun)(%s, void*),
    void* payload,
    void (*dtor)(void*)) {
  try {
    std::shared_ptr<void> cpp_payload = nullptr;
    if (dtor) {
      cpp_payload = std::shared_ptr<void>(payload, dtor);
    } else {
      cpp_payload = std::shared_ptr<void>(payload, [](void*) {});
    }
    auto cpp_closure = [fun, cpp_payload](%s) {
      %s
      %s
      auto status = fun(%s, %s, cpp_payload.get());
      %s
      if(status) {
        %s
        throw std::runtime_error("%s returned a non-zero value");
      }
      %s
      %s
      return cpp_res;
    };
    return %s_new_(cpp_closure);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return %s_new_();
  }
}

extern "C" int %s_apply(%s, %s cls, %s) {
  try {
    %s
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
`,
		// new
		name, name, name, name,
		// set
		name, name, name, name, name,
		// free
		name, name, name,
		// new_func
		name, name, unnamedFull,
		strings.Join(cppArgsTypeName, ", "),
		cppArgsToCArgsStr, rcNew,
		rcArgUntyped, strings.Join(cArgsUntyped, ", "),
		cArgsFreeStr, rcFree, name,
		rcToCPP, rcFree,
		name, name,
		// new_func_payload
		name, name, unnamedFull,
		strings.Join(cppArgsTypeName, ", "),
		cppArgsToCArgsStr, rcNew,
		rcArgUntyped, strings.Join(cArgsUntyped, ", "),
		cArgsFreeStr, rcFree, name,
		rcToCPP, rcFree,
		name, name,
		// apply
		name, rcArg, name, strings.Join(cArgsDecl, ", "),
		assignCls,
	)
}

func buildCPPArgs(spec ClosureSpec) string {
	var args []string
	for i, inp := range spec.InputCPPs {
		t, ok := types.FindByCPP(inp)
		if !ok {
			continue
		}
		suffix := ""
		if len(spec.InputCPPs) > 1 {
			suffix = fmt.Sprintf("_%d", i)
		}
		if t.CToCPP != nil {
			args = append(args, t.CToCPP("input"+suffix))
		}
	}
	return strings.Join(args, ", ")
}

const closureUnaryImpl = `
extern "C" mlx_closure mlx_closure_new_unary(
    int (*fun)(mlx_array*, const mlx_array)) {
  try {
    auto cpp_closure = [fun](const std::vector<mlx::core::array>& cpp_input) {
      if (cpp_input.size() != 1) {
        throw std::runtime_error("closure: expected unary input");
      }
      auto input = mlx_array_new_(cpp_input[0]);
      auto res = mlx_array_new_();
      auto status = fun(&res, input);
      if(status) {
        mlx_array_free_(res);
        throw std::runtime_error("mlx_closure returned a non-zero value");
      }
      mlx_array_free(input);
      std::vector<mlx::core::array> cpp_res = {mlx_array_get_(res)};
      mlx_array_free(res);
      return cpp_res;
    };
    return mlx_closure_new_(cpp_closure);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return mlx_closure_new_();
  }
}
`

// ── boilerplate ───────────────────────────────────────────────────────────────

const closureDeclBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_H
#define MLX_CLOSURE_H

#include "mlx/c/array.h"
#include "mlx/c/map.h"
#include "mlx/c/optional.h"
#include "mlx/c/stream.h"
#include "mlx/c/vector.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * \defgroup mlx_closure Closures
 * MLX closure objects.
 */
/**@{*/
`

const closureDeclEnd = `
/**@}*/

#ifdef __cplusplus
}
#endif

#endif
`

const closureImplBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#include "mlx/c/closure.h"
#include "mlx/c/error.h"
#include "mlx/c/private/mlx.h"
`

const closureImplEnd = "\n"

const closurePrivBegin = `/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_PRIVATE_H
#define MLX_CLOSURE_PRIVATE_H

#include "mlx/c/closure.h"
#include "mlx/mlx.h"

`

const closurePrivEnd = `
#endif
`

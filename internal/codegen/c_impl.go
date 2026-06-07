package codegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/hooks"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/model"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/types"
)

// GenerateCImpl writes a C++ implementation file (.cpp) to w.
// headers is the semicolon-separated list of original C++ header paths.
// headername is the base name used in the #include directive.
func GenerateCImpl(
	w io.Writer,
	funcs map[string][]model.Function,
	enums map[string]model.Enum,
	headers []string,
	headername string,
) {
	p := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	pr := func(s string) { fmt.Fprint(w, s) }

	pr(cFileHeader)

	p(`#include "mlx/c/%s.h"`, headername)
	for _, h := range headers {
		// Extract the path from "mlx/" onwards to produce a portable include.
		if idx := strings.Index(h, "mlx/"); idx >= 0 {
			p(`#include "%s"`, h[idx:])
		}
	}
	p(`#include "mlx/c/error.h"`)
	p(`#include "mlx/c/private/mlx.h"`)
	p("")

	for _, f := range sortedFunctions(funcs) {
		cname := FuncCName(f.Namespace, f.Name, f.Variant)

		if r := hooks.Apply(cname, true); r.Handled {
			if r.Code != "" {
				pr(r.Code)
				p("")
			}
			continue
		}

		retT, ok := types.FindByCPP(f.ReturnType)
		if !ok {
			continue
		}

		sig, ok := buildCSignature(f)
		if !ok {
			continue
		}

		// Build the C++ call arguments.
		var cppArgs []string
		for i, param := range f.Params {
			if f.UseDefaults && f.Defaults[i] != "" {
				continue
			}
			pT, ok := types.FindByCPP(param.Type)
			if !ok {
				goto skipFunc
			}
			if pT.CToCPP != nil {
				cppArgs = append(cppArgs, pT.CToCPP(param.Name))
			}
		}

		{
			cppCall := f.Namespace + "::" + f.Name + "(" + strings.Join(cppArgs, ", ") + ")"
			assignStmt := ""
			if retT.CAssignFromCPP != nil {
				assignStmt = retT.CAssignFromCPP("res", cppCall)
			}

			p(`extern "C" %s {`, sig)
			p(`  try {`)
			p(`    %s;`, assignStmt)
			p(`  } catch (std::exception& e) {`)
			p(`    mlx_error(e.what());`)
			p(`    return 1;`)
			p(`  }`)
			p(`  return 0;`)
			p(`}`)
			p("")
		}
		continue

	skipFunc:
	}

	// Module-level hook (e.g. export) — emits content regardless of parsed functions.
	if code := hooks.ApplyModule(headername, true); code != "" {
		pr(code)
	}
}

package codegen

import (
	"fmt"
	"io"
	"strings"
)

// GeneratePrivateHeader writes a mlx/c/private/*.h file for a C opaque handle
// type, providing inline _new_, _set_, _get_, and _free_ helpers.
//
// This is a direct port of python/type_private_generator.py.
//
// Parameters:
//   - ctype    : C typedef name, e.g. "mlx_array"
//   - cpptype  : fully-qualified C++ type, e.g. "mlx::core::array"
//   - noCopy   : when true the set helper uses move-only semantics
//   - shortName: override for the include guard / #include path (defaults to ctype without "mlx_")
//   - mlxInclude: the mlx header to include (defaults to "mlx/mlx.h")
//   - using    : optional "using ALIAS = CPPTYPE;" declaration
func GeneratePrivateHeader(
	w io.Writer,
	ctype, cpptype string,
	noCopy bool,
	shortName, mlxInclude, usingDecl string,
) {
	// Support semicolon-separated multiple types (e.g. device has mlx_device + mlx_device_info)
	ctypes := strings.Split(ctype, ";")
	cpptypes := strings.Split(cpptype, ";")
	usings := strings.Split(usingDecl, ";")
	for len(usings) < len(ctypes) {
		usings = append(usings, "")
	}

	firstCtype := ctypes[0]
	if shortName == "" {
		shortName = strings.TrimPrefix(firstCtype, "mlx_")
	}
	if mlxInclude == "" {
		mlxInclude = "mlx/mlx.h"
	}

	p := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }

	p(`/* Copyright © 2023-2024 Apple Inc.                   */`)
	p(`/*                                                    */`)
	p(`/* This file is auto-generated. Do not edit manually. */`)
	p(`/*                                                    */`)
	p(``)
	guard := "MLX_" + strings.ToUpper(shortName) + "_PRIVATE_H"
	p("#ifndef %s", guard)
	p("#define %s", guard)
	p(``)
	p(`#include "mlx/c/%s.h"`, shortName)
	p(`#include "%s"`, mlxInclude)

	for i, ct := range ctypes {
		cp := cpptypes[i]
		ud := usings[i]
		if ud != "" {
			if i > 0 {
				p(``)
			}
			p(`using %s = %s;`, ud, cp)
		}
		code := privateCode(ct, cp, noCopy, ud)
		fmt.Fprint(w, code)
	}

	p(``)
	p("#endif")
}

// privateCode returns the inline helper code block for a given type pair.
func privateCode(ctype, cpptype string, noCopy bool, usingDecl string) string {
	effectiveCPP := cpptype
	if usingDecl != "" {
		effectiveCPP = usingDecl
	}

	ctorCopy := fmt.Sprintf(`
inline %s %s_new_(const %s& s) {
  return %s({new %s(s)});
}
`, ctype, ctype, effectiveCPP, ctype, effectiveCPP)

	var ctorCode string
	if noCopy {
		ctorCode = fmt.Sprintf(`
inline %s %s_new_() {
  return %s({nullptr});
}

inline %s %s_new_(%s&& s) {
  return %s({new %s(std::move(s))});
}
`, ctype, ctype, ctype,
			ctype, ctype, effectiveCPP, ctype, effectiveCPP)
	} else {
		ctorCode = fmt.Sprintf(`
inline %s %s_new_() {
  return %s({nullptr});
}
%s
inline %s %s_new_(%s&& s) {
  return %s({new %s(std::move(s))});
}
`, ctype, ctype, ctype,
			ctorCopy,
			ctype, ctype, effectiveCPP, ctype, effectiveCPP)
	}

	var setCode string
	if noCopy {
		setCode = fmt.Sprintf(`
inline %s& %s_set_(%s& d, %s&& s) {
  if (d.ctx) {
    delete static_cast<%s*>(d.ctx);
  }
  d.ctx = new %s(std::move(s));
  return d;
}
`, ctype, ctype, ctype, effectiveCPP, effectiveCPP, effectiveCPP)
	} else {
		setCode = fmt.Sprintf(`
inline %s& %s_set_(%s& d, const %s& s) {
  if (d.ctx) {
    *static_cast<%s*>(d.ctx) = s;
  } else {
    d.ctx = new %s(s);
  }
  return d;
}

inline %s& %s_set_(%s& d, %s&& s) {
  if (d.ctx) {
    *static_cast<%s*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new %s(std::move(s));
  }
  return d;
}
`, ctype, ctype, ctype, effectiveCPP, effectiveCPP, effectiveCPP,
			ctype, ctype, ctype, effectiveCPP, effectiveCPP, effectiveCPP)
	}

	rest := fmt.Sprintf(`
inline %s& %s_get_(%s d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty %s");
  }
  return *static_cast<%s*>(d.ctx);
}

inline void %s_free_(%s d) {
  if (d.ctx) {
    delete static_cast<%s*>(d.ctx);
  }
}
`, effectiveCPP, ctype, ctype, ctype, effectiveCPP,
		ctype, ctype, effectiveCPP)

	return ctorCode + setCode + rest
}

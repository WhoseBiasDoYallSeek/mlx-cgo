// Copyright (c) 2025-2026 WhoseBiasDoYallSeek
// SPDX-License-Identifier: MIT
// See LICENSE file for full text.
//
// Command mlxgen is the MLX C bindings generator, rewritten in Go.
//
// It replaces the Python generator.py from the original mlx-c project.
// Given a C++ MLX header, it parses the AST via clang and emits either
// a C header, C implementation file, or Go/CGo wrapper.
//
// Usage:
//
//	# From C++ headers (dynamic generators):
//	mlxgen --header path/to/ops.h [--implementation] [--headername ops]
//
//	# Static generators (no C++ header needed):
//	mlxgen --type=vector [--implementation]
//	mlxgen --type=map    [--implementation]
//	mlxgen --type=closure [--implementation]
//	mlxgen --type=private --ctype mlx_array --cpptype mlx::core::array [--no-copy] [--mlx-include mlx/mlx.h]
//
//	# CGo wrappers from generated C header:
//	mlxgen --cgo-from mlx/c/fft.h [--package mlx]
//
//	# Go package scaffolding (base types + helpers):
//	mlxgen --type=go-base [--package mlx]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/cheader"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/codegen"
	cgogen "github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/codegen/cgo"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/parser"
)

func main() {
	// C generator flags
	header := flag.String("header", "", "Path to C++ header(s), semicolon-separated")
	implementation := flag.Bool("implementation", false, "Emit implementation (.cpp) instead of header (.h)")
	language := flag.String("language", "C", "Output language (currently only C is supported)")
	headernameFlag := flag.String("headername", "", "Override the base name used in include guards and #includes")
	docstring := flag.String("docstring", "", "Optional Doxygen docstring for the module group")

	// Static generator type
	genType := flag.String("type", "", "Static generator type: vector, map, closure, private, go-base")

	// --type=private flags (mirrors python/type_private_generator.py CLI)
	privCtype := flag.String("ctype", "", "C type name(s) semicolon-separated (for --type=private)")
	privCpptype := flag.String("cpptype", "", "C++ type name(s) semicolon-separated (for --type=private)")
	privNoCopy := flag.Bool("no-copy", false, "Use move-only set helper (for --type=private)")
	privInclude := flag.String("include", "", "Short name for include guard override (for --type=private)")
	privMlxInclude := flag.String("mlx-include", "mlx/mlx.h", "MLX header to include (for --type=private)")
	privUsing := flag.String("using", "", "Optional 'using ALIAS = CPPTYPE' semicolon-separated (for --type=private)")

	// CGo generator flags
	cgoFrom := flag.String("cgo-from", "", "Path to C header (.h) to generate Go/CGo wrappers from")
	pkg := flag.String("package", "mlx", "Go package name for --cgo-from and --type=go-base output")

	// Include path for C generator (passed as -I to clang)
	includeDirs := flag.String("include-dir", "", "Semicolon-separated -I paths passed to clang (for --header mode)")

	flag.Parse()

	// Static generator mode (--type)
	if *genType != "" {
		switch *genType {
		case "vector":
			if *implementation {
				codegen.GenerateVectorImpl(os.Stdout)
			} else {
				codegen.GenerateVectorHeader(os.Stdout)
			}
		case "map":
			if *implementation {
				codegen.GenerateMapImpl(os.Stdout)
			} else {
				codegen.GenerateMapHeader(os.Stdout)
			}
		case "closure":
			if *implementation {
				codegen.GenerateClosureImpl(os.Stdout)
			} else {
				codegen.GenerateClosureHeader(os.Stdout)
			}
		case "closure-private":
			codegen.GenerateClosurePrivate(os.Stdout)
		case "private":
			if *privCtype == "" || *privCpptype == "" {
				fmt.Fprintln(os.Stderr, "error: --type=private requires --ctype and --cpptype")
				os.Exit(1)
			}
			shortName := *privInclude
			if shortName == "" {
				// derive from first ctype: "mlx_array" → "array"
				shortName = strings.TrimPrefix(strings.SplitN(*privCtype, ";", 2)[0], "mlx_")
			}
			codegen.GeneratePrivateHeader(
				os.Stdout,
				*privCtype, *privCpptype,
				*privNoCopy,
				shortName, *privMlxInclude, *privUsing,
			)
		case "go-base":
			fmt.Print(cgogen.GenerateGoBase(*pkg))
		default:
			fmt.Fprintf(os.Stderr, "error: unknown --type %q (valid: vector, map, closure, private, go-base)\n", *genType)
			os.Exit(1)
		}
		return
	}

	// CGo mode: generate Go wrappers from an existing C header
	if *cgoFrom != "" {
		src, err := os.ReadFile(*cgoFrom)
		if err != nil {
			log.Fatalf("read %q: %v", *cgoFrom, err)
		}
		cfuncs := cheader.ParseHeader(src)
		out := cgogen.GenerateGoFile(cfuncs, cgogen.Options{PackageName: *pkg})
		fmt.Print(out)
		return
	}

	// C generator mode (--header)
	if *header == "" {
		fmt.Fprintln(os.Stderr, "error: one of --header, --type, or --cgo-from is required")
		flag.Usage()
		os.Exit(1)
	}
	if *language != "C" {
		fmt.Fprintf(os.Stderr, "error: unsupported language %q\n", *language)
		os.Exit(1)
	}

	headername := *headernameFlag
	if headername == "" {
		first := strings.SplitN(*header, ";", 2)[0]
		base := filepath.Base(first)
		if !strings.HasSuffix(base, ".h") {
			log.Fatalf("header %q does not end in .h; use --headername to set explicitly", base)
		}
		headername = strings.TrimSuffix(base, ".h")
	}

	root, err := parser.ParseHeader(*header, clangIncludeArgs(*includeDirs))
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	funcs, enums := parser.Extract(root)
	headers := strings.Split(*header, ";")

	if *implementation {
		codegen.GenerateCImpl(os.Stdout, funcs, enums, headers, headername)
	} else {
		codegen.GenerateCHeader(os.Stdout, funcs, enums, headername, *docstring)
	}
}

// clangIncludeArgs converts a semicolon-separated list of directories into
// the -I<dir> arguments expected by clang.
func clangIncludeArgs(dirs string) []string {
	if dirs == "" {
		return nil
	}
	var args []string
	for _, d := range strings.Split(dirs, ";") {
		d = strings.TrimSpace(d)
		if d != "" {
			args = append(args, "-I"+d)
		}
	}
	return args
}

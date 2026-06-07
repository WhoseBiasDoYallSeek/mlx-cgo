package parser

import (
	"strings"

	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/model"
)

// Extract walks the clang AST rooted at node and returns all functions and
// enums found under the mlx namespace tree.
//
// Returned maps use the fully-qualified name as key, e.g.
//
//	funcs["mlx::core::abs"]  → []model.Function{...}
//	enums["mlx::core::Dtype"] → model.Enum{...}
func Extract(root *Node) (funcs map[string][]model.Function, enums map[string]model.Enum) {
	funcs = make(map[string][]model.Function)
	enums = make(map[string]model.Enum)
	walkNamespace(root, "", funcs, enums)
	return
}

func walkNamespace(node *Node, ns string, funcs map[string][]model.Function, enums map[string]model.Enum) {
	for i := range node.Inner {
		child := &node.Inner[i]
		switch child.Kind {
		case "NamespaceDecl":
			childNS := joinNS(ns, child.Name)
			// Only descend into the mlx namespace tree to avoid picking up
			// standard library internals from transitively included headers.
			if ns == "" && child.Name != "mlx" {
				continue
			}
			walkNamespace(child, childNS, funcs, enums)

		case "FunctionDecl":
			// Only collect functions that live inside the mlx namespace tree
			// AND are defined directly in our target headers (not transitive includes).
			if !strings.HasPrefix(ns, "mlx") {
				continue
			}
			if !child.Loc.IsFromStdin() {
				continue
			}
			if f, ok := extractFunction(child, ns); ok {
				key := joinNS(ns, f.Name)
				funcs[key] = append(funcs[key], f)
			}

		case "EnumDecl":
			// Only collect enums that live inside the mlx namespace tree
			// AND are defined directly in our target headers.
			if !strings.HasPrefix(ns, "mlx") {
				continue
			}
			if !child.Loc.IsFromStdin() {
				continue
			}
			if child.Name == "" {
				continue
			}
			e := extractEnum(child, ns)
			enums[joinNS(ns, e.Name)] = e
		}
	}
}

func extractFunction(node *Node, ns string) (model.Function, bool) {
	name := node.Name
	if name == "" || strings.HasPrefix(name, "operator") {
		return model.Function{}, false
	}

	retType := returnTypeFromQual(node.Type.QualType)

	var params []model.Param
	var defaults []string
	for i := range node.Inner {
		child := &node.Inner[i]
		if child.Kind != "ParmVarDecl" {
			continue
		}
		pName := child.Name
		if pName == "" {
			pName = "param"
		}
		pType := normalizeType(child.Type.QualType)
		defaultVal := ""
		if child.Init != "" {
			// clang sets Init to non-empty when a default exists; we don't
			// need the actual value (the Python generator didn't use it
			// either — it only checked whether a default was present).
			defaultVal = "<default>"
		}
		params = append(params, model.Param{Name: pName, Type: pType, Default: defaultVal})
		defaults = append(defaults, defaultVal)
	}

	return model.Function{
		Name:       name,
		Namespace:  ns,
		ReturnType: retType,
		Params:     params,
		Defaults:   defaults,
	}, true
}

func extractEnum(node *Node, ns string) model.Enum {
	var values []string
	for i := range node.Inner {
		child := &node.Inner[i]
		if child.Kind == "EnumConstantDecl" {
			values = append(values, child.Name)
		}
	}
	return model.Enum{
		Name:      node.Name,
		Namespace: ns,
		Values:    values,
	}
}

// returnTypeFromQual extracts the return type from a clang function qualType
// string like "int (int, float)" or "std::pair<A, B> (const A &, const B &)".
//
// The strategy is to find the outermost '(' that starts the parameter list by
// scanning from the right while tracking <> and () depth.
func returnTypeFromQual(qualType string) string {
	depth := 0
	for i := len(qualType) - 1; i >= 0; i-- {
		switch qualType[i] {
		case ')':
			depth++
		case '(':
			depth--
			if depth == 0 {
				// Found the opening paren of the parameter list.
				ret := strings.TrimSpace(qualType[:i])
				return normalizeType(ret)
			}
		}
	}
	// Fallback: no parameter list found — treat whole string as type.
	return normalizeType(qualType)
}

// normalizeType strips C++ qualifiers (const, references, move-references)
// that the Python generator also stripped through cxxheaderparser, so that
// the resulting string matches the keys in the type registry.
func normalizeType(t string) string {
	t = strings.TrimSpace(t)
	// Strip leading const.
	t = strings.TrimPrefix(t, "const ")
	// Strip trailing reference/move-reference/pointer markers.
	t = strings.TrimSuffix(t, " &&")
	t = strings.TrimSuffix(t, " &")
	t = strings.TrimSpace(t)
	// clang sometimes adds _Bool; normalise to bool.
	if t == "_Bool" {
		return "bool"
	}
	// For types that contain std::function<...>, normalize the inner signature:
	// remove "const " and " &" inside the function parameter list so that
	// e.g. "std::function<R (const T &)>" → "std::function<R(T)>"
	if strings.Contains(t, "std::function<") {
		t = normalizeFunctionType(t)
	}
	return t
}

// normalizeFunctionType normalizes a std::function<R(params)> type string
// by removing const/ref qualifiers inside the function signature and
// collapsing extra whitespace, so the result matches the registry keys.
func normalizeFunctionType(t string) string {
	// Remove all occurrences of "const " inside the type.
	for strings.Contains(t, "const ") {
		t = strings.ReplaceAll(t, "const ", "")
	}
	// Remove all trailing " &" and " &&" patterns inside template params.
	for strings.Contains(t, " &") {
		t = strings.ReplaceAll(t, " &&", "")
		t = strings.ReplaceAll(t, " &", "")
	}
	// Collapse " (" → "(" to normalize space before param list.
	t = strings.ReplaceAll(t, " (", "(")
	// Collapse " >" → ">" (space before closing angle bracket).
	t = strings.ReplaceAll(t, " >", ">")
	return t
}

func joinNS(ns, name string) string {
	if ns == "" {
		return name
	}
	return ns + "::" + name
}

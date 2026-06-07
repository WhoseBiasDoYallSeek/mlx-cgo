// Package parser handles parsing C++ headers via the clang AST JSON output.
package parser

// Node is the top-level clang AST JSON node (TranslationUnitDecl).
type Node struct {
	Kind  string  `json:"kind"`
	Name  string  `json:"name"`
	Type  astType `json:"type"`
	Inner []Node  `json:"inner"`
	Init  string  `json:"init"` // non-empty means ParmVarDecl has a default value
	Loc   astLoc  `json:"loc"`
}

type astType struct {
	QualType string `json:"qualType"`
}

// astLoc represents the source location of a node.
// includedFrom is non-nil when the node comes from an included file.
type astLoc struct {
	IncludedFrom *astIncludedFrom `json:"includedFrom"`
}

type astIncludedFrom struct {
	File string `json:"file"`
}

// IsFromStdin reports whether this location is directly from the stdin TU
// (i.e., the function is defined in one of the headers we directly included).
// Functions with includedFrom.file = "<stdin>" are in our target headers.
// Functions with includedFrom.file = a real path are from transitive includes.
func (l astLoc) IsFromStdin() bool {
	if l.IncludedFrom == nil {
		return false
	}
	return l.IncludedFrom.File == "<stdin>"
}

// Package model defines the core domain types used throughout the mlxgen
// code generation pipeline.
package model

// Function represents a C++ function extracted from an MLX header.
type Function struct {
	Name       string
	Namespace  string
	ReturnType string
	Params     []Param
	Defaults   []string
	// Variant is the suffix assigned to disambiguate overloads (e.g. "axes", "axis", "").
	// Empty string means no suffix is appended.
	Variant string
	// UseDefaults signals that parameters with a default value should be omitted
	// from the generated C signature.
	UseDefaults bool
}

// Param is a single parameter in a function signature.
type Param struct {
	Name    string
	Type    string
	Default string
}

// Enum represents a C++ enum extracted from an MLX header.
type Enum struct {
	Name      string
	Namespace string
	Values    []string
}

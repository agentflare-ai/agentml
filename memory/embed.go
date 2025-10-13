//go:build !generate
// +build !generate

package memory

import (
	_ "embed"
)

// GraphExtension returns the appropriate extension for the current platform
//
//go:embed graph_extension.so
var GraphExtension []byte

// VecExtension returns the appropriate extension for the current platform
//
//go:embed vec_extension.so
var VecExtension []byte

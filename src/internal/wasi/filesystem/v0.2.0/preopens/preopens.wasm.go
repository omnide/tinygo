// Code generated by wit-bindgen-go. DO NOT EDIT.

package preopens

import (
	"internal/cm"
)

// This file contains wasmimport and wasmexport declarations for "wasi:filesystem@0.2.0".

//go:wasmimport wasi:filesystem/preopens@0.2.0 get-directories
//go:noescape
func wasmimport_GetDirectories(result *cm.List[cm.Tuple[Descriptor, string]])

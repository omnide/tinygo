//go:build wasip2

package syscall

import (
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cm"
)

type __wasi_list_tuple struct {
	ptr unsafe.Pointer
	len uint32
}

// TODO(ydnar): replace __wasi_string with Go string
type __wasi_string struct {
	data unsafe.Pointer
	len  uint32
}

// TODO(ydnar): replace _string with Go string
type _string struct {
	ptr    *byte
	length uintptr
}

//go:wasmimport wasi:cli/environment@0.2.0-rc-2023-12-05 get-environment
func __wasi_cli_environment_get_environment() cm.List[cm.Tuple[string, string]]

func Environ() []string {
	var env []string
	wasiEnv := __wasi_cli_environment_get_environment().Slice()
	for _, kv := range wasiEnv {
		env = append(env, kv.V0+"="+kv.V1)
	}
	return env
}

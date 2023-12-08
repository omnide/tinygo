//go:build wasip2

package syscall

import (
	"unsafe"
)

type __wasi_list_tuple struct {
	ptr unsafe.Pointer
	len uint32
}

type __wasi_tuple_string_string struct {
	first  __wasi_string
	second __wasi_string
}

type __wasi_string struct {
	data unsafe.Pointer
	len  uint32
}

type _string struct {
	ptr    *byte
	length uintptr
}

//go:wasmimport wasi:cli/environment@0.2.0-rc-2023-11-10 get-environment
func __wasi_cli_environment_get_environment() __wasi_list_tuple

func Environ() []string {
	var envs []string
	list_tuple := __wasi_cli_environment_get_environment()
	envs = make([]string, list_tuple.len)
	ptr := list_tuple.ptr
	for i := uint32(0); i < list_tuple.len; i++ {
		tuple := *(*__wasi_tuple_string_string)(ptr)
		first, second := tuple.first, tuple.second
		envKey := _string{
			(*byte)(first.data), uintptr(first.len),
		}
		envValue := _string{
			(*byte)(second.data), uintptr(second.len),
		}
		envs[i] = *(*string)(unsafe.Pointer(&envKey)) + "=" + *(*string)(unsafe.Pointer(&envValue))
		ptr = unsafe.Add(ptr, unsafe.Sizeof(__wasi_tuple_string_string{}))
	}

	return envs
}

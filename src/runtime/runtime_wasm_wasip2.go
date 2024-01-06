//go:build tinygo.wasm && (wasi || wasip1) && wasip2

package runtime

import (
	"unsafe"
)

type timeUnit int64

// libc constructors
//
//export __wasm_call_ctors
func __wasm_call_ctors()

//export wasi:cli/run@0.2.0-rc-2023-12-05#run
func __wasi_cli_run_run() uint32 {
	_start()
	return 0
}

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
}

// Read the command line arguments from WASI.
// For example, they can be passed to a program with wasmtime like this:
//
//	wasmtime run ./program.wasm arg1 arg2
func init() {
	__wasm_call_ctors()
}

var args []string

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	if args == nil {
		list_string := __wasi_cli_environment_get_arguments()
		args = make([]string, list_string.len)
		ptr := list_string.ptr
		for i := uint32(0); i < list_string.len; i++ {
			sbuf := *(*__wasi_string)(ptr)
			argString := _string{
				(*byte)(sbuf.data), uintptr(sbuf.len),
			}
			args[i] = *(*string)(unsafe.Pointer(&argString))
			ptr = unsafe.Add(ptr, unsafe.Sizeof(__wasi_string{}))
		}
	}
	return args
}

//export cabi_realloc
func cabi_realloc(ptr, oldsize, align, newsize unsafe.Pointer) unsafe.Pointer {
	return realloc(ptr, uintptr(newsize))
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

const timePrecisionNanoseconds = 1000 // TODO: how can we determine the appropriate `precision`?

var ()

//go:wasmimport wasi:clocks/monotonic-clock@0.2.0-rc-2023-11-10 subscribe-duration
func __wasi_clocks_monotonic_subscribe_duration(d uint64) uint32

//go:wasmimport wasi:io/poll@0.2.0-rc-2023-11-10 [method]pollable.block
func __wasi_io_poll_pollable_block(pollable uint32)

func sleepTicks(d timeUnit) {
	p := __wasi_clocks_monotonic_subscribe_duration(uint64(d))
	__wasi_io_poll_pollable_block(p)
}

func ticks() timeUnit {
	var now __wasi_clocks_wallclock_datetime
	__wasi_clocks_wallclock_now(&now)
	nano := now.seconds*1e9 + uint64(now.nanoseconds)
	return timeUnit(nano)
}

type __wasi_list_string struct {
	ptr unsafe.Pointer
	len uint32
}

type __wasi_string struct {
	data unsafe.Pointer
	len  uint32
}

//go:wasmimport wasi:cli/environment@0.2.0-rc-2023-12-05 get-arguments
func __wasi_cli_environment_get_arguments() __wasi_list_string

type __wasi_clocks_wallclock_datetime struct {
	seconds     uint64
	nanoseconds uint32
}

//go:wasmimport wasi:clocks/wall-clock@0.2.0-rc-2023-11-10 now
func __wasi_clocks_wallclock_now(t *__wasi_clocks_wallclock_datetime)

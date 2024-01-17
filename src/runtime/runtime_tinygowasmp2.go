//go:build tinygo.wasm && wasip2

package runtime

import (
	"internal/wasm/wasi/clocks/monotonicclock"
	"internal/wasm/wasi/clocks/wallclock"
	"unsafe"
)

type __wasi_io_stream_list struct {
	buf unsafe.Pointer
	len uint32
}

//go:wasmimport wasi:cli/stdout@0.2.0-rc-2023-12-05 get-stdout
func __wasi_cli_stdout_get_stdout() int32

//go:wasmimport wasi:io/streams@0.2.0-rc-2023-11-10 [method]output-stream.blocking-write-and-flush
func __wasi_io_streams_blocking_write_and_flush(stream int32, buf __wasi_io_stream_list, err unsafe.Pointer)

//go:wasmimport wasi:cli/exit@0.2.0-rc-2023-12-05 exit
func __wasi_exit_exit(status uint32)

const (
	putcharBufferSize = 120
)

// Using global variables to avoid heap allocation.
var (
	stdout                 = __wasi_cli_stdout_get_stdout()
	putcharBuffer          = [putcharBufferSize]byte{}
	putcharPosition uint32 = 0
	putcharList            = __wasi_io_stream_list{
		buf: unsafe.Pointer(&putcharBuffer[0]),
	}
)

func putchar(c byte) {
	putcharBuffer[putcharPosition] = c
	putcharPosition++

	var err uint64

	if c == '\n' || putcharPosition >= putcharBufferSize {
		putcharList.len = putcharPosition
		// TODO(dgryski): need actual ptr for error return
		__wasi_io_streams_blocking_write_and_flush(stdout, putcharList, unsafe.Pointer(&err))
		putcharPosition = 0
	}
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	now := wallclock.Now()
	sec = int64(now.Seconds)
	nsec = int32(now.Nanoseconds)
	mono = int64(monotonicclock.Now())
	return
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	__wasi_exit_exit(uint32(code))
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

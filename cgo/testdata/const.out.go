package main

import "syscall"
import "unsafe"

var _ unsafe.Pointer

//go:linkname C.CString runtime.cgo_CString
func C.CString(string) *C.char

//go:linkname C.GoString runtime.cgo_GoString
func C.GoString(*C.char) string

//go:linkname C.__GoStringN runtime.cgo_GoStringN
func C.__GoStringN(*C.char, uintptr) string

func C.GoStringN(cstr *C.char, length C.int) string {
	return C.__GoStringN(cstr, uintptr(length))
}

//go:linkname C.__GoBytes runtime.cgo_GoBytes
func C.__GoBytes(unsafe.Pointer, uintptr) []byte

func C.GoBytes(ptr unsafe.Pointer, length C.int) []byte {
	return C.__GoBytes(ptr, uintptr(length))
}

//go:linkname C.__CBytes runtime.cgo_CBytes
func C.__CBytes([]byte) unsafe.Pointer

func C.CBytes(b []byte) unsafe.Pointer {
	return C.__CBytes(b)
}

//go:linkname C.__get_errno_num runtime.cgo_errno
func C.__get_errno_num() uintptr

func C.__get_errno() error {
	return syscall.Errno(C.__get_errno_num())
}

type (
	C.char      uint8
	C.schar     int8
	C.uchar     uint8
	C.short     int16
	C.ushort    uint16
	C.int       int32
	C.uint      uint32
	C.long      int32
	C.ulong     uint32
	C.longlong  int64
	C.ulonglong uint64
)

const C.foo = 3
const C.bar = C.foo
const C.unreferenced = 4
const C.referenced = C.unreferenced
const C.fnlike_val = 5
const C.square_val = (20 * 20)
const C.add_val = (3 + 5)

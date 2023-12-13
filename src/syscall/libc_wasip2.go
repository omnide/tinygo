//go:build wasip2

// mini libc wrapping wasi preview2 calls in a libc api

package syscall

import (
	"strings"
	"unsafe"
)

//go:export strlen
func strlen(ptr unsafe.Pointer) uintptr {
	if ptr == nil {
		return 0
	}
	var i uintptr
	for p := (*byte)(ptr); *p != 0; p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1)) {
		i++
	}
	return i
}

// ssize_t write(int fd, const void *buf, size_t count)
//
//go:export write
func write(fd int32, buf *byte, count uint) int {
	return 0
}

// char *getenv(const char *name);
//
//go:export getenv
func getenv(name *byte) *byte {
	return nil
}

// ssize_t read(int fd, void *buf, size_t count);
//
//go:export read
func read(fd int32, buf *byte, count uint) int {
	return 0
}

// ssize_t pread(int fd, void *buf, size_t count, off_t offset);
//
//go:export pread
func pread(fd int32, buf *byte, count uint, offset int64) int {
	return 0
}

// ssize_t pwrite(int fd, void *buf, size_t count, off_t offset);
//
//go:export pwrite
func pwrite(fd int32, buf *byte, count uint, offset int64) int {
	return 0
}

// ssize_t lseek(int fd, off_t offset, int whence);
//
//go:export lseek
func lseek(fd int32, offset int64, whence int) int64 {
	return 0
}

// int close(int fd)
//
//go:export close
func close(fd int32) int32 {
	return 0
}

// int dup(int fd)
//
//go:export dup
func dup(fd int32) int32 {
	return 0
}

// void *mmap(void *addr, size_t length, int prot, int flags, int fd, off_t offset);
//
//go:export mmap
func mmap(addr unsafe.Pointer, length uintptr, prot, flags, fd int32, offset uintptr) unsafe.Pointer {
	return nil
}

// int munmap(void *addr, size_t length);
//
//go:export munmap
func munmap(addr unsafe.Pointer, length uintptr) int32 {
	return 0
}

// int mprotect(void *addr, size_t len, int prot);
//
//go:export mprotect
func mprotect(addr unsafe.Pointer, len uintptr, prot int32) int32 {
	return 0
}

// int chdir(const char *pathname, mode_t mode);
//
//go:export chdir
func chdir(pathname *byte) int32 {
	return 0
}

// int chmod(const char *pathname, mode_t mode);
//
//go:export chmod
func chmod(pathname *byte, mode uint32) int32 {
	return 0
}

// int mkdir(const char *pathname, mode_t mode);
//
//go:export mkdir
func mkdir(pathname *byte, mode uint32) int32 {
	return 0
}

// int rmdir(const char *pathname);
//
//go:export rmdir
func rmdir(pathname *byte) int32 {
	return 0
}

// int rename(const char *from, *to);
//
//go:export rename
func rename(from, to *byte) int32 {
	return 0
}

// int symlink(const char *from, *to);
//
//go:export symlink
func symlink(from, to *byte) int32 {
	return 0
}

// int fsync(int fd);
//
//go:export fsync
func fsync(fd int32) int32 {
	return 0

}

// ssize_t readlink(const char *path, void *buf, size_t count);
//
//go:export readlink
func readlink(path *byte, buf *byte, count uint) int {
	return 0
}

// int unlink(const char *pathname);
//
//go:export unlink
func unlink(pathname *byte) int32 {
	return 0
}

// int getpagesize(void);
//
//go:export getpagesize
func getpagesize() int {
	return 0

}

// int stat(const char *path, struct stat * buf);
//
//go:export stat
func stat(pathname *byte, ptr unsafe.Pointer) int32 {
	return 0

}

// int fstat(int fd, struct stat * buf);
//
//go:export fstat
func fstat(fd int32, ptr unsafe.Pointer) int32 {
	return 0

}

// int lstat(const char *path, struct stat * buf);
//
//go:export lstat
func lstat(pathname *byte, ptr unsafe.Pointer) int32 {
	return 0
}

func init() {
	populatePreopens()
}

var __wasi_cwd_descriptor __wasi_filesystem_descriptor

var __wasi_filesystem_preopens map[string]__wasi_filesystem_descriptor

func populatePreopens() {
	println("preopens")
	list_tuple := __wasi_filesystem_preopens_get_directories()
	dirs := make(map[string]__wasi_filesystem_descriptor, list_tuple.len)
	ptr := list_tuple.ptr
	for i := uint32(0); i < list_tuple.len; i++ {
		tuple := *(*__wasi_tuple_descriptor_string)(ptr)
		descriptor, pathWasiStr := tuple.first, tuple.second
		pathStr := _string{
			(*byte)(pathWasiStr.data), uintptr(pathWasiStr.len),
		}
		path := *(*string)(unsafe.Pointer(&pathStr))
		dirs[path] = descriptor
		println("path:", path, "descriptor:", descriptor)
		if path == "." {
			__wasi_cwd_descriptor = descriptor
		}
		ptr = unsafe.Add(ptr, unsafe.Sizeof(__wasi_tuple_descriptor_string{}))
	}
	__wasi_filesystem_preopens = dirs
}

//go:wasmimport wasi:filesystem/preopens@0.2.0-rc-2023-11-10 get-directories
func __wasi_filesystem_preopens_get_directories() __wasi_list_tuple

type __wasi_filesystem_descriptor int32

type __wasi_filesystem_error int16

type __wasi_tuple_descriptor_string struct {
	first  __wasi_filesystem_descriptor
	second __wasi_string
}

// int open(const char *pathname, int flags, mode_t mode);
//
//go:export open
func open(pathname *byte, flags int32, mode uint32) int32 {
	pathLen := strlen(unsafe.Pointer(pathname))
	pathStr := _string{
		pathname, uintptr(pathLen),
	}
	path := *(*string)(unsafe.Pointer(&pathStr))
	// TODO(dgryski): This path searching logic isn't right; need to strip out path prefix when we find it I think
	var dir __wasi_filesystem_descriptor = __wasi_cwd_descriptor
	for k, v := range __wasi_filesystem_preopens {
		println("preopen:", k, v)

		if strings.HasPrefix(path, k) {
			println("matched, using dir=", v)
			dir = v
			break
		}
	}

	pathWasiStr := __wasi_string{
		unsafe.Pointer(pathname), uint32(pathLen),
	}
	var ret __wasi_result_descriptor_error
	__wasi_filesystem_types_descriptor_open_at(dir, 0 /* path_flags */, pathWasiStr.data, int32(pathWasiStr.len), 0 /* open flags */, 0 /* flags */, &ret)
	if ret.isErr {
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(ret.val)))
		return -1
	}
	return int32(ret.val)
}

const (
	__wasi_filesystem_error_access                __wasi_filesystem_error = iota + 1 /// Permission denied, similar to `EACCES` in POSIX.
	__wasi_filesystem_error_would_block                                              /// Resource unavailable, or operation would block, similar to `EAGAIN` and `EWOULDBLOCK` in POSIX.
	__wasi_filesystem_error_already                                                  /// Connection already in progress, similar to `EALREADY` in POSIX.
	__wasi_filesystem_error_bad_descriptor                                           /// Bad descriptor, similar to `EBADF` in POSIX.
	__wasi_filesystem_error_busy                                                     /// Device or resource busy, similar to `EBUSY` in POSIX.
	__wasi_filesystem_error_deadlock                                                 /// Resource deadlock would occur, similar to `EDEADLK` in POSIX.
	__wasi_filesystem_error_quota                                                    /// Storage quota exceeded, similar to `EDQUOT` in POSIX.
	__wasi_filesystem_error_exist                                                    /// File exists, similar to `EEXIST` in POSIX.
	__wasi_filesystem_error_file_too_large                                           /// File too large, similar to `EFBIG` in POSIX.
	__wasi_filesystem_error_illegal_byte_sequence                                    /// Illegal byte sequence, similar to `EILSEQ` in POSIX.
	__wasi_filesystem_error_in_progress                                              /// Operation in progress, similar to `EINPROGRESS` in POSIX.
	__wasi_filesystem_error_interrupted                                              /// Interrupted function, similar to `EINTR` in POSIX.
	__wasi_filesystem_error_invalid                                                  /// Invalid argument, similar to `EINVAL` in POSIX.
	__wasi_filesystem_error_io                                                       /// I/O error, similar to `EIO` in POSIX.
	__wasi_filesystem_error_is_directory                                             /// Is a directory, similar to `EISDIR` in POSIX.
	__wasi_filesystem_error_loop                                                     /// Too many levels of symbolic links, similar to `ELOOP` in POSIX.
	__wasi_filesystem_error_too_many_links                                           /// Too many links, similar to `EMLINK` in POSIX.
	__wasi_filesystem_error_message_size                                             /// Message too large, similar to `EMSGSIZE` in POSIX.
	__wasi_filesystem_error_name_too_long                                            /// Filename too long, similar to `ENAMETOOLONG` in POSIX.
	__wasi_filesystem_error_no_device                                                /// No such device, similar to `ENODEV` in POSIX.
	__wasi_filesystem_error_no_entry                                                 /// No such file or directory, similar to `ENOENT` in POSIX.
	__wasi_filesystem_error_no_lock                                                  /// No locks available, similar to `ENOLCK` in POSIX.
	__wasi_filesystem_error_insufficient_memory                                      /// Not enough space, similar to `ENOMEM` in POSIX.
	__wasi_filesystem_error_insufficient_space                                       /// No space left on device, similar to `ENOSPC` in POSIX.
	__wasi_filesystem_error_not_directory                                            /// Not a directory or a symbolic link to a directory, similar to `ENOTDIR` in POSIX.
	__wasi_filesystem_error_not_empty                                                /// Directory not empty, similar to `ENOTEMPTY` in POSIX.
	__wasi_filesystem_error_not_recoverable                                          /// State not recoverable, similar to `ENOTRECOVERABLE` in POSIX.
	__wasi_filesystem_error_unsupported                                              /// Not supported, similar to `ENOTSUP` and `ENOSYS` in POSIX.
	__wasi_filesystem_error_no_tty                                                   /// Inappropriate I/O control operation, similar to `ENOTTY` in POSIX.
	__wasi_filesystem_error_no_such_device                                           /// No such device or address, similar to `ENXIO` in POSIX.
	__wasi_filesystem_error_overflow                                                 /// Value too large to be stored in data type, similar to `EOVERFLOW` in POSIX.
	__wasi_filesystem_error_not_permitted                                            /// Operation not permitted, similar to `EPERM` in POSIX.
	__wasi_filesystem_error_pipe                                                     /// Broken pipe, similar to `EPIPE` in POSIX.
	__wasi_filesystem_error_read_only                                                /// Read_only file system, similar to `EROFS` in POSIX.
	__wasi_filesystem_error_invalid_seek                                             /// Invalid seek, similar to `ESPIPE` in POSIX.
	__wasi_filesystem_error_text_file_busy                                           /// Text file busy, similar to `ETXTBSY` in POSIX.
	__wasi_filesystem_error_cross_device                                             /// Cross_device link, similar to `EXDEV` in POSIX.
)

func __wasi_filesystem_err_to_errno(err __wasi_filesystem_error) Errno {
	switch err {
	case __wasi_filesystem_error_access:
		return EACCES
	case __wasi_filesystem_error_would_block:
		return EAGAIN
	case __wasi_filesystem_error_already:
		return EALREADY
	case __wasi_filesystem_error_bad_descriptor:
		return EBADF
	case __wasi_filesystem_error_busy:
		return EBUSY
	case __wasi_filesystem_error_deadlock:
		return EDEADLK
	case __wasi_filesystem_error_quota:
		return EDQUOT
	case __wasi_filesystem_error_exist:
		return EEXIST
	case __wasi_filesystem_error_file_too_large:
		return EFBIG
	case __wasi_filesystem_error_illegal_byte_sequence:
		return EILSEQ
	case __wasi_filesystem_error_in_progress:
		return EINPROGRESS
	case __wasi_filesystem_error_interrupted:
		return EINTR
	case __wasi_filesystem_error_invalid:
		return EINVAL
	case __wasi_filesystem_error_io:
		return EIO
	case __wasi_filesystem_error_is_directory:
		return EISDIR
	case __wasi_filesystem_error_loop:
		return ELOOP
	case __wasi_filesystem_error_too_many_links:
		return EMLINK
	case __wasi_filesystem_error_message_size:
		return EMSGSIZE
	case __wasi_filesystem_error_name_too_long:
		return ENAMETOOLONG
	case __wasi_filesystem_error_no_device:
		return ENODEV
	case __wasi_filesystem_error_no_entry:
		return ENOENT
	case __wasi_filesystem_error_no_lock:
		return ENOLCK
	case __wasi_filesystem_error_insufficient_memory:
		return ENOMEM
	case __wasi_filesystem_error_insufficient_space:
		return ENOSPC
	case __wasi_filesystem_error_not_directory:
		return ENOTDIR
	case __wasi_filesystem_error_not_empty:
		return ENOTEMPTY
	case __wasi_filesystem_error_not_recoverable:
		return ENOTRECOVERABLE
	case __wasi_filesystem_error_unsupported:
		return ENOSYS
	case __wasi_filesystem_error_no_tty:
		return ENOTTY
	case __wasi_filesystem_error_no_such_device:
		return ENXIO
	case __wasi_filesystem_error_overflow:
		return EOVERFLOW
	case __wasi_filesystem_error_not_permitted:
		return EPERM
	case __wasi_filesystem_error_pipe:
		return EPIPE
	case __wasi_filesystem_error_read_only:
		return EROFS
	case __wasi_filesystem_error_invalid_seek:
		return ESPIPE
	case __wasi_filesystem_error_text_file_busy:
		return ETXTBSY
	case __wasi_filesystem_error_cross_device:
		return EXDEV

	}
	return Errno(err)
}

type __wasi_result_descriptor_error struct {
	isErr bool
	val   int32 // fd or err
}

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.open-at
func __wasi_filesystem_types_descriptor_open_at(dir __wasi_filesystem_descriptor, path_flags int32, path_ptr unsafe.Pointer, path_len int32, open_flags int32, flags int32, ret *__wasi_result_descriptor_error)

// DIR *fdopendir(int);
//
//go:export fdopendir
func fdopendir(fd int32) unsafe.Pointer {
	return nil
}

// int fdclosedir(DIR *);
//
//go:export fdclosedir
func fdclosedir(unsafe.Pointer) int32 {
	return 0
}

// struct dirent *readdir(DIR *);
//
//go:export readdir
func readdir(unsafe.Pointer) *Dirent {
	return nil
}

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
	if -1 <= fd && fd <= Stderr {
		return writeStdout(fd, buf, count, 0)
	}

	stream, ok := wasiStreams[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	if stream.d == -1 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	n := pwrite(fd, buf, count, int64(stream.offset))
	if n == -1 {
		return -1
	}
	stream.offset += int64(n)
	return int(n)
}

// ssize_t read(int fd, void *buf, size_t count);
//
//go:export read
func read(fd int32, buf *byte, count uint) int {
	if -1 <= fd && fd <= Stderr {
		return readStdin(fd, buf, count, 0)
	}

	stream, ok := wasiStreams[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	if stream.d == -1 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	n := pread(fd, buf, count, int64(stream.offset))
	if n == -1 {
		// error during pread
		return -1
	}
	stream.offset += int64(n)
	return int(n)
}

// char *getenv(const char *name);
//
//go:export getenv
func getenv(name *byte) *byte {
	return nil
}

type __wasi_filesystem_descriptor int32

type __wasi_io_streams_input_stream int32

type __wasi_io_streams_output_stream int32

type __wasi_io_stream int32

// At the moment, each time we have a file read or write we create a new stream.  Future implementations
// could change the current in or out file stream lazily.  We could do this by tracking input and output
// offsets individually, and if they don't match the current main offset, reopen the file stream at that location.

type wasiFile struct {
	d      __wasi_filesystem_descriptor
	oflag  int32 // orignal open flags: O_RDONLY, O_WRONLY, O_RDWR
	offset int64 // current fd offset; updated with each read/write
}

// Need to figure out which system calls we're using:
//   stdin/stdout/stderr want streams, so we use stream read/write
//   but for regular files we can use the descriptor and explicitly write a buffer to the offset?
//   The mismatch comes from trying to combine these.

var wasiStreams map[int32]*wasiFile = make(map[int32]*wasiFile)
var nextLibcFd = int32(Stderr) + 1

var wasiErrno error

func readStdin(fd int32, buf *byte, count uint, offset int64) int {
	if fd != 0 {
		// TODO(dgryski): Not sure this will place nicely wit `dup()` and processes which close stdin
		panic("non-stdin passed to readStdin")
	}
	wasifd := __wasi_cli_stdout_get_stdin()
	if offset != 0 {
		libcErrno = uintptr(EINVAL)
		return -1
	}

	var ret [12]byte
	libcErrno = 0
	__wasi_io_streams_method_input_stream_blocking_read(wasifd, int64(count), unsafe.Pointer(&ret))
	result := (*__wasi_result)(unsafe.Pointer(&ret))
	if result.isErr {
		stream_error_tag := (*__wasi_io_stream_error_tag)(unsafe.Add(unsafe.Pointer(&ret), 4))
		switch stream_error_tag.tag {
		case 0: // last operation failed
			var0 := (*__wasi_io_stream_error_variant_last_operation_failed)(unsafe.Add(unsafe.Pointer(&ret), 8))
			wasiErrno = __wasi_io_error_to_error(var0.err)
			libcErrno = uintptr(EWASIERROR)
			return -1

		case 1: // closed == EOF was reached
			libcErrno = 0
		}
	}

	list_u8 := (*__wasi_list_u8)(unsafe.Add(unsafe.Pointer(&ret), 4))

	return int(list_u8.len)
}

func writeStdout(fd int32, buf *byte, count uint, offset int64) int {
	var stream __wasi_io_streams_output_stream
	switch fd {
	case 1:
		stream = __wasi_cli_stdout_get_stdout()
	case 2:
		stream = __wasi_cli_stdout_get_stderr()
	default:
		panic("non-stdout/err passed to writeStdout")
	}

	if offset != 0 {
		libcErrno = uintptr(EINVAL)
		return -1
	}

	ptr := unsafe.Pointer(buf)
	var remaining = count

	// The blocking-write-and-flush call allows a maximum of 4096 bytes at a time.
	// We loop here by instead of doing subscribe/check-write/poll-one/write by hand.
	for remaining > 0 {
		len := uint(4096)
		if len > remaining {
			len = remaining
		}
		list_u8 := __wasi_list_u8{
			data: ptr, len: uintptr(len),
		}

		var ret [12]byte
		__wasi_io_streams_method_output_stream_blocking_write_and_flush(stream, list_u8.data, list_u8.len, unsafe.Pointer(&ret))
		result := (*__wasi_result)(unsafe.Pointer(&ret))
		if result.isErr {
			stream_error_tag := (*__wasi_io_stream_error_tag)(unsafe.Add(unsafe.Pointer(&ret), 4))
			switch stream_error_tag.tag {
			case 0: // last operation failed
				var0 := (*__wasi_io_stream_error_variant_last_operation_failed)(unsafe.Add(unsafe.Pointer(&ret), 8))
				wasiErrno = __wasi_io_error_to_error(var0.err)
				libcErrno = uintptr(EWASIERROR)
				return -1

			case 1: // closed == EOF was reached
				libcErrno = 0
			}
		}
		ptr = unsafe.Add(ptr, list_u8.len)
		remaining -= uint(list_u8.len)
	}

	return int(count)
}

//go:wasmimport wasi:cli/stdin@0.2.0-rc-2023-11-10 get-stdin
func __wasi_cli_stdout_get_stdin() __wasi_io_streams_input_stream

//go:wasmimport wasi:cli/stdout@0.2.0-rc-2023-11-10 get-stdout
func __wasi_cli_stdout_get_stdout() __wasi_io_streams_output_stream

//go:wasmimport wasi:cli/stderr@0.2.0-rc-2023-11-10 get-stderr
func __wasi_cli_stdout_get_stderr() __wasi_io_streams_output_stream

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)

type WASIError string

func (e WASIError) Error() string {
	return string(e)
}

func __wasi_io_error_to_error(err int32) error {
	var debug __wasi_string
	__wasi_io_error_error_to_debug_string(err, &debug)

	s := *(*string)(unsafe.Pointer(&debug))
	return WASIError(s)
}

//go:wasmimport wasi:io/error@0.2.0-rc-2023-11-10 [method]error.to-debug-string
func __wasi_io_error_error_to_debug_string(err int32, s *__wasi_string)

type __wasi_result struct {
	isErr bool
}

type __wasi_list_u8 struct {
	data unsafe.Pointer
	len  uintptr
}

type __wasi_io_stream_error_tag struct {
	tag uint8
}

type __wasi_io_stream_error_variant_last_operation_failed struct {
	err int32
}

type __wasi_io_stream_error_variant_closed struct {
}

//go:wasmimport wasi:io/streams@0.2.0-rc-2023-11-10 [method]input-stream.blocking-read
func __wasi_io_streams_method_input_stream_blocking_read(self __wasi_io_streams_input_stream, len int64, ret unsafe.Pointer)

//go:wasmimport wasi:io/streams@0.2.0-rc-2023-11-10 [method]output-stream.blocking-write-and-flush
func __wasi_io_streams_method_output_stream_blocking_write_and_flush(self __wasi_io_streams_output_stream, list_u8_data unsafe.Pointer, list_u8_len uintptr, ret unsafe.Pointer)

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.read
func __wasi_filesystem_types_method_descriptor_read(self __wasi_filesystem_descriptor, len uint64, offset uint64, ret unsafe.Pointer)

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.write
func __wasi_filesystem_types_method_descriptor_write(self __wasi_filesystem_descriptor, list_u8_data unsafe.Pointer, list_u8_len int32, offset int64, ret unsafe.Pointer)

type __wasi_tuple_list_u8_bool struct {
	list_u8 __wasi_list_u8
	b       bool
}

// ssize_t pread(int fd, void *buf, size_t count, off_t offset);
//
//go:export pread
func pread(fd int32, buf *byte, count uint, offset int64) int {
	// TODO(dgryski): Need to be consistent about all these checks; EBADF/EINVAL/... ?

	if -1 < fd && fd <= Stderr {
		if fd == Stdin {
			return readStdin(fd, buf, count, offset)
		}

		// stdout/stderr not open for reading
		libcErrno = uintptr(EBADF)
		return -1
	}

	streams, ok := wasiStreams[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	if streams.d == -1 {
		libcErrno = uintptr(EBADF)
		return -1
	}
	if streams.oflag&O_RDONLY == 0 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	var ret [unsafe.Sizeof(__wasi_result{}) + unsafe.Sizeof(__wasi_tuple_list_u8_bool{})]byte
	__wasi_filesystem_types_method_descriptor_read(streams.d, uint64(count), uint64(offset), unsafe.Pointer(&ret))
	result := (*__wasi_result)(unsafe.Pointer(&ret))
	if result.isErr {
		errCode := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 4))
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(errCode)))
		return -1
	}

	tuple := (*__wasi_tuple_list_u8_bool)(unsafe.Add(unsafe.Pointer(&ret), 4))
	memcpy(unsafe.Pointer(buf), tuple.list_u8.data, tuple.list_u8.len)
	// TODO(dgryski): EOF bool is ignored?

	return int(tuple.list_u8.len)
}

// ssize_t pwrite(int fd, void *buf, size_t count, off_t offset);
//
//go:export pwrite
func pwrite(fd int32, buf *byte, count uint, offset int64) int {
	// TODO(dgryski): Need to be consistent about all these checks; EBADF/EINVAL/... ?
	if -1 <= fd && fd <= Stderr {
		return writeStdout(fd, buf, count, offset)
	}

	streams, ok := wasiStreams[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	if streams.d == -1 {
		libcErrno = uintptr(EBADF)
		return -1
	}
	if streams.oflag&O_WRONLY == 0 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	list_u8 := __wasi_list_u8{
		data: unsafe.Pointer(buf), len: uintptr(count),
	}

	var ret [unsafe.Sizeof(__wasi_result{}) + unsafe.Sizeof(uint64(0))]byte
	__wasi_filesystem_types_method_descriptor_write(streams.d, list_u8.data, int32(list_u8.len), offset, unsafe.Pointer(&ret))
	result := (*__wasi_result)(unsafe.Pointer(&ret))
	if result.isErr {
		// TODO(dgryski):
		errCode := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 8))
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(errCode)))
		return -1
	}

	size := *(*uint64)(unsafe.Add(unsafe.Pointer(&ret), 8))
	return int(size)
}

// ssize_t lseek(int fd, off_t offset, int whence);
//
//go:export lseek
func lseek(fd int32, offset int64, whence int) int64 {
	streams, ok := wasiStreams[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}

	switch whence {
	case 0: // SEEK_SET
		streams.offset = offset
	case 1: // SEEK_CUR
		streams.offset += offset
	case 2: // SEEK_END
		// TODO(dgryski): query current file size, then add offset
	}

	return int64(streams.offset)
}

// int close(int fd)
//
//go:export close
func close(fd int32) int32 {
	streams, ok := wasiStreams[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}

	if streams.d != -1 {
		__wasi_filesystem_resource_drop_descriptor(streams.d)
	}

	delete(wasiStreams, fd)

	return 0
}

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [resource-drop]descriptor
func __wasi_filesystem_resource_drop_descriptor(d __wasi_filesystem_descriptor)

//go:wasmimport wasi:io/streams@0.2.0-rc-2023-11-10 [resource-drop]input-stream
func __wasi_io_streams_resource_drop_input_stream(stream __wasi_io_streams_input_stream)

//go:wasmimport wasi:io/streams@0.2.0-rc-2023-11-10 [resource-drop]output-stream
func __wasi_io_streams_resource_drop_output_stream(stream __wasi_io_streams_output_stream)

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
	pathLen := strlen(unsafe.Pointer(pathname))
	pathStr := _string{
		pathname, uintptr(pathLen),
	}
	path := *(*string)(unsafe.Pointer(&pathStr))
	dir, path := findPreopenForPath(path)

	pathWasiStr := __wasi_string{
		unsafe.Pointer(pathname), uint32(pathLen),
	}

	var ret __wasi_result_filesystem_descriptor_stat_error
	__wasi_filesystem_types_method_descriptor_stat_at(dir, __wasi_filesystem_path_flag_symlink_follow, pathWasiStr, &ret)
	if ret.isErr {
		error_code := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 8))
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(error_code)))
		return -1

	}

	stat := (*Stat_t)(ptr)
	setStatFromWASIStat(stat, &ret.stat)

	return 0
}

type __wasi_filesystem_descriptor_linkcount uint64
type __wasi_filesystem_descriptor_filesize uint64

type __wasi_clocks_wallclock_datetime struct {
	seconds uint64
	nano    uint32
}

type __wasi_option_datetime struct {
	isSome bool
	t      __wasi_clocks_wallclock_datetime
}

type __wasi_filesystem_descriptor_stat struct {
	typ                         __wasi_filesystem_descriptor_type
	link_count                  __wasi_filesystem_descriptor_linkcount
	filesize                    __wasi_filesystem_descriptor_filesize
	data_access_timestamp       __wasi_option_datetime
	data_modification_timestamp __wasi_option_datetime
	status_change_timestamp     __wasi_option_datetime
}

type __wasi_filesystem_descriptor_type uint8

const (
	__wasi_filesystem_descriptor_type_unknown          __wasi_filesystem_descriptor_type = iota /// The type of the descriptor or file is unknown or is different from any of the other types specified.
	__wasi_filesystem_descriptor_type_block_device                                              /// The descriptor refers to a block device inode.
	__wasi_filesystem_descriptor_type_character_device                                          /// The descriptor refers to a character device inode.
	__wasi_filesystem_descriptor_type_directory                                                 /// The descriptor refers to a directory inode.
	__wasi_filesystem_descriptor_type_fifo                                                      /// The descriptor refers to a named pipe.
	__wasi_filesystem_descriptor_type_symbolic_link                                             /// The file refers to a symbolic link inode.
	__wasi_filesystem_descriptor_type_regular_file                                              /// The descriptor refers to a regular file inode.
	__wasi_filesystem_descriptor_type_socket                                                    /// The descriptor refers to a socket.
)

type __wasi_result_filesystem_descriptor_stat_error struct {
	isErr bool
	stat  __wasi_filesystem_descriptor_stat
}

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.stat
func __wasi_filesystem_types_method_descriptor_stat(d __wasi_filesystem_descriptor, ret *__wasi_result_filesystem_descriptor_stat_error)

// int fstat(int fd, struct stat * buf);
//
//go:export fstat
func fstat(fd int32, ptr unsafe.Pointer) int32 {
	if -1 < fd && fd <= Stderr {
		// TODO(dgryski): fill in stat buffer for stdin etc
		return -1
	}

	stream, ok := wasiStreams[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}
	if stream.d == -1 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	var ret __wasi_result_filesystem_descriptor_stat_error
	__wasi_filesystem_types_method_descriptor_stat(stream.d, &ret)
	if ret.isErr {
		error_code := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 8))
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(error_code)))
		return -1
	}

	stat := (*Stat_t)(ptr)
	setStatFromWASIStat(stat, &ret.stat)

	return 0
}

func setStatFromWASIStat(sstat *Stat_t, wstat *__wasi_filesystem_descriptor_stat) {
	// This will cause problems for people who want to compare inodes
	sstat.Dev = 0
	sstat.Ino = 0
	sstat.Rdev = 0

	sstat.Nlink = uint64(wstat.link_count)

	// No mode bits
	sstat.Mode = 0

	// No uid/gid
	sstat.Uid = 0
	sstat.Gid = 0
	sstat.Size = int64(wstat.filesize)

	// made up numbers
	sstat.Blksize = 512
	sstat.Blocks = (sstat.Size + 511) / int64(sstat.Blksize)

	setOptTime := func(t *Timespec, o *__wasi_option_datetime) {
		t.Sec = 0
		t.Nsec = 0
		if o.isSome {
			t.Sec = int32(o.t.seconds)
			t.Nsec = int64(o.t.nano)
		}
	}

	setOptTime(&sstat.Atim, &wstat.data_access_timestamp)
	setOptTime(&sstat.Mtim, &wstat.data_modification_timestamp)
	setOptTime(&sstat.Ctim, &wstat.status_change_timestamp)
}

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.stat-at
func __wasi_filesystem_types_method_descriptor_stat_at(d __wasi_filesystem_descriptor, flags __wasi_filesystem_path_flags, path __wasi_string, ret *__wasi_result_filesystem_descriptor_stat_error)

// int lstat(const char *path, struct stat * buf);
//
//go:export lstat
func lstat(pathname *byte, ptr unsafe.Pointer) int32 {
	pathLen := strlen(unsafe.Pointer(pathname))
	pathStr := _string{
		pathname, uintptr(pathLen),
	}
	path := *(*string)(unsafe.Pointer(&pathStr))
	dir, path := findPreopenForPath(path)

	pathWasiStr := __wasi_string{
		unsafe.Pointer(pathname), uint32(pathLen),
	}

	var ret __wasi_result_filesystem_descriptor_stat_error
	__wasi_filesystem_types_method_descriptor_stat_at(dir, 0, pathWasiStr, &ret)
	if ret.isErr {
		error_code := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 8))
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(error_code)))
		return -1

	}

	stat := (*Stat_t)(ptr)
	setStatFromWASIStat(stat, &ret.stat)

	return 0
}

func init() {
	populatePreopens()
}

var __wasi_cwd_descriptor __wasi_filesystem_descriptor

var __wasi_filesystem_preopens map[string]__wasi_filesystem_descriptor

func populatePreopens() {
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
		if path == "." {
			__wasi_cwd_descriptor = descriptor
		}
		ptr = unsafe.Add(ptr, unsafe.Sizeof(__wasi_tuple_descriptor_string{}))
	}
	__wasi_filesystem_preopens = dirs
}

//go:wasmimport wasi:filesystem/preopens@0.2.0-rc-2023-11-10 get-directories
func __wasi_filesystem_preopens_get_directories() __wasi_list_tuple

type __wasi_filesystem_error int16

type __wasi_tuple_descriptor_string struct {
	first  __wasi_filesystem_descriptor
	second __wasi_string
}

type __wasi_filesystem_path_flags uint32

const (
	__wasi_filesystem_path_flag_symlink_follow __wasi_filesystem_path_flags = 1 << iota
)

type __wasi_filesystem_descriptor_flags uint32

const (
	__wasi_filesystem_descriptor_flag_read __wasi_filesystem_descriptor_flags = 1 << iota
	__wasi_filesystem_descriptor_flag_write
	__wasi_filesystem_descriptor_flag_file_integrity_sync
	__wasi_filesystem_descriptor_flag_data_integrity_sync
	__wasi_filesystem_descriptor_flag_requested_write_sync
	__wasi_filesystem_descriptor_flag_mutate_directory
)

type __wasi_filesystem_open_at_flags uint32

const (
	__wasi_filesystem_open_at_flags_create __wasi_filesystem_open_at_flags = 1 << iota
	__wasi_filesystem_open_at_flags_directory
	__wasi_filesystem_open_at_flags_exclusive
	__wasi_filesystem_open_at_flags_truncate
)

func findPreopenForPath(path string) (__wasi_filesystem_descriptor, string) {
	for k, v := range __wasi_filesystem_preopens {
		if strings.HasPrefix(path, k) {
			path = strings.TrimPrefix(path, k+"/")
			return v, path
		}
	}
	return __wasi_cwd_descriptor, path
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

	dir, path := findPreopenForPath(path)

	var dflags __wasi_filesystem_descriptor_flags
	if (flags & O_RDONLY) == O_RDONLY {
		dflags |= __wasi_filesystem_descriptor_flag_read
	}
	if (flags & O_WRONLY) == O_WRONLY {
		dflags |= __wasi_filesystem_descriptor_flag_write
	}

	var oflags __wasi_filesystem_open_at_flags
	if flags&O_CREAT == O_CREAT {
		oflags |= __wasi_filesystem_open_at_flags_create
	}
	if flags&O_DIRECTORY == O_DIRECTORY {
		oflags |= __wasi_filesystem_open_at_flags_directory
	}
	if flags&O_EXCL == O_EXCL {
		oflags |= __wasi_filesystem_open_at_flags_exclusive
	}
	if flags&O_TRUNC == O_TRUNC {
		oflags |= __wasi_filesystem_open_at_flags_truncate
	}

	// By default, follow symlinks for open() unless O_NOFOLLOW was passed
	var pflags __wasi_filesystem_path_flags = __wasi_filesystem_path_flag_symlink_follow
	if flags&O_NOFOLLOW == 0 {
		pflags &^= __wasi_filesystem_path_flag_symlink_follow
	}

	pathWasiStr := __wasi_string{
		unsafe.Pointer(pathname), uint32(pathLen),
	}

	var ret __wasi_result_descriptor_error
	__wasi_filesystem_types_descriptor_open_at(dir, pflags, pathWasiStr.data, int32(pathWasiStr.len), oflags, dflags, &ret)
	if ret.isErr {
		libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(ret.val)))
		return -1
	}

	stream := wasiFile{
		d:     __wasi_filesystem_descriptor(ret.val),
		oflag: flags,
	}

	if flags&(O_WRONLY|O_APPEND) == (O_WRONLY | O_APPEND) {
		var ret __wasi_result_filesystem_descriptor_stat_error

		__wasi_filesystem_types_method_descriptor_stat(stream.d, &ret)

		if ret.isErr {
			error_code := *(*__wasi_filesystem_error)(unsafe.Add(unsafe.Pointer(&ret), 8))
			libcErrno = uintptr(__wasi_filesystem_err_to_errno(__wasi_filesystem_error(error_code)))
			return -1

		}

		stream.offset = int64(ret.stat.filesize)
	}

	libcfd := nextLibcFd
	nextLibcFd++

	wasiStreams[libcfd] = &stream

	return int32(libcfd)
}

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.read-via-stream
func __wasi_filesystem_types_method_descriptor_read_via_stream(d __wasi_filesystem_descriptor, offset int64, ret *__wasi_result_descriptor_error)

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.write-via-stream
func __wasi_filesystem_types_method_descriptor_write_via_stream(d __wasi_filesystem_descriptor, offset int64, ret *__wasi_result_descriptor_error)

//go:wasmimport wasi:filesystem/types@0.2.0-rc-2023-11-10 [method]descriptor.append-via-stream
func __wasi_filesystem_types_method_descriptor_append_via_stream(d __wasi_filesystem_descriptor, ret *__wasi_result_descriptor_error)

const (
	__wasi_filesystem_error_access                __wasi_filesystem_error = iota /// Permission denied, similar to `EACCES` in POSIX.
	__wasi_filesystem_error_would_block                                          /// Resource unavailable, or operation would block, similar to `EAGAIN` and `EWOULDBLOCK` in POSIX.
	__wasi_filesystem_error_already                                              /// Connection already in progress, similar to `EALREADY` in POSIX.
	__wasi_filesystem_error_bad_descriptor                                       /// Bad descriptor, similar to `EBADF` in POSIX.
	__wasi_filesystem_error_busy                                                 /// Device or resource busy, similar to `EBUSY` in POSIX.
	__wasi_filesystem_error_deadlock                                             /// Resource deadlock would occur, similar to `EDEADLK` in POSIX.
	__wasi_filesystem_error_quota                                                /// Storage quota exceeded, similar to `EDQUOT` in POSIX.
	__wasi_filesystem_error_exist                                                /// File exists, similar to `EEXIST` in POSIX.
	__wasi_filesystem_error_file_too_large                                       /// File too large, similar to `EFBIG` in POSIX.
	__wasi_filesystem_error_illegal_byte_sequence                                /// Illegal byte sequence, similar to `EILSEQ` in POSIX.
	__wasi_filesystem_error_in_progress                                          /// Operation in progress, similar to `EINPROGRESS` in POSIX.
	__wasi_filesystem_error_interrupted                                          /// Interrupted function, similar to `EINTR` in POSIX.
	__wasi_filesystem_error_invalid                                              /// Invalid argument, similar to `EINVAL` in POSIX.
	__wasi_filesystem_error_io                                                   /// I/O error, similar to `EIO` in POSIX.
	__wasi_filesystem_error_is_directory                                         /// Is a directory, similar to `EISDIR` in POSIX.
	__wasi_filesystem_error_loop                                                 /// Too many levels of symbolic links, similar to `ELOOP` in POSIX.
	__wasi_filesystem_error_too_many_links                                       /// Too many links, similar to `EMLINK` in POSIX.
	__wasi_filesystem_error_message_size                                         /// Message too large, similar to `EMSGSIZE` in POSIX.
	__wasi_filesystem_error_name_too_long                                        /// Filename too long, similar to `ENAMETOOLONG` in POSIX.
	__wasi_filesystem_error_no_device                                            /// No such device, similar to `ENODEV` in POSIX.
	__wasi_filesystem_error_no_entry                                             /// No such file or directory, similar to `ENOENT` in POSIX.
	__wasi_filesystem_error_no_lock                                              /// No locks available, similar to `ENOLCK` in POSIX.
	__wasi_filesystem_error_insufficient_memory                                  /// Not enough space, similar to `ENOMEM` in POSIX.
	__wasi_filesystem_error_insufficient_space                                   /// No space left on device, similar to `ENOSPC` in POSIX.
	__wasi_filesystem_error_not_directory                                        /// Not a directory or a symbolic link to a directory, similar to `ENOTDIR` in POSIX.
	__wasi_filesystem_error_not_empty                                            /// Directory not empty, similar to `ENOTEMPTY` in POSIX.
	__wasi_filesystem_error_not_recoverable                                      /// State not recoverable, similar to `ENOTRECOVERABLE` in POSIX.
	__wasi_filesystem_error_unsupported                                          /// Not supported, similar to `ENOTSUP` and `ENOSYS` in POSIX.
	__wasi_filesystem_error_no_tty                                               /// Inappropriate I/O control operation, similar to `ENOTTY` in POSIX.
	__wasi_filesystem_error_no_such_device                                       /// No such device or address, similar to `ENXIO` in POSIX.
	__wasi_filesystem_error_overflow                                             /// Value too large to be stored in data type, similar to `EOVERFLOW` in POSIX.
	__wasi_filesystem_error_not_permitted                                        /// Operation not permitted, similar to `EPERM` in POSIX.
	__wasi_filesystem_error_pipe                                                 /// Broken pipe, similar to `EPIPE` in POSIX.
	__wasi_filesystem_error_read_only                                            /// Read_only file system, similar to `EROFS` in POSIX.
	__wasi_filesystem_error_invalid_seek                                         /// Invalid seek, similar to `ESPIPE` in POSIX.
	__wasi_filesystem_error_text_file_busy                                       /// Text file busy, similar to `ETXTBSY` in POSIX.
	__wasi_filesystem_error_cross_device                                         /// Cross_device link, similar to `EXDEV` in POSIX.
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
func __wasi_filesystem_types_descriptor_open_at(dir __wasi_filesystem_descriptor, path_flags __wasi_filesystem_path_flags, path_ptr unsafe.Pointer, path_len int32, open_flags __wasi_filesystem_open_at_flags, flags __wasi_filesystem_descriptor_flags, ret *__wasi_result_descriptor_error)

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

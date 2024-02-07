//go:build wasip2

// mini libc wrapping wasi preview2 calls in a libc api

package syscall

import (
	"strings"
	"unsafe"

	"github.com/ydnar/wasm-tools-go/cm"
	"github.com/ydnar/wasm-tools-go/wasi/cli/stderr"
	"github.com/ydnar/wasm-tools-go/wasi/cli/stdin"
	"github.com/ydnar/wasm-tools-go/wasi/cli/stdout"
	"github.com/ydnar/wasm-tools-go/wasi/clocks/wallclock"
	"github.com/ydnar/wasm-tools-go/wasi/filesystem/preopens"
	"github.com/ydnar/wasm-tools-go/wasi/filesystem/types"
	"github.com/ydnar/wasm-tools-go/wasi/io/streams"
)

func goString(cstr *byte) string {
	return unsafe.String(cstr, strlen(cstr))
}

//go:export strlen
func strlen(cstr *byte) uintptr {
	if cstr == nil {
		return 0
	}
	ptr := unsafe.Pointer(cstr)
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

	stream, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	// TODO(ydnar): -1 not possible in uint32
	// if stream.d == -1 {
	// 	libcErrno = uintptr(EBADF)
	// 	return -1
	// }

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

	stream, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	// TODO(ydnar): -1 not possible in uint32
	// if stream.d == -1 {
	// 	libcErrno = uintptr(EBADF)
	// 	return -1
	// }

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

// At the moment, each time we have a file read or write we create a new stream.  Future implementations
// could change the current in or out file stream lazily.  We could do this by tracking input and output
// offsets individually, and if they don't match the current main offset, reopen the file stream at that location.

type wasiFile struct {
	d      types.Descriptor
	oflag  int32 // orignal open flags: O_RDONLY, O_WRONLY, O_RDWR
	offset int64 // current fd offset; updated with each read/write
}

// Need to figure out which system calls we're using:
//   stdin/stdout/stderr want streams, so we use stream read/write
//   but for regular files we can use the descriptor and explicitly write a buffer to the offset?
//   The mismatch comes from trying to combine these.

var wasiFiles map[int32]*wasiFile = make(map[int32]*wasiFile)
var nextLibcFd = int32(Stderr) + 1

var wasiErrno error

func readStdin(fd int32, buf *byte, count uint, offset int64) int {
	if fd != 0 {
		// TODO(dgryski): Not sure this will place nicely wit `dup()` and processes which close stdin
		panic("non-stdin passed to readStdin")
	}
	wasiStdin := stdin.GetStdin()
	if offset != 0 {
		libcErrno = uintptr(EINVAL)
		return -1
	}

	libcErrno = 0
	result := wasiStdin.BlockingRead(uint64(count))
	if err, isErr := result.Err(); isErr {
		if err.Closed() {
			libcErrno = 0
			return 0
		} else if err, ok := err.LastOperationFailed(); ok {
			wasiErrno = err
			libcErrno = uintptr(EWASIERROR)
		}
		return -1
	}

	list, _ := result.OK()
	dst := unsafe.Slice(buf, count)
	copy(dst, list.Slice())
	return int(list.Len())
}

func writeStdout(fd int32, buf *byte, count uint, offset int64) int {
	var stream streams.OutputStream
	switch fd {
	case 1:
		stream = stdout.GetStdout()
	case 2:
		stream = stderr.GetStderr()
	default:
		panic("non-stdout/err passed to writeStdout")
	}

	if offset != 0 {
		libcErrno = uintptr(EINVAL)
		return -1
	}

	src := unsafe.Slice(buf, count)
	var remaining = count

	// The blocking-write-and-flush call allows a maximum of 4096 bytes at a time.
	// We loop here by instead of doing subscribe/check-write/poll-one/write by hand.
	for remaining > 0 {
		len := uint(4096)
		if len > remaining {
			len = remaining
		}
		result := stream.BlockingWriteAndFlush(cm.ToList(src[:len]))
		if err, isErr := result.Err(); isErr {
			if err.Closed() {
				libcErrno = 0
				return 0
			} else if err, ok := err.LastOperationFailed(); ok {
				wasiErrno = err
				libcErrno = uintptr(EWASIERROR)
			}
			return -1
		}
		remaining -= len
	}

	return int(count)
}

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)

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

	streams, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	// TODO(ydnar): -1 not possible in uint32
	// if streams.d == -1 {
	// 	libcErrno = uintptr(EBADF)
	// 	return -1
	// }
	if streams.oflag&O_RDONLY == 0 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	result := streams.d.Read(types.FileSize(count), types.FileSize(offset))
	if code, isErr := result.Err(); isErr {
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	listAndEOF, _ := result.OK()
	list := listAndEOF.V0
	copy(unsafe.Slice(buf, count), list.Slice())

	// TODO(dgryski): EOF bool is ignored?
	return int(list.Len())
}

// ssize_t pwrite(int fd, void *buf, size_t count, off_t offset);
//
//go:export pwrite
func pwrite(fd int32, buf *byte, count uint, offset int64) int {
	// TODO(dgryski): Need to be consistent about all these checks; EBADF/EINVAL/... ?
	if -1 <= fd && fd <= Stderr {
		return writeStdout(fd, buf, count, offset)
	}

	streams, ok := wasiFiles[fd]
	if !ok {
		// TODO(dgryski): EINVAL?
		libcErrno = uintptr(EBADF)
		return -1
	}
	// TODO(ydnar): -1 not possible in uint32
	// if streams.d == -1 {
	// 	libcErrno = uintptr(EBADF)
	// 	return -1
	// }
	if streams.oflag&O_WRONLY == 0 {
		libcErrno = uintptr(EBADF)
		return -1
	}

	result := streams.d.Write(cm.NewList(buf, count), types.FileSize(offset))
	if code, isErr := result.Err(); isErr {
		// TODO(dgryski):
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	size, _ := result.OK()
	return int(size)
}

// ssize_t lseek(int fd, off_t offset, int whence);
//
//go:export lseek
func lseek(fd int32, offset int64, whence int) int64 {
	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}

	switch whence {
	case 0: // SEEK_SET
		stream.offset = offset
	case 1: // SEEK_CUR
		stream.offset += offset
	case 2: // SEEK_END
		result := stream.d.Stat()
		if code, isErr := result.Err(); isErr {
			libcErrno = uintptr(errorCodeToErrno(code))
			return -1
		}
		stat, _ := result.OK()
		stream.offset = int64(stat.Size) + offset
	}

	return int64(stream.offset)
}

// int close(int fd)
//
//go:export close
func close(fd int32) int32 {
	streams, ok := wasiFiles[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}

	// TODO(ydnar): -1 not possible in uint32
	// if streams.d != -1 {
	streams.d.ResourceDrop()
	// }

	delete(wasiFiles, fd)

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
func stat(pathname *byte, dst *Stat_t) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)

	result := dir.StatAt(0, relPath)
	if code, isErr := result.Err(); isErr {
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	wasiStat, _ := result.OK()
	setStatFromWASIStat(dst, &wasiStat)

	return 0
}

// int fstat(int fd, struct stat * buf);
//
//go:export fstat
func fstat(fd int32, dst *Stat_t) int32 {
	if -1 < fd && fd <= Stderr {
		// TODO(dgryski): fill in stat buffer for stdin etc
		return -1
	}

	stream, ok := wasiFiles[fd]
	if !ok {
		libcErrno = uintptr(EBADF)
		return -1
	}
	// TODO(ydnar): -1 not possible in uint32
	// if stream.d == -1 {
	// 	libcErrno = uintptr(EBADF)
	// 	return -1
	// }

	result := stream.d.Stat()
	if code, isErr := result.Err(); isErr {
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	wasiStat, _ := result.OK()
	setStatFromWASIStat(dst, &wasiStat)

	return 0
}

func setStatFromWASIStat(sstat *Stat_t, wstat *types.DescriptorStat) {
	// This will cause problems for people who want to compare inodes
	sstat.Dev = 0
	sstat.Ino = 0
	sstat.Rdev = 0

	sstat.Nlink = uint64(wstat.LinkCount)

	// No mode bits
	sstat.Mode = 0

	// No uid/gid
	sstat.Uid = 0
	sstat.Gid = 0
	sstat.Size = int64(wstat.Size)

	// made up numbers
	sstat.Blksize = 512
	sstat.Blocks = (sstat.Size + 511) / int64(sstat.Blksize)

	setOptTime := func(t *Timespec, o *cm.Option[wallclock.DateTime]) {
		t.Sec = 0
		t.Nsec = 0
		if some, ok := o.Some(); ok {
			t.Sec = int32(some.Seconds)
			t.Nsec = int64(some.Nanoseconds)
		}
	}

	setOptTime(&sstat.Atim, &wstat.DataAccessTimestamp)
	setOptTime(&sstat.Mtim, &wstat.DataModificationTimestamp)
	setOptTime(&sstat.Ctim, &wstat.StatusChangeTimestamp)
}

// int lstat(const char *path, struct stat * buf);
//
//go:export lstat
func lstat(pathname *byte, dst *Stat_t) int32 {
	path := goString(pathname)
	dir, relPath := findPreopenForPath(path)

	result := dir.StatAt(0, relPath)
	if code, isErr := result.Err(); isErr {
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	wasiStat, _ := result.OK()
	setStatFromWASIStat(dst, &wasiStat)

	return 0
}

func init() {
	populatePreopens()
}

var wasiCWD types.Descriptor

var wasiPreopens map[string]types.Descriptor

func populatePreopens() {
	dirs := preopens.GetDirectories().Slice()
	preopens := make(map[string]types.Descriptor, len(dirs))
	for _, tup := range dirs {
		desc, path := tup.V0, tup.V1
		preopens[path] = desc
		if path == "." {
			wasiCWD = desc
		}
	}
	wasiPreopens = preopens
}

// FIXME(ydnar): opening a stripped path fails, so ignore it.
func findPreopenForPath(path string) (types.Descriptor, string) {
	for k, v := range wasiPreopens {
		if strings.HasPrefix(path, k) {
			path = strings.TrimPrefix(path, k+"/")
			return v, path
		}
	}
	return wasiCWD, path
}

// int open(const char *pathname, int flags, mode_t mode);
//
//go:export open
func open(pathname *byte, flags int32, mode uint32) int32 {
	path := goString(pathname)
	dir, _ := findPreopenForPath(path)

	var dflags types.DescriptorFlags
	if (flags & O_RDONLY) == O_RDONLY {
		dflags |= types.DescriptorFlagsRead
	}
	if (flags & O_WRONLY) == O_WRONLY {
		dflags |= types.DescriptorFlagsWrite
	}

	var oflags types.OpenFlags
	if flags&O_CREAT == O_CREAT {
		oflags |= types.OpenFlagsCreate
	}
	if flags&O_DIRECTORY == O_DIRECTORY {
		oflags |= types.OpenFlagsDirectory
	}
	if flags&O_EXCL == O_EXCL {
		oflags |= types.OpenFlagsExclusive
	}
	if flags&O_TRUNC == O_TRUNC {
		oflags |= types.OpenFlagsTruncate
	}

	// By default, follow symlinks for open() unless O_NOFOLLOW was passed
	var pflags types.PathFlags = types.PathFlagsSymlinkFollow
	if flags&O_NOFOLLOW == O_NOFOLLOW {
		// O_NOFOLLOW was passed, so turn off SymlinkFollow
		pflags &^= types.PathFlagsSymlinkFollow
	}

	result := dir.OpenAt(pflags, path, oflags, dflags)
	if code, isErr := result.Err(); isErr {
		libcErrno = uintptr(errorCodeToErrno(code))
		return -1
	}

	desc, _ := result.OK()
	stream := wasiFile{
		d:     desc,
		oflag: flags,
	}

	if flags&(O_WRONLY|O_APPEND) == (O_WRONLY | O_APPEND) {
		result := stream.d.Stat()
		if code, isErr := result.Err(); isErr {
			libcErrno = uintptr(errorCodeToErrno(code))
			return -1
		}
		stat, _ := result.OK()
		stream.offset = int64(stat.Size)
	}

	libcfd := nextLibcFd
	nextLibcFd++

	wasiFiles[libcfd] = &stream

	return int32(libcfd)
}

func errorCodeToErrno(err types.ErrorCode) Errno {
	switch err {
	case types.ErrorCodeAccess:
		return EACCES
	case types.ErrorCodeWouldBlock:
		return EAGAIN
	case types.ErrorCodeAlready:
		return EALREADY
	case types.ErrorCodeBadDescriptor:
		return EBADF
	case types.ErrorCodeBusy:
		return EBUSY
	case types.ErrorCodeDeadlock:
		return EDEADLK
	case types.ErrorCodeQuota:
		return EDQUOT
	case types.ErrorCodeExist:
		return EEXIST
	case types.ErrorCodeFileTooLarge:
		return EFBIG
	case types.ErrorCodeIllegalByteSequence:
		return EILSEQ
	case types.ErrorCodeInProgress:
		return EINPROGRESS
	case types.ErrorCodeInterrupted:
		return EINTR
	case types.ErrorCodeInvalid:
		return EINVAL
	case types.ErrorCodeIo:
		return EIO
	case types.ErrorCodeIsDirectory:
		return EISDIR
	case types.ErrorCodeLoop:
		return ELOOP
	case types.ErrorCodeTooManyLinks:
		return EMLINK
	case types.ErrorCodeMessageSize:
		return EMSGSIZE
	case types.ErrorCodeNameTooLong:
		return ENAMETOOLONG
	case types.ErrorCodeNoDevice:
		return ENODEV
	case types.ErrorCodeNoEntry:
		return ENOENT
	case types.ErrorCodeNoLock:
		return ENOLCK
	case types.ErrorCodeInsufficientMemory:
		return ENOMEM
	case types.ErrorCodeInsufficientSpace:
		return ENOSPC
	case types.ErrorCodeNotDirectory:
		return ENOTDIR
	case types.ErrorCodeNotEmpty:
		return ENOTEMPTY
	case types.ErrorCodeNotRecoverable:
		return ENOTRECOVERABLE
	case types.ErrorCodeUnsupported:
		return ENOSYS
	case types.ErrorCodeNoTTY:
		return ENOTTY
	case types.ErrorCodeNoSuchDevice:
		return ENXIO
	case types.ErrorCodeOverflow:
		return EOVERFLOW
	case types.ErrorCodeNotPermitted:
		return EPERM
	case types.ErrorCodePipe:
		return EPIPE
	case types.ErrorCodeReadOnly:
		return EROFS
	case types.ErrorCodeInvalidSeek:
		return ESPIPE
	case types.ErrorCodeTextFileBusy:
		return ETXTBSY
	case types.ErrorCodeCrossDevice:
		return EXDEV
	}
	return Errno(err)
}

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

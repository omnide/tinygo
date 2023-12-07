//go:build wasip2

// mini libc wrapping wasi preview2 calls in a libc api

package syscall

import "unsafe"

//go:export strlen
func strlen(ptr unsafe.Pointer) uintptr {
	return 0
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

//go:export environ
var environ *unsafe.Pointer

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

// int open(const char *pathname, int flags, mode_t mode);
//
//go:export open
func open(pathname *byte, flags int32, mode uint32) int32 {
	return 0
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

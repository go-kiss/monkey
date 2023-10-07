package monkey

import (
	"syscall"
	"unsafe"
)

func rawMemoryAccess(p uintptr, length int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), length)
}

func pageStart(ptr uintptr) uintptr {
	return ptr & ^(uintptr(syscall.Getpagesize() - 1))
}

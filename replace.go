package monkey

import (
	"reflect"
	"syscall"
	"unsafe"
)

func rawMemoryAccess(p uintptr, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: p,
		Len:  length,
		Cap:  length,
	}))
}

func pageStart(ptr uintptr) uintptr {
	return ptr & ^(uintptr(syscall.Getpagesize() - 1))
}

func allowExec(p uintptr, length int) {
	mprotectCrossPage(p, length, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)
}

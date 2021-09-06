package monkey

import (
	"fmt"
	"reflect"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/arch/x86/x86asm"
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

func setX(p uintptr, l int) {
	mprotectCrossPage(p, l, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)
}

func alginPatch(from uintptr) (original []byte) {
	f := rawMemoryAccess(from, 32)

	s := 0
	for {
		i, err := x86asm.Decode(f[s:], 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		original = append(original, f[s:s+i.Len]...)
		s += i.Len
		if s >= 13 {
			return
		}
	}
}

func findPadding(to uintptr) (begin, end uintptr, sp int64) {
	f := rawMemoryAccess(to, 256)

	s := 0
	findMagic := false
	mn := 0
	for {
		i, err := x86asm.Decode(f[s:], 64)
		if err != nil {
			return
		}

		switch i.Op {
		case x86asm.MOV:
			if i.Args[1].String() == "0x12faac9" {
				if !findMagic {
					findMagic = true
					begin = uintptr(s)
				}
				if findMagic {
					mn++
				}
			} else {
				findMagic = false
				mn = 0
				begin = 0
				end = 0
			}
		case x86asm.RET:
			return
		case x86asm.SUB:
			if i.Args[0].String() == "RSP" {
				sp, _ = strconv.ParseInt(i.Args[1].String(), 0, 64)
			}
		}

		s += i.Len
		if mn == 8 {
			end = uintptr(s)
			return
		}
	}
}

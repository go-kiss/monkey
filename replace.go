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

// from is a pointer to the actual function
// to is a pointer to a go funcvalue
func replaceFunction(from, to reflect.Value) (original []byte) {
	_from := from.Pointer()
	// __from := (uintptr)(getPtr(from))

	_to := to.Pointer()
	__to := (uintptr)(getPtr(to))

	t := alginPatch(_from)
	if begin, end, sp := findPadding(_to); begin > 0 {
		jumpToReal := jmp(_to + end)
		for len(jumpToReal) < 2*9 {
			jumpToReal = append(jumpToReal, 0x90)
		}
		copyToLocation(_to+begin, jumpToReal)

		old := jmp(_from + uintptr(len(t)))
		ss := []byte{
			0x48, 0x83, 0xc4, // add rsp,0x??
			byte(sp),
		}
		t = append(ss, t...)
		jumpToOld := append(t, old...)
		for len(jumpToOld) < 6*9 {
			jumpToOld = append(jumpToOld, 0x90)
		}
		copyToLocation(_to+end-6*9, jumpToOld)
		// dump("", rawMemoryAccess(_to, 128))
	}

	jumpData := jmpToFunctionValue(__to)
	f := rawMemoryAccess(_from, len(jumpData))
	original = make([]byte, len(f))
	copy(original, f)

	copyToLocation(_from, jumpData)
	return
}

func dump(msg string, f []byte) {
	fmt.Println("-------", msg)
	s := 0
	for {
		i, err := x86asm.Decode(f[s:], 64)
		if err != nil {
			return
		}
		s += i.Len
		fmt.Println(i)
		if s >= len(f) {
			return
		}
	}
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

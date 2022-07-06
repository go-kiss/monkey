package monkey

import (
	"golang.org/x/arch/x86/x86asm"
)

// Assembles a jump to a function value
func jmpToFunctionValue(to uintptr) []byte {
	return []byte{
		0x49, 0xBD,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56),   // movabs r13,to
		0x41, 0xFF, 0xE5, // jmp r13
	}
}

func littleEndian(to uintptr) []byte {
	return []byte{
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56),
	}
}

// Assembles a jump to a function value
func jmpToGoFn(to uintptr) []byte {
	return []byte{
		0x48, 0xBA,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56), // movabs rdx,to
		0xFF, 0x22,     // jmp QWORD PTR [rdx]
	}
}

func jmpTable(g, to uintptr, gofn bool) []byte {
	b := []byte{
		// movq r13, g
		0x49, 0xBD,
		byte(g),
		byte(g >> 8),
		byte(g >> 16),
		byte(g >> 24),
		byte(g >> 32),
		byte(g >> 40),
		byte(g >> 48),
		byte(g >> 56),
		// cmp r12, r13
		0x4D, 0x39, 0xEC,
		// jne $+(2+12)
		0x75, 0x0c,
	}
	if gofn {
		b = append(b, jmpToGoFn(to)...)
	} else {
		b = append(b, jmpToFunctionValue(to)...)
	}
	return b
}

func alginPatch(from uintptr) (original []byte) {
	f := rawMemoryAccess(from, 32)

	s := 0
	for {
		i, err := x86asm.Decode(f[s:], 64)
		if err != nil {
			panic(err)
		}
		original = append(original, f[s:s+i.Len]...)
		s += i.Len
		if s >= 13 {
			return
		}
	}
}

func getFirstCallFunc(from uintptr) uintptr {
	f := rawMemoryAccess(from, 1024)

	s := 0
	for {
		i, err := x86asm.Decode(f[s:], 64)
		if err != nil {
			panic(err)
		}
		if i.Op == x86asm.CALL {
			arg := i.Args[0]
			imm := arg.(x86asm.Rel)
			next := from + uintptr(s+i.Len)
			var to uintptr
			if imm > 0 {
				to = next + uintptr(imm)
			} else {
				to = next - uintptr(-imm)
			}
			return to
		}
		s += i.Len
		if s >= 1024 {
			panic("Can not find CALL instruction")
		}
	}
}

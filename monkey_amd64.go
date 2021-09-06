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

func getg() []byte {
	return []byte{
		// movq r12, gs:0x30
		0x65, 0x4C, 0x8B, 0x24, 0x25, 0x30, 0x00, 0x00, 0x00,
	}
}

func jmpTable(g, to uintptr) []byte {
	return []byte{
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
		// jne $+13
		0x75, 0x0d,
		// movq r13, to
		0x49, 0xBD,
		byte(to),
		byte(to >> 8),
		byte(to >> 16),
		byte(to >> 24),
		byte(to >> 32),
		byte(to >> 40),
		byte(to >> 48),
		byte(to >> 56),
		// jmp r13
		0x41, 0xFF, 0xE5,
	}
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

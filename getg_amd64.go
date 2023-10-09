package monkey

// getg 获取当前协程的指针，根据go内部abi的文档在amd64架构中使用r14寄存器保存当前的协程g地址
// 参见 https://go.googlesource.com/go/+/refs/heads/dev.regabi/src/cmd/compile/internal-abi.md
func getg() []byte {
	return []byte{
		// mov r12, r14
		0x4D, 0x89, 0xF4,
	}
}

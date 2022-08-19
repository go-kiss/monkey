# Go语言猴子补丁框架 🙉 🐒

![test workflow](https://github.com/go-kiss/monkey/actions/workflows/go.yml/badge.svg)

Go 语言猴子补丁（monkey patching）框架。

本项目对 [Bouke](https://bou.ke/blog/monkey-patching-in-go/) 的项目做了优化，不同协程可以独立 patch 同一个函数而互不影响。从而可以并发运行单元测试。

工作原理请参考我的系列文章：

- [Go语言实现猴子补丁](https://taoshu.in/go/monkey.html)
- [Go语言实现猴子补丁【二】](https://taoshu.in/go/monkey-2.html)
- [Go语言实现猴子补丁【三】](https://taoshu.in/go/monkey-3.html)

Bouke 已经不再维护原项目，所以只能开一个新项目了🤣。

有兴趣的同学也可以加微信 `taoshu-in` 讨论，拉你进群。

## 快速入门

首先，引入 monkey 包

```bash
go get github.com/go-kiss/monkey
```

然后，调用 `monkey.Patch` 方法 mock 指定函数。

```go
package main

import (
	"fmt"

	"github.com/go-kiss/monkey"
)

func sum(a, b int) int { return a + b }

func sub1[T int|float64](a, b T) T { return a - b }
func sub2[T int|float64](a, b T) T { return a + b }

type S struct { i int }

func (s *S) Get() int { return i }

func main() {
	// mock 普通函数
	monkey.Patch(sum, func(a b int) int { return a - b })
	fmt.Println(sum(1, 2)) // 输出 -1
	// mock 泛型函数
	monkey.Patch(sub1[int], sub2[int])
	fmt.Println(sub1(1, 2)) // 输出 3
	// mock 结构体方法
	monkey.Patch((*s).Get, func(s *S) int { return -1 })
	var s S
	fmt.Println(s.Get()) // 输出 -1
}
```

更多用法请参考[使用示例](./examples)和[测试用例](./monkey_test.go)。

## 注意事项

1. Monkey 需要关闭 Go 语言的内联优化才能生效，比如测试的时候需要：`go test -gcflags=all=-l`。
2. Monkey 需要在运行的时候修改内存代码段，因而无法在一些对安全性要求比较高的系统上工作。
3. Monkey 不应该用于生产系统，但用来 mock 测试代码还是没有问题的。
4. Monkey 目前仅支持 amd64 指令架构，支持 linux/macos/windows 平台。

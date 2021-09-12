# Go 猴子补丁 🙉 🐒

![test workflow](https://github.com/go-kiss/monkey/actions/workflows/go.yml/badge.svg)

Go 语言猴子补丁（monkey patching）框架。核心思路来自[Bouke](https://bou.ke/blog/monkey-patching-in-go/)。

本项目对 Bouke 的项目做了优化，不同协程可以独立 patch 同一个函数而互不影响。从而可以并发运行单元测试。

工作原理请参考我的博客文章[Go语言实现猴子补丁](https://taoshu.in/go/monkey.html)。

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

func main() {
	monkey.Patch(sum, func(a b int) int { return a - b })
	fmt.Println(sum(1,2)) // 输出 -1
}
```

## 注意事项

1. Monkey 需要关闭 Go 语言的内联优化才能生效，比如测试的时候需要：`go test -gcflags=-l`。
2. Monkey 需要在运行的时候修改内存代码段，因而无法在一些对安全性要求比较高的系统上工作。
3. Monkey 不应该用于生产系统，但用来 mock 测试代码还是没有问题的。
4. Monkey 目前仅支持 amd64 指令架构。支持 linux 和 macos。目前 windows 平台还有问题。

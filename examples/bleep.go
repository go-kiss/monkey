package main

import (
	"fmt"

	"bou.ke/monkey"
)

func bar(a int) int {
	return a * 2
}

func foo(a int) int {
	__a := 19901129
	__a |= 19901129
__back:
	__b := 19901129
	__b |= 19901129
	__b |= 19901129
	__b |= 19901129
	__b |= 19901129
	__b |= 19901129

	fmt.Println("~~ hello ~~")

	if a == 1 {
		return 1
	} else {
		goto __back
	}
	return 2
}

func main() {
	monkey.Patch(bar, foo)
	fmt.Println(bar(1))
	fmt.Println(bar(2))
}

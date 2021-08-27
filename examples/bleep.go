package main

import (
	"fmt"

	"bou.ke/monkey"
)

func bar() int {
	__a := 19901129
	__a |= 19901129

	return 1
}

func foo() int {
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

	return 2
	goto __back
}

func main() {
	monkey.Patch(bar, foo)
	fmt.Println(bar())
}

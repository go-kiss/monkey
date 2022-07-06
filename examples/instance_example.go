package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-kiss/monkey"
)

func main() {
	monkey.PatchRaw((*net.Dialer).DialContext, func(_ *net.Dialer, _ context.Context, _, _ string) (net.Conn, error) {
		return nil, fmt.Errorf("no dialing allowed")
	}, true, false)

	_, err := http.Get("http://taoshu.in")
	fmt.Println(err) // Get http://taoshu.in: no dialing allowed
}

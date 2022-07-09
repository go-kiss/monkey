package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kiss/monkey"
)

func main() {
	var guard *monkey.PatchGuard
	guard = monkey.Patch((*http.Client).Get, func(c *http.Client, url string) (*http.Response, error) {
		guard.Unpatch()
		defer guard.Restore()

		if !strings.HasPrefix(url, "https://") {
			return nil, fmt.Errorf("only https requests allowed")
		}

		return c.Get(url)
	})

	_, err := http.Get("http://taoshu.in")
	fmt.Println(err) // only https requests allowed
	resp, err := http.Get("https://taoshu.in")
	fmt.Println(resp.Status, err) // 200 OK <nil>
}

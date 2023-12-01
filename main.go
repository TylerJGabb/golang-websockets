package main

import (
	"fmt"
	"net/http"

	socks "tg.sandbox/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}

func main() {
	socks.Start()
}

package main

import (
	"fmt"
	"math/rand"
	"time"

	"tg.sandbox/demos"
)

func main() {
	// socks.Start()
	x := demos.NewForWireCloseable()
	x.Send("hello", "world", "!")
	x.Send("A", "B", "C")
	x.Send("1")
	x.Send("2")
	x.Send("3")
	base := 500 * (rand.Intn(5) + 5)
	time.Sleep(time.Millisecond * time.Duration(base))
	x.Close()
	var input string
	fmt.Scan(&input)
}

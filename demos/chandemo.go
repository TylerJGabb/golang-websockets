package demos

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func f(n int, name string) {
	for i := 0; i < 10; i++ {
		fmt.Printf("%v: %d\n", name, i)
		amt := time.Duration(rand.Intn(250))
		time.Sleep(time.Millisecond * amt)
	}
}

func demo1() {
	arr := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	for i := 0; i < 10; i++ {
		go f(10, "goroutine"+arr[i])
	}
	var input string
	fmt.Scanln(&input)
}

// read only chan
func printer(c <-chan string) {
	for {
		msg := <-c
		fmt.Println("printer: " + msg)
		time.Sleep(time.Second * 1)
	}
}

// write only chan
func pinger(c chan<- string) {
	for i := 0; ; i++ {
		c <- "ping" + strconv.Itoa(i)
		fmt.Printf("ping %d\n", i)
	}
}

func ponger(c chan<- string) {
	for i := 0; ; i++ {
		c <- "pong" + strconv.Itoa(i)
		fmt.Printf("pong %d\n", i)
	}
}

func demo2() {
	var c chan string = make(chan string)
	go pinger(c)
	go ponger(c)
	go printer(c)
	var input string
	fmt.Scanln(&input)
}

func demo3() {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		for {
			c1 <- "from 1"
			time.Sleep(time.Second * 2)
		}
	}()

	go func() {
		for {
			c2 <- "from 2"
			time.Sleep(time.Second * 3)
		}
	}()

	go func() {
		for {
			select {
			case msg1 := <-c1:
				fmt.Println(msg1)
			case msg2 := <-c2:
				fmt.Println(msg2)
			case timeout := <-time.After(500 * time.Millisecond):
				fmt.Println("timeout" + timeout.String())
			}
		}
	}()

	var input string
	fmt.Scanln(&input)
}

type ForWireCloseable struct {
	outChan chan []string
	ctx     context.Context
	cancel  context.CancelFunc
}

func (fwc *ForWireCloseable) Send(str ...string) {
	fmt.Printf("CHAN_WRITE_0 str=%s\n", str)
	fwc.outChan <- str
	fmt.Printf("CHAN_WRITE_1 str=%s\n", str)
}

func (fwc *ForWireCloseable) Close() {
	fwc.cancel()
}

func NewForWireCloseable() *ForWireCloseable {
	outChan := make(chan []string, 10)
	ctx, cancel := context.WithCancel(context.TODO())
	fwcP := &ForWireCloseable{
		outChan: outChan,
		ctx:     ctx,
		cancel:  cancel,
	}
	go func() {
	Loop:
		for {
			select {
			case msg := <-outChan:
				dur := time.Millisecond * time.Duration(500+rand.Intn(500))
				fmt.Printf("CHAN_READ msg=%v, dur=%v\n", msg, dur)
				time.Sleep(dur)

			case <-ctx.Done():
				fmt.Printf("closeChan\n")
				break Loop

			case <-time.After(time.Second * 2):
				fmt.Printf("chanloop tick %p\n", fwcP)

			}
		}
		fmt.Println("Terminating chan loop")
	}()
	fmt.Printf("%+v\n", fwcP)
	return fwcP
}

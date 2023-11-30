package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	//	"html"

	"net/http"
	// yaml "gopkg.in/yaml.v3"
)

var (
	WS_KEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

func mirrorHeaders(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v %s %v\n", r.Method, r.URL, r.Proto)
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
	w.WriteHeader(http.StatusSwitchingProtocols)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}

type MyHandler struct{}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("WEBSOCKET INITIALIZING -- %v %s %v\n", r.Method, r.URL, r.Proto)
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	concat := key + WS_KEY
	fmt.Printf("Concat: %v -- %v\n", concat, len(concat))
	sha1 := sha1.New()
	sha1.Write([]byte(concat))
	sha1Sum := sha1.Sum(nil)
	fmt.Printf("SHA1: %v -- %v\n", string(sha1Sum), len(sha1Sum))
	secWebSockerAccept := base64.StdEncoding.EncodeToString(sha1Sum)
	fmt.Printf("Sec-WebSocket-Accept: %v\n", secWebSockerAccept)
	w.Header().Add("Sec-WebSocket-Accept", secWebSockerAccept)
	w.WriteHeader(http.StatusSwitchingProtocols)
	go func() {
		for {
			fmt.Printf("Waiting for data...\n")
			time.Sleep(5 * time.Second)
			fmt.Print("Reading data...\n")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				fmt.Printf("Error reading body: %v\n", err)
				continue
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			fmt.Printf("Body: %v\n", string(body))
		}
	}()
}

func SockClientSendMsg() {
	// init
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("Error resolving TCP address: %v\n", err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("Error dialing: %v\n", err)
	}
	// send message
	_, err = conn.Write([]byte(time.Now().UTC().String()))
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
	}
	// receive message
	var buf [1024]byte
	_, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Printf("Error reading: %v\n", err)
	}
}

// https://yalantis.com/blog/how-to-build-websockets-in-go/
func StartSockServer() {
	fmt.Printf("Resolving TCP address...\n")
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("Error resolving TCP address: %v\n", err)
		return
	}
	fmt.Printf("Listening...\n")
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		return
	}
	for {
		fmt.Printf("Accepting...\n")
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Error accepting: %v\n", err)
			return
		}
		fmt.Printf("Writing...\n")
		if _, err := conn.Write([]byte("hello")); err != nil {
			fmt.Printf("Error writing: %v\n", err)
			return
		}
		buf := make([]byte, 1024)
		fmt.Printf("Waiting for data...\n")
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Printf("Error reading: %v\n", err)
			return
		}
		fmt.Printf("Read %v bytes: %v\n", n, string(buf))
	}
}

func main() {
	// port, portSet := os.LookupEnv("PORT")
	// if !portSet {
	// 	port = "8080"
	// }

	// http.HandleFunc("/mirrorHeaders", mirrorHeaders)
	// http.HandleFunc("/ws", wsHandler)
	// http.Handle("/", &MyHandler{})
	// fmt.Println("Server is running on port", port)
	// go func() {
	// 	if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 		panic(err)
	// 	}
	// }()
	go func() {
		StartSockServer()
	}()
	for {
		time.Sleep(5 * time.Second)
		SockClientSendMsg()
	}
	// sleep forever
	select {}

}

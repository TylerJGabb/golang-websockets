package websocket

import (
	"fmt"
	"net/http"
)

// wait for a new message to arrive, then send it to the channel?
// need to learn about channels

func loopWsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling websocket request")
	fmt.Printf("%v %s %v\n", r.Method, r.URL, r.Proto)
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}
	hwsp, err := NewHttpHijackingWebSocket(w, r)
	if err != nil {
		fmt.Println("Error creating new websocket")
		return
	}
	var ws WebSocket = hwsp
	err = ws.Handshake()
	if err != nil {
		fmt.Println("Error handshaking")
		return
	}
	// read indefinitely until the connection is closed by the reader
	// send read frames to the channel
	// that seems like the best way to do this, reading frame by frame.

	frame, err := ws.ReadFrame()
	if err != nil {
		fmt.Printf("Error reading frame: %v\n", err)
		ws.ForceClose()
		return
	}
	preview := frame.Payload[0:min(30, frame.Header.Length)]
	fmt.Printf("Frame: %s\n", preview)
	responsePayload := "Echo: " + string(preview)
	echoFrame := WebSocketFrame{
		Header: FrameHeader{
			Fin:    true,
			Opcode: OPCODE_TEXT,
			Length: uint64(len(responsePayload)),
		},
		Payload: []byte(responsePayload),
	}
	ws.WriteFrame(echoFrame)
}

func Start() {
	fmt.Println("Starting websocket server")
	http.HandleFunc("/ws", loopWsHandler)
	go func() {
		err := http.ListenAndServe("localhost:8080", nil)
		if err != nil {
			fmt.Println("Error starting websocket server")
		}
	}()
	fmt.Println("Websocket server started, listening on localhost:8080")
	select {}
}

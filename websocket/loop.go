package websocket

import (
	"fmt"
	"net/http"
)

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
	for {
		frame, err := ws.ReadFrame()
		if err != nil {
			fmt.Printf("Error reading frame: %v\n", err)
			ws.ForceClose()
			return
		}
		switch frame.Header.Opcode {
		case OPCODE_CLOSE:
			ws.SendCloseFrame(STATUS_CODE_NORMAL_CLOSURE)
			ws.ForceClose()
			return
		case OPCODE_PING:
			fmt.Println("Received ping")
			ws.WriteFrame(PONG_FRAME)
			return
		case OPCODE_PONG:
			fmt.Println("Received pong")
			return
		default:
			fmt.Printf("Received opcode: %v\n", frame.Header.Opcode)
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
}

func Start() {
	fmt.Println("Starting websocket server")
	http.HandleFunc("/ws", loopWsHandler)
	go func() {
		err := http.ListenAndServe("localhost:8080", nil)
		if err != nil {
			panic(err)
		}
	}()
	fmt.Println("Websocket server started, listening on localhost:8080")
	select {}
}

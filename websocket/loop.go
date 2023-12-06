package websocket

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	log "tg.sandbox/logging"
)

func loopWsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := log.NewContextWithLogger(uuid.NewString())
	logger := log.SugaredLoggerFromContext(ctx)
	logger.Infow("New websocket connection",
		"method", r.Method,
		"url", r.URL,
		"proto", r.Proto,
		"headers", r.Header,
	)
	hwsp, err := NewHttpHijackingWebSocket(w, r)
	if err != nil {
		logger.Error("Error creating websocket", "error", err)
		return
	}
	var ws WebSocket = hwsp
	err = ws.Handshake()
	if err != nil {
		logger.Error("Error handshaking", "error", err)
		return
	}
	for {
		frame, err := ws.ReadFrame()
		if err != nil {
			logger.Error("Error reading frame", "error", err)
			ws.ForceClose()
			return
		}
		switch frame.Header.Opcode {
		case OPCODE_CLOSE:
			ws.SendCloseFrame(STATUS_CODE_NORMAL_CLOSURE)
			ws.ForceClose()
			return
		case OPCODE_PING:
			logger.Debug("Received ping")
			ws.WriteFrame(PONG_FRAME)
			return
		case OPCODE_PONG:
			logger.Debug("Received pong")
			return
		default:
			logger.Info("Received frame", "opcode", frame.Header.Opcode)
		}

		preview := frame.Payload[0:min(30, frame.Header.Length)]
		logger.Info("Frame Preview", "frame", frame)
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

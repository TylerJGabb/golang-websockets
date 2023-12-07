package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	messageBroker *Broker = NewBroker()
)

func loopWsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling websocket request")
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
		message := &Message{}
		err = json.Unmarshal(frame.Payload, message)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			ws.SendText(fmt.Sprintf("Error unmarshalling message: %v", err))
		}
		switch message.Operation {
		case "subscribe":
			messageBroker.AddSubscriber(message.Topic, &ws)
			ws.SendText(fmt.Sprintf("Subscribed to topic: %s", message.Topic))
		case "unsubscribe":
			err := messageBroker.RemoveSubscriber(message.Topic, &ws)
			if err != nil {
				ws.SendText(fmt.Sprintf("Error unsubscribing from topic %s: %v", message.Topic, err))
			} else {

				ws.SendText(fmt.Sprintf("Unsubscribed from topic: %s", message.Topic))
			}
		case "publish":
			err := messageBroker.SendMessageToTopic(message.Topic, message.Message)
			if err != nil {
				ws.SendText(fmt.Sprintf("Error publishing to topic %s: %v", message.Topic, err))
			} else {
				ws.SendText(fmt.Sprintf("Published to topic: %s", message.Topic))
			}
		default:
			fmt.Printf("Unknown operation: %s\n", message.Operation)
		}
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

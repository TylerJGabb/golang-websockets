package websocket

import "fmt"

type Broker struct {
	topicSubs map[string][]*WebSocket
}

type Message struct {
	// operation: subscribe, unsubscribe, publish
	Operation string `json:"operation"`
	// topic name
	Topic string `json:"topic"`
	// message
	Message string `json:"message"`
}

func NewBroker() *Broker {
	return &Broker{
		topicSubs: make(map[string][]*WebSocket),
	}
}

func (b *Broker) AddSubscriber(topicName string, ws *WebSocket) {
	subs, ok := b.topicSubs[topicName]
	if !ok {
		fmt.Printf("Topic %s does not have any subscribers, creating the list\n", topicName)
		subs = make([]*WebSocket, 0)
	}
	subs = append(subs, ws)
	b.topicSubs[topicName] = subs
}

func (b *Broker) RemoveSubscriber(topicName string, ws *WebSocket) error {
	subs, ok := b.topicSubs[topicName]
	if !ok {
		return fmt.Errorf("Topic %s does not exist\n", topicName)
	} else {
		for i, wsP := range subs {
			if ws == wsP {
				subs = append(subs[:i], subs[i+1:]...)
				b.topicSubs[topicName] = subs
				return nil
			}
		}
		return fmt.Errorf("Websocket %p not found in topic %s\n", ws, topicName)
	}

}

func (b *Broker) SendMessageToTopic(topicName string, message string) error {
	subs, ok := b.topicSubs[topicName]
	if ok {
		for _, wsP := range subs {
			ws := *wsP
			fmt.Printf("Sending message to topic=%s for ws=%p\n", topicName, wsP)
			err := ws.SendText(message)
			if err != nil {
				fmt.Printf("Error sending message to topic %s: %v\n", topicName, err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("Topic %s does not exist\n", topicName)
	}

}

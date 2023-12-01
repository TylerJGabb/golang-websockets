package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

const (
	OPCODE_CONTINUATION = 0x0
	OPCODE_TEXT         = 0x1
	OPCODE_BINARY       = 0x2
	OPCODE_CLOSE        = 0x8
	OPCODE_PING         = 0x9
	OPCODE_PONG         = 0xA
)

const (
	FRAME_FIN_NO_MORE_FRAMES = 0x0
	FRAME_FIN_MORE_FRAMES    = 0x1
)

var (
	readCounter int = 0
)

type WS struct {
	Connection net.Conn
	ReadWriter *bufio.ReadWriter
	Header     http.Header
}

type Frame struct {
	Fin     bool   `json:"fin"`
	Rsv1    bool   `json:"rsv1"`
	Rsv2    bool   `json:"rsv2"`
	Rsv3    bool   `json:"rsv3"`
	Opcode  byte   `json:"opcode"`
	Masked  bool   `json:"masked"`
	Length  uint64 `json:"length"`
	Mask    uint32 `json:"mask"`
	Payload string `json:"payload"`
}

func (ws *WS) Read() (Frame, error) {
	readCounter++
	fmt.Printf("DEBUG: Inside Read %d\n", readCounter)
	header := make([]byte, 2)
	fmt.Printf("DEBUG: About to read Header\n")
	io.ReadFull(ws.ReadWriter, header)
	o1 := header[0]
	o2 := header[1]

	fin := o1&0x80 != 0
	rsv1 := o1&0x40 != 0
	rsv2 := o1&0x20 != 0
	rsv3 := o1&0x10 != 0
	isMasked := o2&0x80 != 0
	opcode := o2 & 0x7f
	fmt.Printf("FIN: %v\n", fin)
	fmt.Printf("RSV1: %v\n", rsv1)
	fmt.Printf("RSV2: %v\n", rsv2)
	fmt.Printf("RSV3: %v\n", rsv3)
	fmt.Printf("IS_MASKED: %v\n", isMasked)
	fmt.Printf("OPCODE: %x\n", opcode)

	length := uint64(o2 & 0x7f)
	fmt.Printf("LENGTH: %v\n", length)
	if length == 126 {
		// length is held in the next 2 bytes
		lengthBytes := make([]byte, 2)
		io.ReadFull(ws.ReadWriter, lengthBytes)
		fmt.Print("LENGTH: 0X")
		for _, b := range lengthBytes {
			fmt.Printf("%02X", b)
		}
		fmt.Println()
		length = uint64(binary.BigEndian.Uint16(lengthBytes))
	} else if length == 127 {
		// length is held in the next 8 bytes
		lengthBytes := make([]byte, 8)
		io.ReadFull(ws.ReadWriter, lengthBytes)
		fmt.Print("LENGTH: 0X")
		for _, b := range lengthBytes {
			fmt.Printf("%02X", b)
		}
		fmt.Println()
		length = binary.BigEndian.Uint64(lengthBytes)
	}
	fmt.Printf("LENGTH: %v\n", length)

	maskBytes := make([]byte, 4)
	if isMasked {
		io.ReadFull(ws.ReadWriter, maskBytes)
		fmt.Print("MASK: ")
		for _, b := range maskBytes {
			fmt.Printf("%08b ", b)
		}
		fmt.Println()
	}
	mask := binary.BigEndian.Uint32(maskBytes)
	payload := make([]byte, length)
	io.ReadFull(ws.ReadWriter, payload)
	for i, octet := range payload {
		j := i % 4
		// it would be faster to bit shift, but this is easier to read
		transform := maskBytes[j]
		after := octet ^ transform
		fmt.Printf("TRACE: octet: %08b, transform: %08b, after: %08b\n", octet, transform, after)
		payload[i] = after
	}

	return Frame{
		Fin:     fin,
		Rsv1:    rsv1,
		Rsv2:    rsv2,
		Rsv3:    rsv3,
		Opcode:  opcode,
		Masked:  isMasked,
		Mask:    mask,
		Length:  length,
		Payload: string(payload),
	}, nil
}

func (ws *WS) Send(fr Frame) error {
	return nil
}

func (ws *WS) Handshake() error {
	key := ws.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return fmt.Errorf("No Sec-WebSocket-Key header found")
	}
	sha := sha1.New()
	sha.Write([]byte(key))
	sha.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	hash := base64.StdEncoding.EncodeToString(sha.Sum(nil))
	var header http.Header = make(http.Header)
	header.Add("Upgrade", "websocket")
	header.Add("Connection", "Upgrade")
	header.Add("Sec-WebSocket-Accept", hash)
	ws.ReadWriter.Write([]byte("HTTP/1.1 101 Switching Protocols\n"))
	header.Write(ws.ReadWriter)
	ws.ReadWriter.Write([]byte("\n"))
	ws.ReadWriter.Flush()
	return nil
}

func NewWS(w http.ResponseWriter, r *http.Request) (*WS, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("Unable to typeassert http.ResponseWriter to http.Hijacker")
	}
	connection, readWriter, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("Unable to hijack connection: %v", err)
	}
	return &WS{
		Connection: connection,
		ReadWriter: readWriter,
		Header:     r.Header,
	}, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// log all the headers for debugging
	fmt.Printf("WEBSOCKET INITIALIZING -- %v %s %v\n", r.Method, r.URL, r.Proto)
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}
	ws, err := NewWS(w, r)
	if err != nil {
		fmt.Printf("Error Creating WS Handler: %v\n", err)
		return
	}
	fmt.Printf("Beginning Handshake\n")
	if err = ws.Handshake(); err != nil {
		fmt.Printf("Error Handshaking: %v\n", err)
		return
	}
	fmt.Printf("Handshake Complete\n")
	for {
		frame, err := ws.Read()
		if err != nil {
			fmt.Printf("Error Reading: %v\n", err)
			return
		}
		frameBytes, err := json.MarshalIndent(frame, "", "  ")
		if err != nil {
			fmt.Printf("Error Marshaling Frame: %v\n", err)
			return
		}
		fmt.Printf("Frame: %v\n", string(frameBytes))
	}
}

func Server() {
	http.HandleFunc("/ws", wsHandler)
}

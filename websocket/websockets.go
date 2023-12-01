package websocket

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
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
	Payload []byte `json:"payload"`
}

func newCloseFrame(statusCode uint16) Frame {
	return Frame{
		Fin:     true,
		Opcode:  OPCODE_CLOSE,
		Length:  2,
		Payload: binary.BigEndian.AppendUint16(make([]byte, 0), statusCode),
	}
}

var (
	// https://datatracker.ietf.org/doc/html/rfc6455#section-5.5.1
	CLOSE_FRAME            = newCloseFrame(STATUS_CODE_NORMAL_CLOSURE)
	CLOSE_GOING_AWAY_FRAME = newCloseFrame(STATUS_CODE_GOING_AWAY)
)

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
	opcode := o1 & 0x0F
	isMasked := o2&0x80 != 0
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
	maskKey := binary.BigEndian.Uint32(maskBytes)
	payload := make([]byte, length)
	io.ReadFull(ws.ReadWriter, payload)
	mask(maskKey, payload)

	if opcode == OPCODE_CLOSE {
		status_code := binary.BigEndian.Uint16(payload)
		fmt.Printf("STATUS_CODE: %v\n", status_code)
	}

	return Frame{
		Fin:     fin,
		Rsv1:    rsv1,
		Rsv2:    rsv2,
		Rsv3:    rsv3,
		Opcode:  opcode,
		Masked:  isMasked,
		Mask:    maskKey,
		Length:  length,
		Payload: payload,
	}, nil
}

func (ws *WS) Send(fr Frame) error {
	buffer := bytes.Buffer{}
	firstByte := byte(0)
	if fr.Fin {
		firstByte |= 0x80
	}
	firstByte |= fr.Opcode
	buffer.WriteByte(firstByte)

	secondByte := byte(0)
	if fr.Masked {
		secondByte |= 0x80
	}
	if fr.Length < 126 {
		secondByte |= byte(fr.Length)
		buffer.WriteByte(secondByte)
	} else if fr.Length < 65536 {
		secondByte |= 126
		buffer.WriteByte(secondByte)
		binary.Write(&buffer, binary.BigEndian, uint16(fr.Length))
	} else {
		secondByte |= 127
		buffer.WriteByte(secondByte)
		binary.Write(&buffer, binary.BigEndian, uint64(fr.Length))
	}

	fmt.Printf("TRACE: first two bytes: %08b %08b\n", firstByte, secondByte)
	if fr.Masked {
		binary.Write(&buffer, binary.BigEndian, fr.Mask)
	}

	// fmt.Printf("TRACE: payload: %v\n", fr.Payload)
	payload := fr.Payload
	if fr.Masked {
		mask(fr.Mask, payload)
	}
	buffer.Write(payload)
	ws.ReadWriter.Write(buffer.Bytes())
	ws.ReadWriter.Flush()
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

func (ws *WS) Close() error {
	fmt.Printf("TRACE: Sending Close Frame\n")
	if err := ws.Send(CLOSE_FRAME); err != nil {
		return fmt.Errorf("Error Sending Close Frame: %v", err)
	}
	fmt.Printf("TRACE: Closing Connection\n")
	if err := ws.Connection.Close(); err != nil {
		return fmt.Errorf("Error Closing Connection: %v", err)
	}
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
	fmt.Printf("Reading Frames\n")
	for {
		frame, err := ws.Read()
		if err != nil {
			fmt.Printf("Error Reading: %v\n", err)
			return
		}
		if frame.Opcode == OPCODE_CLOSE {
			fmt.Printf("Closing Connection\n")
			err := ws.Send(CLOSE_FRAME)
			if err != nil {
				fmt.Printf("Error Sending Close Frame: %v\n", err)
			}
			return
		}
		outFrame := Frame{}
		outFrame.Fin = true
		outFrame.Opcode = OPCODE_TEXT
		outFrame.Payload = []byte("ECHO: " + string(frame.Payload))
		outFrame.Length = uint64(len(outFrame.Payload))
		ws.Send(outFrame)
	}
}

func Server() {
	http.HandleFunc("/ws", wsHandler)
}

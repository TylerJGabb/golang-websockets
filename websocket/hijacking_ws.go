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
	HANDSHAKE_KEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

type HttpHijackingWebSocket struct {
	connection net.Conn
	rw         *bufio.ReadWriter
	header     http.Header
}

func NewHttpHijackingWebSocket(w http.ResponseWriter, r *http.Request) (*HttpHijackingWebSocket, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("Unable to typeassert http.ResponseWriter to http.Hijacker")
	}
	connection, rw, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("Unable to hijack HTTP connection: %v", err)
	}
	return &HttpHijackingWebSocket{
		connection: connection,
		rw:         rw,
		header:     r.Header,
	}, nil
}

func (ws *HttpHijackingWebSocket) Handshake() error {
	key := ws.header.Get("Sec-WebSocket-Key")
	if key == "" {
		return fmt.Errorf("No Sec-WebSocket-Key header found")
	}
	sha := sha1.New()
	sha.Write([]byte(key))
	sha.Write([]byte(HANDSHAKE_KEY))
	hash := base64.StdEncoding.EncodeToString(sha.Sum(nil))
	var header http.Header = make(http.Header)
	header.Add("Upgrade", "websocket")
	header.Add("Connection", "Upgrade")
	header.Add("Sec-WebSocket-Accept", hash)
	ws.rw.Write([]byte("HTTP/1.1 101 Switching Protocols\n"))
	header.Write(ws.rw)
	ws.rw.Write([]byte("\n"))
	err := ws.rw.Flush()
	if err != nil {
		return fmt.Errorf("Error flushing http.ResponseWriter: %v", err)
	}
	fmt.Println("Handshake complete")
	return nil
}

func (ws *HttpHijackingWebSocket) ForceClose() error {
	fmt.Printf("WARN: Closing Connection Forcefully\n")
	return ws.connection.Close()
}

func (ws *HttpHijackingWebSocket) ReadFrame() (WebSocketFrame, error) {
	var frame WebSocketFrame
	var err error
	header, err := ws.readHeader()
	if err != nil {
		return frame, fmt.Errorf("Error reading header: %v", err)
	}
	frame.Header = header
	frame.Payload, err = ws.readPayload(header)
	if err != nil {
		return frame, fmt.Errorf("Error reading payload: %v", err)
	}
	return frame, nil
}

func (ws *HttpHijackingWebSocket) WriteFrame(frame WebSocketFrame) error {
	buffer := bytes.Buffer{}
	buffer.Write(frame.Header.ForWire())
	if frame.Header.Masked {
		mask(frame.Header.Mask, frame.Payload)
	}
	buffer.Write(frame.Payload)
	ws.rw.Write(buffer.Bytes())
	return ws.rw.Flush()
}

func (ws *HttpHijackingWebSocket) readHeader() (FrameHeader, error) {
	buf := make([]byte, 2)
	_, err := io.ReadFull(ws.rw, buf)
	if err != nil {
		return FrameHeader{}, err
	}
	fmt.Printf("TRACE: first two bytes: %08b %08b\n", buf[0], buf[1])

	section0 := buf[0]
	section1 := buf[1]
	fin := section0&0x80 != 0
	rsv1 := section0&0x40 != 0
	rsv2 := section0&0x20 != 0
	rsv3 := section0&0x10 != 0
	opcode := section0 & 0x0F
	isMasked := section1&0x80 != 0

	fmt.Printf("TRACE: fin: %v\n", fin)
	fmt.Printf("TRACE: rsv1: %v\n", rsv1)
	fmt.Printf("TRACE: rsv2: %v\n", rsv2)
	fmt.Printf("TRACE: rsv3: %v\n", rsv3)
	fmt.Printf("TRACE: opcode: %v\n", opcode)
	fmt.Printf("TRACE: isMasked: %v\n", isMasked)

	length := uint64(section1 & 0x7f)
	if length == 126 {
		// length is held in the next 2 bytes
		lengthBytes := make([]byte, 2)
		io.ReadFull(ws.rw, lengthBytes)
		length = uint64(binary.BigEndian.Uint16(lengthBytes))
	} else if length == 127 {
		// length is held in the next 8 bytes
		lengthBytes := make([]byte, 8)
		io.ReadFull(ws.rw, lengthBytes)
		length = binary.BigEndian.Uint64(lengthBytes)
	}
	fmt.Printf("TRACE: length: %v\n", length)

	maskBytes := make([]byte, 4)
	if isMasked {
		io.ReadFull(ws.rw, maskBytes)
	}
	maskKey := binary.BigEndian.Uint32(maskBytes)
	fmt.Printf("TRACE: maskKey: %08b_%08b_%08b_%08b\n", maskBytes[0], maskBytes[1], maskBytes[2], maskBytes[3])
	return FrameHeader{
		Fin:    fin,
		Rsv1:   rsv1,
		Rsv2:   rsv2,
		Rsv3:   rsv3,
		Opcode: opcode,
		Masked: isMasked,
		Mask:   maskKey,
		Length: length,
	}, nil
}

func (ws *HttpHijackingWebSocket) readPayload(header FrameHeader) ([]byte, error) {
	if header.Length == 0 {
		return make([]byte, 0), nil
	}
	payload := make([]byte, header.Length)
	io.ReadFull(ws.rw, payload)
	if header.Masked {
		mask(header.Mask, payload)
	}
	return payload, nil
}

func mask(key uint32, payload []byte) error {
	for i, octet := range payload {
		j := i % 4
		shiftDistance := 3 - j
		transform := byte((key >> (shiftDistance * 8)) & 0xFF)
		after := octet ^ transform
		payload[i] = after
	}
	return nil
}

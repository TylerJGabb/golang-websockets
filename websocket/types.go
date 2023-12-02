package websocket

import (
	"bytes"
	"encoding/binary"
)

type FrameHeader struct {
	Fin    bool
	Rsv1   bool
	Rsv2   bool
	Rsv3   bool
	Opcode byte
	Masked bool
	Mask   uint32
	Length uint64
}

type WebSocketFrame struct {
	Header  FrameHeader
	Payload []byte
}

type WebSocket interface {
	Handshake() error
	ReadFrame() (WebSocketFrame, error)
	WriteFrame(frame WebSocketFrame) error
	SendCloseFrame(statusCode uint16) error
	ForceClose() error
}

func (fh *FrameHeader) ForWire() []byte {
	buffer := bytes.Buffer{}
	firstByte := byte(0)
	if fh.Fin {
		firstByte |= 0x80
	}
	firstByte |= fh.Opcode
	buffer.WriteByte(firstByte)

	secondByte := byte(0)
	if fh.Masked {
		secondByte |= 0x80
	}
	if fh.Length < 126 {
		secondByte |= byte(fh.Length)
		buffer.WriteByte(secondByte)
	} else if fh.Length < 65536 {
		secondByte |= 126
		buffer.WriteByte(secondByte)
		binary.Write(&buffer, binary.BigEndian, uint16(fh.Length))
	} else {
		secondByte |= 127
		buffer.WriteByte(secondByte)
		binary.Write(&buffer, binary.BigEndian, uint64(fh.Length))
	}
	if fh.Masked {
		binary.Write(&buffer, binary.BigEndian, fh.Mask)
	}
	return buffer.Bytes()
}

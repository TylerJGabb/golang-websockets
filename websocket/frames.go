package websocket

import "encoding/binary"

func NewCloseFrameHelper(statusCode uint16) WebSocketFrame {
	header := FrameHeader{
		Fin:    true,
		Opcode: OPCODE_CLOSE,
		Length: 2,
	}
	return WebSocketFrame{
		Header:  header,
		Payload: binary.BigEndian.AppendUint16(make([]byte, 0), statusCode),
	}
}

var (
	CLOSE_FRAME_NORMAL           = NewCloseFrameHelper(STATUS_CODE_ABNORMAL_CLOSURE)
	CLOSE_FRAME_GOING_AWAY       = NewCloseFrameHelper(STATUS_CODE_GOING_AWAY)
	CLOSE_FRAME_PROTOCOL_ERROR   = NewCloseFrameHelper(STATUS_CODE_PROTOCOL_ERROR)
	CLOSE_FRAME_UNSUPPORTED      = NewCloseFrameHelper(STATUS_CODE_UNSUPPORTED_DATA_TYPE)
	CLOSE_FRAME_RESERVED         = NewCloseFrameHelper(STATUS_CODE_RESERVED)
	CLOSE_FRAME_NO_STATUS_CODE   = NewCloseFrameHelper(STATUS_CODE_NO_STATUS_CODE_PRESENT)
	CLOSE_FRAME_ABNORMAL         = NewCloseFrameHelper(STATUS_CODE_ABNORMAL_CLOSURE)
	CLOSE_FRAME_INVALID_PAYLOAD  = NewCloseFrameHelper(STATUS_CODE_INVALID_FRAME_PAYLOAD_DATA)
	CLOSE_FRAME_POLICY_VIOLATION = NewCloseFrameHelper(STATUS_CODE_POLICY_VIOLATION)
	CLOSE_FRAME_MESSAGE_TOO_BIG  = NewCloseFrameHelper(STATUS_CODE_MESSAGE_TOO_BIG)
	CLOSE_FRAME_MANDATORY        = NewCloseFrameHelper(STATUS_CODE_MANDATORY_EXTENSION)
	CLOSE_FRAME_INTERNAL_ERROR   = NewCloseFrameHelper(STATUS_CODE_INTERNAL_SERVER_ERROR)
	CLOSE_FRAME_TLS_HANDSHAKE    = NewCloseFrameHelper(STATUS_CODE_TLS_HANDSHAKE)
)

var PING_FRAME = WebSocketFrame{
	Header: FrameHeader{
		Fin:    true,
		Opcode: OPCODE_PING,
	},
	Payload: []byte{},
}

var PONG_FRAME = WebSocketFrame{
	Header: FrameHeader{
		Fin:    true,
		Opcode: OPCODE_PONG,
	},
	Payload: []byte{},
}

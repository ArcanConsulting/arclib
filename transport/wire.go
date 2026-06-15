package transport

import (
	"errors"

	"github.com/gorilla/websocket"

	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

// writeTier0 marshals a plaintext (Tier 0) message and writes it as a binary
// WebSocket frame. Tier-0 framing is used only for the SESSION_INIT/SESSION_ACK
// handshake; all subsequent traffic is Tier-3 AEAD frames written directly.
func writeTier0(ws *websocket.Conn, op protocol.OpCode, payload []byte) error {
	data, err := protocol.NewMessage(protocol.TierPlaintext, op, payload).Marshal()
	if err != nil {
		return err
	}
	return ws.WriteMessage(websocket.BinaryMessage, data)
}

// readTier0 reads one binary WebSocket frame and parses it as a plaintext message.
func readTier0(ws *websocket.Conn) (protocol.OpCode, []byte, error) {
	_, data, err := ws.ReadMessage()
	if err != nil {
		return 0, nil, err
	}
	var msg protocol.Message
	if err := msg.Unmarshal(data); err != nil {
		return 0, nil, err
	}
	return msg.Header.OpCode, msg.Payload, nil
}

// errorFrame is the OpError payload exchanged when a handler fails.
type errorFrame struct {
	Code    uint32 `msgpack:"code"`
	Message string `msgpack:"msg"`
}

func encodeError(err error) []byte {
	data, mErr := msgpack.Marshal(errorFrame{Message: err.Error()})
	if mErr != nil {
		return nil
	}
	return data
}

func decodeError(payload []byte) error {
	var ef errorFrame
	if err := msgpack.Unmarshal(payload, &ef); err != nil || ef.Message == "" {
		return errors.New("transport: remote error")
	}
	return errors.New(ef.Message)
}

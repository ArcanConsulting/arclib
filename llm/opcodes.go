// Package llm defines the generic LLM-over-MyClerk-Protocol wire contract:
// chat, embeddings and text-to-speech request/response types, plus the Maya
// AI Assistant operation codes registered in draft-myclerk-protocol-ops-01
// (Section 3.10, range 0x1000-0x10FF).
//
// These types are the single source of truth for the wire format shared by
// every consumer (MyClerk core, mayaservices, ArcShell). The opcode constants
// mirror the IETF registry; cmd/opcode-verify enforces parity with the draft.
package llm

import "arcan-it.de/arclib/protocol"

// Maya AI Assistant operations (draft-myclerk-protocol-ops-01, Section 3.10).
const (
	OpMayaQuery     protocol.OpCode = 0x1000 // MAYA_QUERY      — chat completion request
	OpMayaResponse  protocol.OpCode = 0x1001 // MAYA_RESPONSE   — chat completion reply
	OpMayaStream    protocol.OpCode = 0x1002 // MAYA_STREAM     — streaming chat chunk
	OpMayaTTS       protocol.OpCode = 0x1032 // MAYA_T_T_S      — text-to-speech
	OpMayaModelsGet protocol.OpCode = 0x1060 // MAYA_MODELS_GET — list available models
	OpMayaHealthGet protocol.OpCode = 0x1061 // MAYA_HEALTH_GET — provider health
	OpMayaEmbed     protocol.OpCode = 0x1080 // MAYA_EMBED      — text embeddings
)

// opCodeNames maps the llm opcodes to their IETF registry names.
var opCodeNames = map[protocol.OpCode]string{
	OpMayaQuery:     "MAYA_QUERY",
	OpMayaResponse:  "MAYA_RESPONSE",
	OpMayaStream:    "MAYA_STREAM",
	OpMayaTTS:       "MAYA_T_T_S",
	OpMayaModelsGet: "MAYA_MODELS_GET",
	OpMayaHealthGet: "MAYA_HEALTH_GET",
	OpMayaEmbed:     "MAYA_EMBED",
}

func init() {
	protocol.RegisterOpCodes("llm", opCodeNames)
}

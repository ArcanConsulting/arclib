package llm

import (
	"testing"

	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

func TestOpCodeNamesRegistered(t *testing.T) {
	cases := map[protocol.OpCode]string{
		OpMayaQuery:     "MAYA_QUERY",
		OpMayaResponse:  "MAYA_RESPONSE",
		OpMayaStream:    "MAYA_STREAM",
		OpMayaTTS:       "MAYA_T_T_S",
		OpMayaModelsGet: "MAYA_MODELS_GET",
		OpMayaHealthGet: "MAYA_HEALTH_GET",
		OpMayaEmbed:     "MAYA_EMBED",
	}
	for op, want := range cases {
		if got := op.String(); got != want {
			t.Errorf("op 0x%04X: String()=%q, want %q", uint16(op), got, want)
		}
		if got := protocol.LookupOpCode(op); got != want {
			t.Errorf("op 0x%04X: LookupOpCode=%q, want %q", uint16(op), got, want)
		}
	}
}

func TestOpCodesMatchDraftRange(t *testing.T) {
	// Maya AI Assistant is registered to 0x1000-0x10FF in draft-myclerk-protocol-ops-01.
	for _, op := range []protocol.OpCode{
		OpMayaQuery, OpMayaResponse, OpMayaStream, OpMayaTTS,
		OpMayaModelsGet, OpMayaHealthGet, OpMayaEmbed,
	} {
		if op < 0x1000 || op > 0x10FF {
			t.Errorf("op %s (0x%04X) outside Maya range 0x1000-0x10FF", op, uint16(op))
		}
		if op.Category() != 0x10 {
			t.Errorf("op %s: category 0x%02X, want 0x10", op, op.Category())
		}
	}
}

func TestChatRequestRoundTrip(t *testing.T) {
	req := ChatRequest{
		Model: "gemma4-128k",
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: "be helpful"},
			{Role: RoleUser, Content: "what time is it?"},
		},
		Tools: []Tool{{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_time",
				Description: "current time",
				Parameters:  map[string]any{"type": "object"},
			},
		}},
		MaxTokens: 256,
		Stream:    true,
		Options:   map[string]any{"top_p": 0.9},
	}
	var got ChatRequest
	roundTrip(t, req, &got)
	if got.Model != req.Model || len(got.Messages) != 2 || !got.Stream || got.MaxTokens != 256 {
		t.Fatalf("mismatch: %+v", got)
	}
	if len(got.Tools) != 1 || got.Tools[0].Function.Name != "get_time" {
		t.Fatalf("tools lost: %+v", got.Tools)
	}
}

func TestChatResponseRoundTrip(t *testing.T) {
	resp := ChatResponse{
		Model: "gemma4-128k",
		Message: ChatMessage{
			Role: RoleAssistant,
			ToolCalls: []ToolCall{{
				ID:       "call_1",
				Type:     "function",
				Function: FunctionCall{Name: "get_time", Arguments: `{"tz":"UTC"}`},
			}},
		},
		FinishReason: FinishToolCalls,
		Usage:        Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		CreatedNs:    1234567890,
	}
	var got ChatResponse
	roundTrip(t, resp, &got)
	if got.FinishReason != FinishToolCalls || got.Usage.TotalTokens != 15 {
		t.Fatalf("mismatch: %+v", got)
	}
	if len(got.Message.ToolCalls) != 1 || got.Message.ToolCalls[0].Function.Arguments != `{"tz":"UTC"}` {
		t.Fatalf("tool call lost: %+v", got.Message)
	}
}

func TestStreamChunkRoundTrip(t *testing.T) {
	chunk := StreamChunk{
		Delta:        ChatMessage{Role: RoleAssistant, Content: "Hel"},
		FinishReason: FinishStop,
		Usage:        &Usage{TotalTokens: 3},
		Done:         true,
	}
	var got StreamChunk
	roundTrip(t, chunk, &got)
	if !got.Done || got.Delta.Content != "Hel" || got.Usage == nil || got.Usage.TotalTokens != 3 {
		t.Fatalf("mismatch: %+v", got)
	}
}

func TestEmbedRoundTrip(t *testing.T) {
	req := EmbedRequest{Model: "nomic-embed-text", Input: []string{"a", "b"}}
	var gotReq EmbedRequest
	roundTrip(t, req, &gotReq)
	if gotReq.Model != req.Model || len(gotReq.Input) != 2 {
		t.Fatalf("req mismatch: %+v", gotReq)
	}
	resp := EmbedResponse{
		Model:      "nomic-embed-text",
		Embeddings: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		Dimensions: 2,
		Usage:      Usage{TotalTokens: 4},
	}
	var gotResp EmbedResponse
	roundTrip(t, resp, &gotResp)
	if gotResp.Dimensions != 2 || len(gotResp.Embeddings) != 2 || gotResp.Embeddings[1][1] != 0.4 {
		t.Fatalf("resp mismatch: %+v", gotResp)
	}
}

func TestModelsAndHealthRoundTrip(t *testing.T) {
	list := ListModelsResponse{Models: []ModelInfo{{
		Name:          "gemma4-128k",
		ContextLength: 262144,
		Capabilities:  []string{CapCompletion, CapTools, CapVision},
	}}}
	var gotList ListModelsResponse
	roundTrip(t, list, &gotList)
	if len(gotList.Models) != 1 || gotList.Models[0].ContextLength != 262144 {
		t.Fatalf("list mismatch: %+v", gotList)
	}

	h := HealthResponse{Healthy: true, Provider: "ollama", LoadedModels: []string{"gemma4-128k"}}
	var gotH HealthResponse
	roundTrip(t, h, &gotH)
	if !gotH.Healthy || gotH.Provider != "ollama" || len(gotH.LoadedModels) != 1 {
		t.Fatalf("health mismatch: %+v", gotH)
	}
}

func TestTTSRoundTrip(t *testing.T) {
	req := TTSRequest{Text: "hallo", Voice: "maya", Engine: "xtts", Language: "de", Format: "opus", Speed: 1.0}
	var gotReq TTSRequest
	roundTrip(t, req, &gotReq)
	if gotReq.Text != "hallo" || gotReq.Engine != "xtts" || gotReq.Format != "opus" {
		t.Fatalf("req mismatch: %+v", gotReq)
	}
	resp := TTSResponse{Audio: []byte{1, 2, 3}, Format: "opus", SampleRate: 24000, DurationMs: 120}
	var gotResp TTSResponse
	roundTrip(t, resp, &gotResp)
	if len(gotResp.Audio) != 3 || gotResp.SampleRate != 24000 {
		t.Fatalf("resp mismatch: %+v", gotResp)
	}
}

func TestModelInfoHasCapability(t *testing.T) {
	m := ModelInfo{Capabilities: []string{CapTools, CapVision}}
	if !m.HasCapability(CapTools) || !m.HasCapability(CapVision) {
		t.Error("expected tools+vision")
	}
	if m.HasCapability(CapEmbed) {
		t.Error("did not expect embed")
	}
	if (ModelInfo{}).HasCapability(CapTools) {
		t.Error("empty model should have no capabilities")
	}
}

// roundTrip marshals v and unmarshals into dst, failing the test on error.
func roundTrip(t *testing.T, v, dst any) {
	t.Helper()
	data, err := msgpack.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := msgpack.Unmarshal(data, dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

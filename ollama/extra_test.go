package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arcan-it.de/arclib/llm"
)

func TestClientOptions(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := New("http://x", WithHTTPClient(custom), WithTimeout(9*time.Second))
	if c.hc != custom {
		t.Fatal("custom http client not applied")
	}
	if c.hc.Timeout != 9*time.Second {
		t.Fatalf("timeout = %v", c.hc.Timeout)
	}
	// nil client must be ignored.
	c2 := New("http://x", WithHTTPClient(nil))
	if c2.hc == nil {
		t.Fatal("nil client should be ignored")
	}
}

func TestChatNilToolArguments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"m","message":{"role":"assistant","tool_calls":[{"function":{"name":"f"}}]},"done":true}`))
	}))
	defer srv.Close()
	resp, err := New(srv.URL).Chat(context.Background(), llm.ChatRequest{Model: "m"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if len(resp.Message.ToolCalls) != 1 || resp.Message.ToolCalls[0].Function.Arguments != "{}" {
		t.Fatalf("expected empty-object args, got %+v", resp.Message.ToolCalls)
	}
}

func TestFromLLMMessageToolCalls(t *testing.T) {
	var got wireChatReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decodeBody(t, r, &got)
		_, _ = w.Write([]byte(`{"done":true}`))
	}))
	defer srv.Close()
	_, err := New(srv.URL).Chat(context.Background(), llm.ChatRequest{
		Model: "m",
		Messages: []llm.ChatMessage{
			{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
				Type: "function", Function: llm.FunctionCall{Name: "f", Arguments: `{"a":1}`},
			}}},
			{Role: llm.RoleTool, Content: "result"},
		},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if len(got.Messages) != 2 || len(got.Messages[0].ToolCalls) != 1 {
		t.Fatalf("messages not forwarded: %+v", got.Messages)
	}
	if got.Messages[0].ToolCalls[0].Function.Arguments["a"] != float64(1) {
		t.Fatalf("arguments not parsed to object: %+v", got.Messages[0].ToolCalls[0].Function.Arguments)
	}
}

func TestShowContextLengthEdgeCases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// context_length present but non-numeric; plus an unrelated key.
		_, _ = w.Write([]byte(`{"capabilities":["embed"],"model_info":{"general.architecture":"gemma","gemma.context_length":"bad"}}`))
	}))
	defer srv.Close()
	info, err := New(srv.URL).Show(context.Background(), "m")
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if info.ContextLength != 0 {
		t.Fatalf("expected 0 context length for non-numeric, got %d", info.ContextLength)
	}
	if !info.HasCapability(llm.CapEmbed) {
		t.Fatal("expected embed capability")
	}
}

func TestDecodeErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()
	c := New(srv.URL)
	if _, err := c.Chat(context.Background(), llm.ChatRequest{}); err == nil {
		t.Error("Chat: expected decode error")
	}
	if _, err := c.List(context.Background()); err == nil {
		t.Error("List: expected decode error")
	}
	if _, err := c.Show(context.Background(), "m"); err == nil {
		t.Error("Show: expected decode error")
	}
	if _, err := c.Embed(context.Background(), llm.EmbedRequest{}); err == nil {
		t.Error("Embed: expected decode error")
	}
	if err := c.ChatStream(context.Background(), llm.ChatRequest{}, func(llm.StreamChunk) error { return nil }); err == nil {
		t.Error("ChatStream: expected decode error")
	}
}

func TestListAndEmbedUnreachable(t *testing.T) {
	c := New("http://127.0.0.1:0")
	if _, err := c.List(context.Background()); err == nil {
		t.Error("List: expected unreachable error")
	}
	if _, err := c.Embed(context.Background(), llm.EmbedRequest{Model: "m"}); err == nil {
		t.Error("Embed: expected unreachable error")
	}
}

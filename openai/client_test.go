package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"arcan-it.de/arclib/llm"
)

func decodeBody(t *testing.T, r *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("server decode: %v", err)
	}
}

func TestChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		decodeBody(t, r, &body)
		if body["model"] != "gemma4" {
			t.Errorf("model = %v", body["model"])
		}
		if body["stream"] != false {
			t.Errorf("stream = %v, want false", body["stream"])
		}
		if body["temperature"] == nil || body["max_tokens"] == nil {
			t.Errorf("temp/max_tokens not set: %v", body)
		}
		// echo an assistant tool call back
		_, _ = w.Write([]byte(`{
			"model": "gemma4",
			"choices": [{
				"message": {"role": "assistant", "content": "",
					"tool_calls": [{"id": "call_0", "type": "function",
						"function": {"name": "get_time", "arguments": "{\"tz\":\"UTC\"}"}}]},
				"finish_reason": "tool_calls"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer srv.Close()

	c := New(srv.URL + "/v1")
	resp, err := c.Chat(context.Background(), llm.ChatRequest{
		Model:       "gemma4",
		Messages:    []llm.ChatMessage{{Role: llm.RoleUser, Content: "time?"}},
		Temperature: 0.7,
		MaxTokens:   128,
		Tools:       []llm.Tool{{Type: "function", Function: llm.ToolFunction{Name: "get_time"}}},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.FinishReason != llm.FinishToolCalls {
		t.Errorf("finish = %s", resp.FinishReason)
	}
	if len(resp.Message.ToolCalls) != 1 {
		t.Fatalf("tool calls = %d", len(resp.Message.ToolCalls))
	}
	tc := resp.Message.ToolCalls[0]
	if tc.Function.Name != "get_time" || tc.Function.Arguments != `{"tz":"UTC"}` {
		t.Errorf("tool call = %+v", tc)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"Hel\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":2,\"total_tokens\":5}}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	c := New(srv.URL + "/v1")
	var content strings.Builder
	var final llm.StreamChunk
	err := c.ChatStream(context.Background(), llm.ChatRequest{
		Model:    "gemma4",
		Messages: []llm.ChatMessage{{Role: llm.RoleUser, Content: "hi"}},
	}, func(chunk llm.StreamChunk) error {
		if chunk.Done {
			final = chunk
			return nil
		}
		content.WriteString(chunk.Delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	if content.String() != "Hello" {
		t.Errorf("content = %q", content.String())
	}
	if !final.Done || final.FinishReason != llm.FinishStop {
		t.Errorf("final = %+v", final)
	}
	if final.Usage == nil || final.Usage.TotalTokens != 5 {
		t.Errorf("final usage = %+v", final.Usage)
	}
}

func TestEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		decodeBody(t, r, &body)
		if body["model"] != "nomic-embed-text" {
			t.Errorf("model = %v", body["model"])
		}
		_, _ = w.Write([]byte(`{"model":"nomic-embed-text","data":[{"embedding":[0.1,0.2,0.3]}],"usage":{"prompt_tokens":4,"total_tokens":4}}`))
	}))
	defer srv.Close()

	c := New(srv.URL + "/v1")
	resp, err := c.Embed(context.Background(), llm.EmbedRequest{Model: "nomic-embed-text", Input: []string{"hello"}})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if resp.Dimensions != 3 || len(resp.Embeddings) != 1 || len(resp.Embeddings[0]) != 3 {
		t.Errorf("embeddings = %+v (dim %d)", resp.Embeddings, resp.Dimensions)
	}
}

func TestListAndHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"gemma4","object":"model"},{"id":"qwen","object":"model"}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL + "/v1")
	list, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Models) != 2 || list.Models[0].Name != "gemma4" {
		t.Errorf("models = %+v", list.Models)
	}

	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !h.Healthy || h.Provider != "openai" {
		t.Errorf("health = %+v", h)
	}
}

func TestHealthUnreachable(t *testing.T) {
	c := New("http://127.0.0.1:1/v1") // nothing listening
	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned transport error: %v", err)
	}
	if h.Healthy {
		t.Error("expected Healthy=false for unreachable backend")
	}
}

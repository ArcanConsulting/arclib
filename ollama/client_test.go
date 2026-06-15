package ollama

import (
	"context"
	"encoding/json"
	"io"
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
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req wireChatReq
		decodeBody(t, r, &req)
		if req.Model != "gemma4-128k" {
			t.Errorf("model = %s", req.Model)
		}
		if req.Stream {
			t.Error("expected non-stream")
		}
		if req.Options["temperature"] == nil || req.Options["num_predict"] == nil {
			t.Errorf("options not mapped: %v", req.Options)
		}
		// echo a tool call back
		_ = json.NewEncoder(w).Encode(wireChatResp{
			Model: "gemma4-128k",
			Message: wireMessage{Role: "assistant", ToolCalls: []wireToolCall{{
				Function: wireFuncCall{Name: "get_time", Arguments: map[string]any{"tz": "UTC"}},
			}}},
			Done: true, DoneReason: "stop",
			PromptEvalCount: 10, EvalCount: 5,
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	resp, err := c.Chat(context.Background(), llm.ChatRequest{
		Model:       "gemma4-128k",
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
	if len(resp.Message.ToolCalls) != 1 || resp.Message.ToolCalls[0].Function.Arguments != `{"tz":"UTC"}` {
		t.Errorf("tool call = %+v", resp.Message.ToolCalls)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestChatError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(wireChatResp{Error: "model not found"})
	}))
	defer srv.Close()
	_, err := New(srv.URL).Chat(context.Background(), llm.ChatRequest{Model: "x"})
	if err == nil || !strings.Contains(err.Error(), "model not found") {
		t.Fatalf("expected ollama error, got %v", err)
	}
}

func TestChatHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	_, err := New(srv.URL).Chat(context.Background(), llm.ChatRequest{Model: "x"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected http error, got %v", err)
	}
}

func TestChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req wireChatReq
		decodeBody(t, r, &req)
		if !req.Stream {
			t.Error("expected stream")
		}
		enc := json.NewEncoder(w)
		_ = enc.Encode(wireChatResp{Message: wireMessage{Role: "assistant", Content: "Hel"}})
		_ = enc.Encode(wireChatResp{Message: wireMessage{Content: "lo"}})
		_ = enc.Encode(wireChatResp{Done: true, DoneReason: "stop", PromptEvalCount: 2, EvalCount: 1})
	}))
	defer srv.Close()

	var content strings.Builder
	var doneSeen bool
	err := New(srv.URL).ChatStream(context.Background(),
		llm.ChatRequest{Model: "gemma4-128k", Messages: []llm.ChatMessage{{Role: llm.RoleUser, Content: "hi"}}},
		func(chunk llm.StreamChunk) error {
			content.WriteString(chunk.Delta.Content)
			if chunk.Done {
				doneSeen = true
				if chunk.Usage == nil || chunk.Usage.TotalTokens != 3 {
					t.Errorf("final usage = %+v", chunk.Usage)
				}
			}
			return nil
		})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}
	if content.String() != "Hello" || !doneSeen {
		t.Errorf("content=%q doneSeen=%v", content.String(), doneSeen)
	}
}

func TestChatStreamCallbackError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(wireChatResp{Message: wireMessage{Content: "x"}})
		_ = json.NewEncoder(w).Encode(wireChatResp{Done: true})
	}))
	defer srv.Close()
	sentinel := io.EOF
	err := New(srv.URL).ChatStream(context.Background(), llm.ChatRequest{Model: "x"},
		func(llm.StreamChunk) error { return sentinel })
	if err != sentinel {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"models":[{"name":"gemma4-128k","details":{"family":"gemma","parameter_size":"11.9B","quantization_level":"Q4_K_M"}}]}`))
	}))
	defer srv.Close()
	resp, err := New(srv.URL).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Models) != 1 || resp.Models[0].Name != "gemma4-128k" || resp.Models[0].Quantization != "Q4_K_M" {
		t.Errorf("models = %+v", resp.Models)
	}
}

func TestShow(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		decodeBody(t, r, &req)
		if req["model"] != "gemma4-128k" {
			t.Errorf("model = %v", req["model"])
		}
		_, _ = w.Write([]byte(`{"capabilities":["completion","tools","vision"],"details":{"family":"gemma"},"model_info":{"gemma.context_length":262144}}`))
	}))
	defer srv.Close()
	info, err := New(srv.URL).Show(context.Background(), "gemma4-128k")
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if !info.HasCapability(llm.CapTools) || info.ContextLength != 262144 {
		t.Errorf("info = %+v", info)
	}
}

func TestEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"model":"nomic-embed-text","embeddings":[[0.1,0.2,0.3]],"prompt_eval_count":4}`))
	}))
	defer srv.Close()
	resp, err := New(srv.URL).Embed(context.Background(), llm.EmbedRequest{Model: "nomic-embed-text", Input: []string{"hi"}})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if resp.Dimensions != 3 || len(resp.Embeddings) != 1 || resp.Usage.TotalTokens != 4 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestEmbedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error":"no embed model"}`))
	}))
	defer srv.Close()
	_, err := New(srv.URL).Embed(context.Background(), llm.EmbedRequest{Model: "x"})
	if err == nil || !strings.Contains(err.Error(), "no embed model") {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`{"version":"0.30.8"}`))
		case "/api/ps":
			_, _ = w.Write([]byte(`{"models":[{"name":"gemma4-128k"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	h, err := New(srv.URL).Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !h.Healthy || h.Detail != "0.30.8" || len(h.LoadedModels) != 1 {
		t.Errorf("health = %+v", h)
	}
}

func TestHealthUnreachable(t *testing.T) {
	c := New("http://127.0.0.1:0") // invalid port -> connection fails
	h, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health should not return error: %v", err)
	}
	if h.Healthy {
		t.Error("expected unhealthy")
	}
}

func TestNewDefaults(t *testing.T) {
	c := New("")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %s", c.baseURL)
	}
	c2 := New("http://x:1/")
	if c2.baseURL != "http://x:1" {
		t.Errorf("trailing slash not trimmed: %s", c2.baseURL)
	}
}

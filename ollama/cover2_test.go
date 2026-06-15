package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"arcan-it.de/arclib/llm"
)

func TestPostMarshalError(t *testing.T) {
	// A channel cannot be JSON-encoded; this exercises post()'s encode-error path.
	c := New("http://127.0.0.1:0")
	_, err := c.Chat(context.Background(), llm.ChatRequest{
		Model:   "m",
		Options: map[string]any{"bad": make(chan int)},
	})
	if err == nil {
		t.Fatal("expected encode error")
	}
}

func TestChatOptionsMerge(t *testing.T) {
	var got wireChatReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decodeBody(t, r, &got)
		_, _ = w.Write([]byte(`{"done":true}`))
	}))
	defer srv.Close()
	_, err := New(srv.URL).Chat(context.Background(), llm.ChatRequest{
		Model:   "m",
		Options: map[string]any{"top_p": 0.9, "seed": 42},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if got.Options["top_p"] != 0.9 || got.Options["seed"] != float64(42) {
		t.Fatalf("options not merged: %+v", got.Options)
	}
}

func TestHealthMalformedPS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = w.Write([]byte(`garbage`)) // decode ignored
		case "/api/ps":
			_, _ = w.Write([]byte(`also-garbage`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	h, err := New(srv.URL).Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !h.Healthy || len(h.LoadedModels) != 0 {
		t.Fatalf("health = %+v", h)
	}
}

// Package openai is a minimal, generic HTTP client for an OpenAI-compatible
// chat/embeddings API (the /v1 surface spoken by llama.cpp's llama-server and
// the mayaservices model manager).
//
// Like arclib/ollama it speaks the arclib/llm wire types, so it can be dropped
// behind any MyClerk-Protocol LLM handler (e.g. mayaservices) without leaking
// backend specifics. Tool calls map 1:1: OpenAI already encodes function
// arguments as a JSON string, exactly like arclib/llm, so no translation is
// needed (unlike the Ollama object-argument form).
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"arcan-it.de/arclib/llm"
)

// DefaultBaseURL is the conventional local model-manager endpoint (mayaservices).
const DefaultBaseURL = "http://127.0.0.1:8200/v1"

// Client talks to an OpenAI-compatible server over HTTP. The base URL must
// include the API version path (e.g. ".../v1").
type Client struct {
	baseURL string
	apiKey  string
	hc      *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithHTTPClient sets the underlying http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.hc = hc
		}
	}
}

// WithTimeout sets the request timeout on a default http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.hc.Timeout = d }
}

// WithAPIKey sets a bearer token sent on every request.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// New creates a Client for the given base URL (defaulting to DefaultBaseURL).
func New(baseURL string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		hc:      &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// --- OpenAI wire types -----------------------------------------------------

type wireMessage struct {
	Role             string         `json:"role"`
	Content          string         `json:"content,omitempty"`
	ReasoningContent string         `json:"reasoning_content,omitempty"`
	Name             string         `json:"name,omitempty"`
	ToolCallID       string         `json:"tool_call_id,omitempty"`
	ToolCalls        []wireToolCall `json:"tool_calls,omitempty"`
}

type wireToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function wireFuncCall `json:"function"`
}

type wireFuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string, OpenAI-compatible
}

type wireTool struct {
	Type     string       `json:"type"`
	Function wireToolFunc `json:"function"`
}

type wireToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type wireUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type wireChatResp struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      wireMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage wireUsage `json:"usage"`
	Error any       `json:"error"`
}

type wireStreamResp struct {
	Choices []struct {
		Delta        wireMessage `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage *wireUsage `json:"usage"`
}

// --- Chat ------------------------------------------------------------------

// Chat performs a non-streaming chat completion.
func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	body, err := c.post(ctx, "/chat/completions", c.buildChatBody(req, false), false)
	if err != nil {
		return llm.ChatResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr wireChatResp
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.ChatResponse{}, fmt.Errorf("decode chat response: %w", err)
	}
	if wr.Error != nil {
		return llm.ChatResponse{}, fmt.Errorf("openai: %v", wr.Error)
	}
	if len(wr.Choices) == 0 {
		return llm.ChatResponse{}, fmt.Errorf("openai: empty choices")
	}
	ch := wr.Choices[0]
	return llm.ChatResponse{
		Model:        wr.Model,
		Message:      toLLMMessage(ch.Message),
		FinishReason: finishReason(ch.FinishReason, ch.Message.ToolCalls),
		Usage: llm.Usage{
			PromptTokens:     wr.Usage.PromptTokens,
			CompletionTokens: wr.Usage.CompletionTokens,
			TotalTokens:      wr.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream performs a streaming chat completion, invoking fn for each chunk.
// Content/tool deltas are forwarded as they arrive; the final chunk has
// Done == true and carries the finish reason plus usage (when the backend
// reports it). If fn returns an error, streaming stops and that error is
// returned. SSE framing ("data: {json}\n\n", terminated by "data: [DONE]") is
// parsed here.
func (c *Client) ChatStream(ctx context.Context, req llm.ChatRequest, fn func(llm.StreamChunk) error) error {
	body, err := c.post(ctx, "/chat/completions", c.buildChatBody(req, true), true)
	if err != nil {
		return err
	}
	defer func() { _ = body.Close() }()

	var (
		usage  *llm.Usage
		finish string
	)
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 8<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var wr wireStreamResp
		if err := json.Unmarshal([]byte(data), &wr); err != nil {
			continue // tolerate keep-alive / non-JSON lines
		}
		if wr.Usage != nil {
			usage = &llm.Usage{
				PromptTokens:     wr.Usage.PromptTokens,
				CompletionTokens: wr.Usage.CompletionTokens,
				TotalTokens:      wr.Usage.TotalTokens,
			}
		}
		if len(wr.Choices) == 0 {
			continue
		}
		choice := wr.Choices[0]
		if choice.FinishReason != "" {
			finish = finishReason(choice.FinishReason, choice.Delta.ToolCalls)
		}
		delta := choice.Delta
		if delta.Content == "" && delta.ReasoningContent == "" && len(delta.ToolCalls) == 0 {
			continue
		}
		if err := fn(llm.StreamChunk{Delta: toLLMMessage(delta)}); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("stream read: %w", err)
	}
	if finish == "" {
		finish = llm.FinishStop
	}
	final := llm.StreamChunk{FinishReason: finish, Usage: usage, Done: true}
	return fn(final)
}

func (c *Client) buildChatBody(req llm.ChatRequest, stream bool) map[string]any {
	msgs := make([]wireMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = fromLLMMessage(m)
	}
	body := map[string]any{
		"model":    req.Model,
		"messages": msgs,
		"stream":   stream,
	}
	if len(req.Tools) > 0 {
		tools := make([]wireTool, 0, len(req.Tools))
		for _, t := range req.Tools {
			typ := t.Type
			if typ == "" {
				typ = "function"
			}
			tools = append(tools, wireTool{
				Type: typ,
				Function: wireToolFunc{
					Name:        t.Function.Name,
					Description: t.Function.Description,
					Parameters:  t.Function.Parameters,
				},
			})
		}
		body["tools"] = tools
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	// Passthrough options (e.g. top_p). Core fields above win on key clash.
	for k, v := range req.Options {
		if _, taken := body[k]; !taken {
			body[k] = v
		}
	}
	return body
}

// --- Models ----------------------------------------------------------------

// List returns the model ids the backend exposes (GET /v1/models).
func (c *Client) List(ctx context.Context) (llm.ListModelsResponse, error) {
	body, err := c.get(ctx, "/models")
	if err != nil {
		return llm.ListModelsResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.ListModelsResponse{}, fmt.Errorf("decode models: %w", err)
	}
	out := llm.ListModelsResponse{Models: make([]llm.ModelInfo, 0, len(wr.Data))}
	for _, m := range wr.Data {
		out.Models = append(out.Models, llm.ModelInfo{Name: m.ID})
	}
	return out, nil
}

// --- Embeddings ------------------------------------------------------------

// Embed generates embeddings for the request inputs (POST /v1/embeddings).
func (c *Client) Embed(ctx context.Context, req llm.EmbedRequest) (llm.EmbedResponse, error) {
	payload := map[string]any{"model": req.Model, "input": req.Input}
	body, err := c.post(ctx, "/embeddings", payload, false)
	if err != nil {
		return llm.EmbedResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr struct {
		Model string `json:"model"`
		Data  []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage wireUsage `json:"usage"`
		Error any       `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.EmbedResponse{}, fmt.Errorf("decode embed: %w", err)
	}
	if wr.Error != nil {
		return llm.EmbedResponse{}, fmt.Errorf("openai: %v", wr.Error)
	}
	embeddings := make([][]float32, 0, len(wr.Data))
	for _, d := range wr.Data {
		embeddings = append(embeddings, d.Embedding)
	}
	dims := 0
	if len(embeddings) > 0 {
		dims = len(embeddings[0])
	}
	model := wr.Model
	if model == "" {
		model = req.Model
	}
	return llm.EmbedResponse{
		Model:      model,
		Embeddings: embeddings,
		Dimensions: dims,
		Usage:      llm.Usage{PromptTokens: wr.Usage.PromptTokens, TotalTokens: wr.Usage.TotalTokens},
	}, nil
}

// --- Health ----------------------------------------------------------------

// Health reports backend reachability by listing models (GET /v1/models). A
// failure is surfaced via Healthy=false rather than as a transport error so
// callers can degrade gracefully.
func (c *Client) Health(ctx context.Context) (llm.HealthResponse, error) {
	models, err := c.List(ctx)
	if err != nil {
		return llm.HealthResponse{Healthy: false, Provider: "openai", Detail: err.Error()}, nil //nolint:nilerr // health failures are surfaced via Healthy=false, not as a transport error
	}
	h := llm.HealthResponse{Healthy: true, Provider: "openai"}
	h.Detail = fmt.Sprintf("%d models available", len(models.Models))
	return h, nil
}

// --- helpers ---------------------------------------------------------------

func (c *Client) post(ctx context.Context, path string, payload any, stream bool) (io.ReadCloser, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	return c.do(req)
}

func (c *Client) get(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) do(req *http.Request) (io.ReadCloser, error) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai unreachable: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("openai %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return resp.Body, nil
}

func finishReason(reason string, toolCalls []wireToolCall) string {
	switch reason {
	case "length":
		return llm.FinishLength
	case "tool_calls":
		return llm.FinishToolCalls
	case "stop":
		return llm.FinishStop
	}
	if len(toolCalls) > 0 {
		return llm.FinishToolCalls
	}
	return llm.FinishStop
}

func toLLMMessage(m wireMessage) llm.ChatMessage {
	out := llm.ChatMessage{Role: m.Role, Content: m.Content, ReasoningContent: m.ReasoningContent, Name: m.Name, ToolCallID: m.ToolCallID}
	for _, tc := range m.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, llm.ToolCall{
			ID:       tc.ID,
			Type:     orFunction(tc.Type),
			Function: llm.FunctionCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments},
		})
	}
	return out
}

func fromLLMMessage(m llm.ChatMessage) wireMessage {
	out := wireMessage{Role: m.Role, Content: m.Content, Name: m.Name, ToolCallID: m.ToolCallID}
	for _, tc := range m.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, wireToolCall{
			ID:       tc.ID,
			Type:     orFunction(tc.Type),
			Function: wireFuncCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments},
		})
	}
	return out
}

func orFunction(typ string) string {
	if typ == "" {
		return "function"
	}
	return typ
}

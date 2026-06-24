// Package ollama is a minimal, generic HTTP client for an Ollama server.
//
// It speaks the arclib/llm wire types so it can be dropped behind any
// MyClerk-Protocol LLM handler (e.g. mayaservices) without leaking Ollama
// specifics. Tool-calling is translated between the OpenAI-style string
// arguments used by arclib/llm and Ollama's object arguments.
package ollama

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

// DefaultBaseURL is the conventional local Ollama endpoint.
const DefaultBaseURL = "http://127.0.0.1:11434"

// Client talks to an Ollama server over HTTP.
type Client struct {
	baseURL string
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

// --- Ollama wire types -----------------------------------------------------

type wireMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []wireToolCall `json:"tool_calls,omitempty"`
}

type wireToolCall struct {
	Function wireFuncCall `json:"function"`
}

type wireFuncCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
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

type wireChatReq struct {
	Model    string         `json:"model"`
	Messages []wireMessage  `json:"messages"`
	Tools    []wireTool     `json:"tools,omitempty"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type wireChatResp struct {
	Model           string      `json:"model"`
	Message         wireMessage `json:"message"`
	Done            bool        `json:"done"`
	DoneReason      string      `json:"done_reason"`
	PromptEvalCount int         `json:"prompt_eval_count"`
	EvalCount       int         `json:"eval_count"`
	Error           string      `json:"error"`
}

// --- Chat ------------------------------------------------------------------

// Chat performs a non-streaming chat completion.
func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	wireReq := c.buildChatReq(req, false)
	body, err := c.post(ctx, "/api/chat", wireReq)
	if err != nil {
		return llm.ChatResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr wireChatResp
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.ChatResponse{}, fmt.Errorf("decode chat response: %w", err)
	}
	if wr.Error != "" {
		return llm.ChatResponse{}, fmt.Errorf("ollama: %s", wr.Error)
	}
	return llm.ChatResponse{
		Model:        wr.Model,
		Message:      toLLMMessage(wr.Message),
		FinishReason: finishReason(wr),
		Usage: llm.Usage{
			PromptTokens:     wr.PromptEvalCount,
			CompletionTokens: wr.EvalCount,
			TotalTokens:      wr.PromptEvalCount + wr.EvalCount,
		},
	}, nil
}

// ChatStream performs a streaming chat completion, invoking fn for each chunk.
// The final chunk has Done == true and carries Usage. If fn returns an error,
// streaming stops and that error is returned.
func (c *Client) ChatStream(ctx context.Context, req llm.ChatRequest, fn func(llm.StreamChunk) error) error {
	wireReq := c.buildChatReq(req, true)
	body, err := c.post(ctx, "/api/chat", wireReq)
	if err != nil {
		return err
	}
	defer func() { _ = body.Close() }()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var wr wireChatResp
		if err := json.Unmarshal(line, &wr); err != nil {
			return fmt.Errorf("decode stream chunk: %w", err)
		}
		if wr.Error != "" {
			return fmt.Errorf("ollama: %s", wr.Error)
		}
		chunk := llm.StreamChunk{Delta: toLLMMessage(wr.Message), Done: wr.Done}
		if wr.Done {
			chunk.FinishReason = finishReason(wr)
			chunk.Usage = &llm.Usage{
				PromptTokens:     wr.PromptEvalCount,
				CompletionTokens: wr.EvalCount,
				TotalTokens:      wr.PromptEvalCount + wr.EvalCount,
			}
		}
		if err := fn(chunk); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (c *Client) buildChatReq(req llm.ChatRequest, stream bool) wireChatReq {
	options := map[string]any{}
	for k, v := range req.Options {
		options[k] = v
	}
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if len(options) == 0 {
		options = nil
	}

	msgs := make([]wireMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = fromLLMMessage(m)
	}

	tools := make([]wireTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, wireTool{
			Type: t.Type,
			Function: wireToolFunc{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		})
	}

	return wireChatReq{Model: req.Model, Messages: msgs, Tools: tools, Stream: stream, Options: options}
}

// --- Models ----------------------------------------------------------------

// List returns the models installed on the server (Ollama /api/tags).
func (c *Client) List(ctx context.Context) (llm.ListModelsResponse, error) {
	body, err := c.get(ctx, "/api/tags")
	if err != nil {
		return llm.ListModelsResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr struct {
		Models []struct {
			Name    string `json:"name"`
			Details struct {
				Family            string `json:"family"`
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.ListModelsResponse{}, fmt.Errorf("decode tags: %w", err)
	}
	out := llm.ListModelsResponse{Models: make([]llm.ModelInfo, 0, len(wr.Models))}
	for _, m := range wr.Models {
		out.Models = append(out.Models, llm.ModelInfo{
			Name:          m.Name,
			Family:        m.Details.Family,
			ParameterSize: m.Details.ParameterSize,
			Quantization:  m.Details.QuantizationLevel,
		})
	}
	return out, nil
}

// Show returns details and capabilities for a single model (Ollama /api/show).
func (c *Client) Show(ctx context.Context, model string) (llm.ModelInfo, error) {
	body, err := c.post(ctx, "/api/show", map[string]any{"model": model})
	if err != nil {
		return llm.ModelInfo{}, err
	}
	defer func() { _ = body.Close() }()

	var wr struct {
		Capabilities []string `json:"capabilities"`
		Details      struct {
			Family            string `json:"family"`
			ParameterSize     string `json:"parameter_size"`
			QuantizationLevel string `json:"quantization_level"`
		} `json:"details"`
		ModelInfo map[string]any `json:"model_info"`
	}
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.ModelInfo{}, fmt.Errorf("decode show: %w", err)
	}
	info := llm.ModelInfo{
		Name:          model,
		Family:        wr.Details.Family,
		ParameterSize: wr.Details.ParameterSize,
		Quantization:  wr.Details.QuantizationLevel,
		Capabilities:  wr.Capabilities,
	}
	for k, v := range wr.ModelInfo {
		if strings.HasSuffix(k, ".context_length") {
			if f, ok := v.(float64); ok {
				info.ContextLength = int(f)
			}
		}
	}
	return info, nil
}

// --- Embeddings ------------------------------------------------------------

// Embed generates embeddings for the request inputs (Ollama /api/embed).
func (c *Client) Embed(ctx context.Context, req llm.EmbedRequest) (llm.EmbedResponse, error) {
	body, err := c.post(ctx, "/api/embed", map[string]any{"model": req.Model, "input": req.Input})
	if err != nil {
		return llm.EmbedResponse{}, err
	}
	defer func() { _ = body.Close() }()

	var wr struct {
		Model           string      `json:"model"`
		Embeddings      [][]float32 `json:"embeddings"`
		PromptEvalCount int         `json:"prompt_eval_count"`
		Error           string      `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&wr); err != nil {
		return llm.EmbedResponse{}, fmt.Errorf("decode embed: %w", err)
	}
	if wr.Error != "" {
		return llm.EmbedResponse{}, fmt.Errorf("ollama: %s", wr.Error)
	}
	dims := 0
	if len(wr.Embeddings) > 0 {
		dims = len(wr.Embeddings[0])
	}
	return llm.EmbedResponse{
		Model:      wr.Model,
		Embeddings: wr.Embeddings,
		Dimensions: dims,
		Usage:      llm.Usage{PromptTokens: wr.PromptEvalCount, TotalTokens: wr.PromptEvalCount},
	}, nil
}

// --- Health ----------------------------------------------------------------

// Health reports server reachability and currently loaded models.
func (c *Client) Health(ctx context.Context) (llm.HealthResponse, error) {
	body, err := c.get(ctx, "/api/version")
	if err != nil {
		return llm.HealthResponse{Healthy: false, Provider: "ollama", Detail: err.Error()}, nil //nolint:nilerr // health failures are surfaced via Healthy=false, not as a transport error
	}
	var ver struct {
		Version string `json:"version"`
	}
	_ = json.NewDecoder(body).Decode(&ver)
	_ = body.Close()

	h := llm.HealthResponse{Healthy: true, Provider: "ollama", Detail: ver.Version}

	// Loaded models are best-effort.
	if psBody, err := c.get(ctx, "/api/ps"); err == nil {
		var ps struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if json.NewDecoder(psBody).Decode(&ps) == nil {
			for _, m := range ps.Models {
				h.LoadedModels = append(h.LoadedModels, m.Name)
			}
		}
		_ = psBody.Close()
	}
	return h, nil
}

// --- helpers ---------------------------------------------------------------

func (c *Client) post(ctx context.Context, path string, payload any) (io.ReadCloser, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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
	// The request URL is built from the operator-configured base URL, not input.
	resp, err := c.hc.Do(req) //nolint:gosec // G704: base URL is operator-configured, not attacker input
	if err != nil {
		return nil, fmt.Errorf("ollama unreachable: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("ollama %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return resp.Body, nil
}

func finishReason(wr wireChatResp) string {
	if len(wr.Message.ToolCalls) > 0 {
		return llm.FinishToolCalls
	}
	if wr.DoneReason == "length" {
		return llm.FinishLength
	}
	return llm.FinishStop
}

func toLLMMessage(m wireMessage) llm.ChatMessage {
	out := llm.ChatMessage{Role: m.Role, Content: m.Content}
	for i, tc := range m.ToolCalls {
		args := "{}"
		if tc.Function.Arguments != nil {
			if b, err := json.Marshal(tc.Function.Arguments); err == nil {
				args = string(b)
			}
		}
		out.ToolCalls = append(out.ToolCalls, llm.ToolCall{
			ID:       fmt.Sprintf("call_%d", i),
			Type:     "function",
			Function: llm.FunctionCall{Name: tc.Function.Name, Arguments: args},
		})
	}
	return out
}

func fromLLMMessage(m llm.ChatMessage) wireMessage {
	out := wireMessage{Role: m.Role, Content: m.Content}
	for _, tc := range m.ToolCalls {
		var args map[string]any
		if tc.Function.Arguments != "" {
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		}
		out.ToolCalls = append(out.ToolCalls, wireToolCall{
			Function: wireFuncCall{Name: tc.Function.Name, Arguments: args},
		})
	}
	return out
}

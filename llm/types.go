package llm

// Role values for ChatMessage.Role.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Finish reason values for ChatResponse.FinishReason and StreamChunk.FinishReason.
const (
	FinishStop      = "stop"
	FinishLength    = "length"
	FinishToolCalls = "tool_calls"
)

// Capability values for ModelInfo.Capabilities.
const (
	CapCompletion = "completion"
	CapTools      = "tools"
	CapVision     = "vision"
	CapEmbed      = "embed"
)

// ChatMessage is a single message in a conversation.
type ChatMessage struct {
	Role string `msgpack:"role"`
	// Content is the answer text. For reasoning models it may be empty when the
	// model put its answer in ReasoningContent instead — callers that only need
	// the answer should fall back to ReasoningContent when Content is empty.
	Content string `msgpack:"content,omitempty"`
	// ReasoningContent is the model's "thinking" channel (response only). It is
	// forwarded so callers can show progress or recover an answer that landed
	// here. Empty for models that don't expose reasoning.
	ReasoningContent string     `msgpack:"reasoning,omitempty"`
	Name             string     `msgpack:"name,omitempty"`
	ToolCallID       string     `msgpack:"tcid,omitempty"`       // set when Role == RoleTool
	ToolCalls        []ToolCall `msgpack:"tool_calls,omitempty"` // set when Role == RoleAssistant
}

// ToolCall is a function call requested by the model.
type ToolCall struct {
	ID       string       `msgpack:"id"`
	Type     string       `msgpack:"type"` // always "function"
	Function FunctionCall `msgpack:"function"`
}

// FunctionCall carries the called function name and its JSON-encoded arguments.
type FunctionCall struct {
	Name      string `msgpack:"name"`
	Arguments string `msgpack:"args"` // JSON string, OpenAI-compatible
}

// Tool is a function the model may call.
type Tool struct {
	Type     string       `msgpack:"type"` // always "function"
	Function ToolFunction `msgpack:"function"`
}

// ToolFunction describes a callable function and its JSON-Schema parameters.
type ToolFunction struct {
	Name        string         `msgpack:"name"`
	Description string         `msgpack:"description,omitempty"`
	Parameters  map[string]any `msgpack:"parameters,omitempty"`
}

// Usage reports token accounting for a completion.
type Usage struct {
	PromptTokens     int `msgpack:"pt"`
	CompletionTokens int `msgpack:"ct"`
	TotalTokens      int `msgpack:"tt"`
}

// ChatRequest is the MAYA_QUERY (0x1000) payload.
//
// Model carries the resolved model name. The MyClerk core resolves an @role
// (e.g. @default) to a concrete model and sets it here; mayaservices is
// model-agnostic and forwards Model to Ollama verbatim. An empty Model lets the
// server fall back to its configured default (fallback only, never the norm).
type ChatRequest struct {
	Model       string         `msgpack:"model"`
	Messages    []ChatMessage  `msgpack:"messages"`
	Tools       []Tool         `msgpack:"tools,omitempty"`
	Temperature float32        `msgpack:"temp,omitempty"`
	MaxTokens   int            `msgpack:"max_tokens,omitempty"`
	Stream      bool           `msgpack:"stream,omitempty"` // true ⇒ reply via MAYA_STREAM
	Options     map[string]any `msgpack:"opts,omitempty"`   // passthrough Ollama options
}

// ChatResponse is the MAYA_RESPONSE (0x1001) payload.
type ChatResponse struct {
	Model        string      `msgpack:"model"`
	Message      ChatMessage `msgpack:"message"` // assistant message (may carry ToolCalls)
	FinishReason string      `msgpack:"finish"`
	Usage        Usage       `msgpack:"usage"`
	CreatedNs    int64       `msgpack:"created_ns,omitempty"`
}

// StreamChunk is one MAYA_STREAM (0x1002) frame. The server emits a sequence of
// chunks terminated by one with Done == true (which also carries final Usage).
type StreamChunk struct {
	Delta        ChatMessage `msgpack:"delta"`
	FinishReason string      `msgpack:"finish,omitempty"`
	Usage        *Usage      `msgpack:"usage,omitempty"`
	Done         bool        `msgpack:"done"`
}

// ModelInfo describes a model available on the backend.
type ModelInfo struct {
	Name          string   `msgpack:"name"`
	Family        string   `msgpack:"family,omitempty"`
	ParameterSize string   `msgpack:"psize,omitempty"`
	Quantization  string   `msgpack:"quant,omitempty"`
	ContextLength int      `msgpack:"ctx,omitempty"`
	Capabilities  []string `msgpack:"caps,omitempty"`
}

// HasCapability reports whether the model advertises the given capability.
func (m ModelInfo) HasCapability(capability string) bool {
	for _, c := range m.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// ListModelsResponse is the MAYA_MODELS_GET (0x1060) reply.
type ListModelsResponse struct {
	Models []ModelInfo `msgpack:"models"`
}

// HealthResponse is the MAYA_HEALTH_GET (0x1061) reply.
type HealthResponse struct {
	Healthy      bool     `msgpack:"healthy"`
	Provider     string   `msgpack:"provider"`
	Detail       string   `msgpack:"detail,omitempty"`
	LoadedModels []string `msgpack:"loaded,omitempty"`
}

// EmbedRequest is the MAYA_EMBED (0x1080) payload.
type EmbedRequest struct {
	Model string   `msgpack:"model"`
	Input []string `msgpack:"input"`
}

// EmbedResponse is the MAYA_EMBED (0x1080) reply.
type EmbedResponse struct {
	Model      string      `msgpack:"model"`
	Embeddings [][]float32 `msgpack:"emb"`
	Dimensions int         `msgpack:"dim"`
	Usage      Usage       `msgpack:"usage"`
}

// TTSRequest is the MAYA_T_T_S (0x1032) payload.
type TTSRequest struct {
	Text     string  `msgpack:"text"`
	Voice    string  `msgpack:"voice,omitempty"`
	Engine   string  `msgpack:"engine,omitempty"` // e.g. "xtts" (interchangeable)
	Language string  `msgpack:"lang,omitempty"`
	Format   string  `msgpack:"format,omitempty"` // "opus" | "wav"
	Speed    float32 `msgpack:"speed,omitempty"`
}

// TTSResponse is the MAYA_T_T_S (0x1032) reply.
type TTSResponse struct {
	Audio      []byte `msgpack:"audio"`
	Format     string `msgpack:"format"`
	SampleRate int    `msgpack:"rate"`
	DurationMs int    `msgpack:"dur_ms"`
}

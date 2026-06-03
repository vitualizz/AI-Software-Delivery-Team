// Package llm defines the Provider port and concrete adapters for LLM backends.
// The domain knows only the Provider interface; concrete adapters (Anthropic,
// OpenAI, Mock) live here as well.
package llm

// Message is a single entry in a conversation exchange.
type Message struct {
	// Role is the speaker: "user", "assistant", or "system".
	Role string
	// Content is the text content of the message.
	Content string
}

// Request is the input to a Provider completion call.
type Request struct {
	// Messages is the ordered conversation history.
	Messages []Message

	// MaxTokens is the maximum number of tokens to generate.
	// If zero, the provider uses its default.
	MaxTokens int

	// Temperature controls output randomness (0.0–1.0).
	// If zero, the provider uses its default.
	Temperature float64
}

// Chunk is a single streaming token delta from a Provider.
type Chunk struct {
	// Content is the incremental text.
	Content string

	// Done signals that the stream is complete.
	Done bool
}

// Response is the aggregated result of a Complete call.
type Response struct {
	// Content is the full generated text (all chunks joined).
	Content string
}

package llm

import "context"

// OpenAIProvider is the OpenAI adapter stub.
// Full implementation is post-MVP.
type OpenAIProvider struct{}

// NewOpenAI constructs an OpenAIProvider stub.
func NewOpenAI() *OpenAIProvider {
	return &OpenAIProvider{}
}

func (p *OpenAIProvider) Stream(_ context.Context, _ Request) (<-chan Chunk, error) {
	return nil, ErrNotImplemented
}

func (p *OpenAIProvider) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{}, ErrNotImplemented
}

func (p *OpenAIProvider) Name() string { return "openai" }

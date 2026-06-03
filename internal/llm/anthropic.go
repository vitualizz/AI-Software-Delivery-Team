package llm

import "context"

// AnthropicProvider is the Anthropic Claude adapter stub.
// Full implementation is post-MVP.
type AnthropicProvider struct{}

// NewAnthropic constructs an AnthropicProvider stub.
func NewAnthropic() *AnthropicProvider {
	return &AnthropicProvider{}
}

func (p *AnthropicProvider) Stream(_ context.Context, _ Request) (<-chan Chunk, error) {
	return nil, ErrNotImplemented
}

func (p *AnthropicProvider) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{}, ErrNotImplemented
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

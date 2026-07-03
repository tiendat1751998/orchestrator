package provider

import "context"

// Provider is the core interface that all AI providers must implement.
//
// Lifecycle:
//
//	Provider lifecycle (Init, Start, Stop) is managed by the Plugin interface
//	in contracts/plugin. This interface only defines runtime behavior.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Send sends a request and waits for the complete response.
	Send(ctx context.Context, req *Request) (*Response, error)

	// Stream sends a request and returns a channel for streaming the response.
	Stream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

	// IsAvailable checks if the provider is ready to accept requests.
	IsAvailable(ctx context.Context) bool

	// Models returns the list of models supported by this provider.
	Models(ctx context.Context) ([]string, error)
}

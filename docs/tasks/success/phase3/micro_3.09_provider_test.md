# Micro-Task 3.09: Create sdk/provider/provider_test.go

## Info
- **File**: `sdk/provider/provider_test.go`
- **Package**: `provider_test`
- **Depends on**: 3.07 (request_builder.md), 3.08 (stream_collector.md)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/provider/...`

## Purpose
Implements integration unit tests for the Provider SDK, verifying the immutability features of `RequestBuilder`, request validation parameters, stream collector integrations, chunk assemblies, and producer thread releases.

## EXACT code to create

```go
package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

// =============================================================================
// Request Builder Tests
// =============================================================================

func TestRequestBuilder_Immutability(t *testing.T) {
	b1 := sdkprovider.NewRequestBuilder("model-a").WithTemperature(0.7)
	b2 := b1.WithTemperature(0.2)

	r1, _ := b1.Build()
	r2, _ := b2.Build()

	if *r1.Temperature != 0.7 {
		t.Errorf("b1 temp mutated: got %f, want 0.7", *r1.Temperature)
	}
	if *r2.Temperature != 0.2 {
		t.Errorf("b2 temp: got %f, want 0.2", *r2.Temperature)
	}
}

func TestRequestBuilder_Validation(t *testing.T) {
	_, err := sdkprovider.NewRequestBuilder("model-x").Build()
	if err == nil {
		t.Error("expected build error for request with no messages")
	}

	_, err = sdkprovider.NewRequestBuilder("model-x").
		AddUserMessage("hello").
		WithTemperature(2.5).
		Build()
	if err == nil {
		t.Error("expected build error for temperature out of bounds")
	}
}

func TestRequestBuilder_Success(t *testing.T) {
	req, err := sdkprovider.NewRequestBuilder("gemini-2.5-pro").
		AddSystemMessage("system instruction").
		AddUserMessage("user question").
		WithMaxTokens(100).
		WithTemperature(0.0).
		WithStream(true).
		Build()

	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if req.Model != "gemini-2.5-pro" {
		t.Errorf("model: got %q", req.Model)
	}
	if len(req.Messages) != 2 {
		t.Errorf("messages count: got %d", len(req.Messages))
	}
	if req.Messages[0].Role != provider.RoleSystem || req.Messages[0].Content != "system instruction" {
		t.Errorf("first message: %v", req.Messages[0])
	}
	if !req.Stream {
		t.Error("expected stream to be true")
	}
	if *req.MaxTokens != 100 {
		t.Errorf("max tokens: got %d", *req.MaxTokens)
	}
	if *req.Temperature != 0.0 {
		t.Errorf("temperature: got %f", *req.Temperature)
	}
}

// =============================================================================
// Stream Collector Tests
// =============================================================================

func TestCollectStream_HappyPath(t *testing.T) {
	ch := make(chan provider.StreamChunk, 3)

	ch <- provider.StreamChunk{Delta: "Hello "}
	ch <- provider.StreamChunk{Delta: "world!"}
	ch <- provider.StreamChunk{Done: true}
	close(ch)

	resp, err := sdkprovider.CollectStream(context.Background(), ch)
	if err != nil {
		t.Fatalf("CollectStream: %v", err)
	}

	if resp.Content != "Hello world!" {
		t.Errorf("content: got %q, want %q", resp.Content, "Hello world!")
	}
}

func TestCollectStream_ToolCallAggregation(t *testing.T) {
	ch := make(chan provider.StreamChunk, 4)

	ch <- provider.StreamChunk{
		ToolCalls: []provider.ToolCall{
			{ID: "call-1", Name: "write_file", Args: []byte(`{"path"`)},
		},
	}
	ch <- provider.StreamChunk{
		ToolCalls: []provider.ToolCall{
			{Args: []byte(`:"main.go"`)},
		},
	}
	ch <- provider.StreamChunk{
		ToolCalls: []provider.ToolCall{
			{Args: []byte(`, "data": "hello"}`)},
		},
	}
	ch <- provider.StreamChunk{Done: true}
	close(ch)

	resp, err := sdkprovider.CollectStream(context.Background(), ch)
	if err != nil {
		t.Fatalf("CollectStream: %v", err)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}

	tc := resp.ToolCalls[0]
	if tc.ID != "call-1" || tc.Name != "write_file" {
		t.Errorf("tool metadata mismatch: ID=%q, Name=%q", tc.ID, tc.Name)
	}

	expectedArgs := `{"path":"main.go", "data": "hello"}`
	if string(tc.Args) != expectedArgs {
		t.Errorf("arguments aggregation: got %q, want %q", string(tc.Args), expectedArgs)
	}
}

func TestCollectStream_ContextCancellationDrain(t *testing.T) {
	ch := make(chan provider.StreamChunk, 5)

	ctx, cancel := context.WithCancel(context.Background())

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for i := 0; i < 5; i++ {
			select {
			case ch <- provider.StreamChunk{Delta: "data"}:
				time.Sleep(50 * time.Millisecond)
			case <-time.After(2 * time.Second):
				return
			}
		}
		close(ch)
	}()

	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()

	_, err := sdkprovider.CollectStream(ctx, ch)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}

	select {
	case <-writerDone:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout: producer writer goroutine was leaked (blocked)")
	}
}

func TestCollectStream_ErrorMidStream(t *testing.T) {
	ch := make(chan provider.StreamChunk, 2)
	ch <- provider.StreamChunk{Delta: "partial"}
	ch <- provider.StreamChunk{Error: errors.New("network disconnect")}
	close(ch)

	_, err := sdkprovider.CollectStream(context.Background(), ch)
	if err == nil {
		t.Fatal("expected error mid-stream, got nil")
	}
}
```

## Rules
1. **Parallel Tests Safety**: Track test status variables using local variables or safe select channel notifications rather than un-synchronized variables.
2. **Channel releases validations**: Verify that producer threads exit cleanly (meaning channels are drained) when contexts are cancelled mid-stream to catch resource leaks.
3. **Deep Copy Assertions**: Verify that clones allocated during request construction create fresh pointer variables rather than sharing underlying addresses.

## ⚠️ Pitfalls

### Pitfall 1: Relying on sleep intervals for thread cancellation tests
Using fixed timeouts to verify writer thread termination can lead to flaky tests on slow machines. Coordinate threads using completion channels (`writerDone`) to ensure the test is robust.

### Pitfall 2: Reusing builder requests across parallel tests without isolation
If test scenarios mutate builders concurrently, they will trigger race conditions. Ensure builders return fresh copied states.

## Verify
```bash
go test -v -race -count=1 ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/provider_test.go` exists
- [ ] Package: `provider_test` (external testing package)
- [ ] Immutability checks verify that builder values do not leak across clones
- [ ] Validators reject requests missing messages or with bad temperature bounds
- [ ] Text chunk aggregations match expected content
- [ ] Tool call arguments are concatenated correctly
- [ ] Context cancellation tests verify that producers exit cleanly
- [ ] Mid-stream errors return standard errors
- [ ] `go test -v -race ./sdk/provider/...` passes

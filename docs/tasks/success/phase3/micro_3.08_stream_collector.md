# Micro-Task 3.08: Create sdk/provider/stream.go

## Info
- **File**: `sdk/provider/stream.go`
- **Package**: `provider`
- **Depends on**: 3.06 (provider.go), 1.10 (provider response contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/provider/...`

## Purpose
Implements the stream collection utilities (`CollectStream`, `ForwardStream`, and `buildResponse` helpers) that gather text chunks and tool call segments streamed from AI models into completed responses, preventing thread leaks during context cancellations.

## EXACT code to create

```go
package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// CollectStream drains the StreamChunk channel, aggregates content delta,
// merges fragmented tool calls, and compiles the final unified provider.Response.
func CollectStream(ctx context.Context, ch <-chan provider.StreamChunk) (*provider.Response, error) {
	if ch == nil {
		return nil, errors.New("sdk/provider: nil stream channel")
	}

	var contentBuilder strings.Builder
	var lastUsage provider.Usage
	var lastModel string
	var lastID string
	var finishReason string

	type accumulatedToolCall struct {
		id   string
		name string
		args strings.Builder
	}
	accumulated := make(map[int]*accumulatedToolCall)

	for {
		select {
		case <-ctx.Done():
			// Drain channel in background to prevent writer goroutines from hanging
			go func() {
				for range ch {
				}
			}()
			return nil, ctx.Err()

		case chunk, ok := <-ch:
			if !ok {
				return buildResponse(lastID, contentBuilder.String(), accumulated, finishReason, lastUsage, lastModel), nil
			}

			if chunk.Error != nil {
				return nil, fmt.Errorf("sdk/provider: error occurred during streaming: %w", chunk.Error)
			}

			if chunk.Delta != "" {
				contentBuilder.WriteString(chunk.Delta)
			}

			for i, tc := range chunk.ToolCalls {
				acc, exists := accumulated[i]
				if !exists {
					acc = &accumulatedToolCall{}
					accumulated[i] = acc
				}
				if tc.ID != "" {
					acc.id = tc.ID
				}
				if tc.Name != "" {
					acc.name = tc.Name
				}
				if len(tc.Args) > 0 {
					acc.args.Write(tc.Args)
				}
			}

			if chunk.FinishReason != "" {
				finishReason = chunk.FinishReason
			}
			if chunk.Usage != nil {
				lastUsage.Add(*chunk.Usage)
			}

			if chunk.Done {
				return buildResponse(lastID, contentBuilder.String(), accumulated, finishReason, lastUsage, lastModel), nil
			}
		}
	}
}

// ForwardStream processes chunks as they arrive and triggers the provided callback.
func ForwardStream(ctx context.Context, ch <-chan provider.StreamChunk, callback func(provider.StreamChunk)) error {
	if ch == nil {
		return errors.New("sdk/provider: nil stream channel")
	}
	if callback == nil {
		return errors.New("sdk/provider: nil callback function")
	}

	for {
		select {
		case <-ctx.Done():
			go func() {
				for range ch {
				}
			}()
			return ctx.Err()

		case chunk, ok := <-ch:
			if !ok {
				return nil
			}
			if chunk.Error != nil {
				return fmt.Errorf("sdk/provider: stream error: %w", chunk.Error)
			}
			callback(chunk)
			if chunk.Done {
				return nil
			}
		}
	}
}

// buildResponse maps accumulated chunks data back to the final Response struct.
func buildResponse(
	id string,
	content string,
	accumulated map[int]*accumulatedToolCall,
	finishReason string,
	usage provider.Usage,
	model string,
) *provider.Response {
	toolCalls := make([]provider.ToolCall, 0, len(accumulated))
	for i := 0; i < len(accumulated); i++ {
		acc, ok := accumulated[i]
		if ok && (acc.id != "" || acc.name != "") {
			toolCalls = append(toolCalls, provider.ToolCall{
				ID:   acc.id,
				Name: acc.name,
				Args: []byte(acc.args.String()),
			})
		}
	}

	if id == "" {
		id = fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	}

	return &provider.Response{
		ID:           id,
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: finishReason,
		Usage:        usage,
		Model:        model,
		CreatedAt:    time.Now(),
	}
}
```

## Rules
1. **Background Channel Draining**: If a context is cancelled before streaming completes, spin off a background routine to drain the channel. This prevents writing goroutines from hanging on unbuffered channel writes.
2. **Fragmented Tool Call Buffering**: Aggregate tool call arguments by slice index and buffer them in `strings.Builder` writers rather than processing partial fragments immediately.
3. **Bubble Mid-Stream Errors**: If `chunk.Error` is non-nil, immediately terminate processing and bubble up the error.

## Pitfalls

### Pitfall 1: Leaking provider goroutines on context cancellation
Always drain the channel in a separate background goroutine to allow the provider's writer thread to exit cleanly.

### Pitfall 2: Fragmented Tool Call Slices Mismatches
Because the `ToolCall` contract does not contain a raw stream index parameter, the provider adapter must pad the `chunk.ToolCalls` slice so that the slice index `i` matches the global index of the active tool call. If the adapter simply appends active deltas to un-padded slices, `i` will shift dynamically across chunks, causing argument strings to mismatch or corrupt.

## Verify
```bash
go build ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/stream.go` exists
- [ ] Package: `provider`
- [ ] `CollectStream` aggregates text chunk deltas
- [ ] Tool call arguments are buffered by index key
- [ ] Background goroutines drain channels on context cancellations
- [ ] Mid-stream errors terminate execution loops immediately
- [ ] `buildResponse` builds complete `provider.Response` structs
- [ ] `go build ./sdk/provider/...` passes

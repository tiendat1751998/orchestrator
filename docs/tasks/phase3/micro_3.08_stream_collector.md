# Micro-Task 3.08: Create sdk/provider/stream.go

## Info
- **File**: `sdk/provider/stream.go`
- **Package**: `provider`
- **Depends on**: 3.06 (provider.go), 1.10 (provider response contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/provider/...`

## Purpose
Triển khai bộ công cụ xử lý luồng dữ liệu (`CollectStream` và `ForwardStream`). Các hàm này gộp các phần nội dung nhỏ (StreamChunks) nhận được từ mô hình AI thành một đối tượng `Response` duy nhất, thực hiện gộp cấu trúc Tool Calls phân mảnh (delta arguments), đồng thời đảm bảo giải phóng và thu gom (drain) kênh dữ liệu khi có sự kiện hủy Context để tránh rò rỉ Goroutines (goroutine leak).

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
//
// Goroutine Leak Protection:
//   - If the context is cancelled before the stream completes, it immediately
//     spawns a background goroutine to drain the remaining chunks and exit cleanly.
//     This ensures the writer goroutine is never blocked writing to a dead reader.
func CollectStream(ctx context.Context, ch <-chan provider.StreamChunk) (*provider.Response, error) {
	if ch == nil {
		return nil, errors.New("sdk/provider: nil stream channel")
	}

	var contentBuilder strings.Builder
	var lastUsage provider.Usage
	var lastModel string
	var lastID string
	var finishReason string

	// Accumulated tool calls tracked by index in the stream
	type accumulatedToolCall struct {
		id   string
		name string
		args strings.Builder
	}
	accumulated := make(map[int]*accumulatedToolCall)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled!
			// Prevent producer leak: drain the channel asynchronously
			go func() {
				for range ch {
					// Discard chunks silently to allow producer to exit
				}
			}()
			return nil, ctx.Err()

		case chunk, ok := <-ch:
			if !ok {
				// Channel closed by producer. Stream ended.
				return buildResponse(lastID, contentBuilder.String(), accumulated, finishReason, lastUsage, lastModel), nil
			}

			// Check if chunk contains mid-stream errors
			if chunk.Error != nil {
				return nil, fmt.Errorf("sdk/provider: error occurred during streaming: %w", chunk.Error)
			}

			// Aggregate standard output delta
			if chunk.Delta != "" {
				contentBuilder.WriteString(chunk.Delta)
			}

			// Aggregate fragmented tool calls
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

			// Save final chunk metadata
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
// Spawns a background drain routine if the context is cancelled mid-stream.
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
			// Drain in background to prevent resource leaks
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

## ⚠️ Pitfalls

### Pitfall 1: Failing to drain the channel on context cancel (Producer Leak)
```go
// ❌ WRONG:
case <-ctx.Done():
    return nil, ctx.Err() // Exits immediately, leaving the producer goroutine blocked writing to the channel forever.

// ✅ CORRECT:
case <-ctx.Done():
    go func() {
        for range ch {} // Drain channel in background so producer writer terminates cleanly.
    }()
    return nil, ctx.Err()
}
```
If the reader halts but the channel is unbuffered, the writer goroutine blocks indefinitely, leaking memory and resources in production.

### Pitfall 2: Recreating ToolCall slice on each chunk instead of appending deltas
LLM providers stream tool call arguments as fragment strings (e.g. `{"pa`, `th":`, `"main.go"}`). You must parse tool calls by index `i` and write bytes directly into a buffer (`strings.Builder` or `bytes.Buffer`) rather than treating each chunk as a complete, separate tool call.

## Verify
```bash
go build ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/stream.go` exists
- [ ] Package: `provider`
- [ ] `CollectStream` resolves fragmented tool call JSON arguments correctly using index keys
- [ ] Spawns background goroutine to drain the channel if context cancels in `CollectStream`
- [ ] Spawns background goroutine to drain the channel if context cancels in `ForwardStream`
- [ ] Mid-stream errors (`chunk.Error != nil`) terminate loop and bubble up as standard Go errors
- [ ] `buildResponse` correctly constructs standard `provider.Response` struct
- [ ] `go build ./sdk/provider/...` passes

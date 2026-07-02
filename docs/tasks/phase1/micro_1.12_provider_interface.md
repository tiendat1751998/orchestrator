# Micro-Task 1.12: Tạo contracts/provider/provider.go

## Thông tin
- **File tạo**: `contracts/provider/provider.go`
- **Package**: `provider`
- **Dependencies trước**: 1.09, 1.10, 1.11
- **Thời gian**: 10 phút

## Nội dung CHÍNH XÁC cần tạo

```go
package provider

import "context"

// Provider is the core interface that all AI providers must implement.
//
// Every AI service (Antigravity CLI, Gemini API, Claude API, Ollama, etc.)
// must have an implementation of this interface. The orchestrator uses
// this interface exclusively — it never knows or cares about the specific
// provider behind it.
//
// Implementation guidelines:
//   - All methods must be safe for concurrent use (thread-safe).
//   - Send() and Stream() must respect context cancellation.
//   - Stream() must close the returned channel when done or on error.
//   - IsAvailable() must NOT block for more than 5 seconds.
//
// Lifecycle:
//   Provider lifecycle (Init, Start, Stop) is managed by the Plugin interface
//   in contracts/plugin. This interface only defines runtime behavior.
//   DO NOT add Init/Start/Stop here — that's Plugin's responsibility.
type Provider interface {
	// Name returns the unique identifier for this provider.
	// Example: "antigravity", "gemini-api", "claude-api", "ollama"
	//
	// Rules:
	//   - Must be lowercase, alphanumeric + hyphens only
	//   - Must be unique across all registered providers
	//   - Must not change after initialization
	Name() string

	// Send sends a request and waits for the complete response.
	//
	// This is a blocking call — it returns only after the AI finishes generating
	// the entire response.
	//
	// Parameters:
	//   - ctx: Must be respected for cancellation and timeout.
	//          Use context.WithTimeout() to enforce time limits.
	//   - req: The request to send. Must pass req.Validate() before calling.
	//
	// Returns:
	//   - *Response: The complete response from the AI.
	//   - error: System-level errors only:
	//       - contracts.ErrProviderTimeout — provider didn't respond in time
	//       - contracts.ErrProviderUnavailable — provider is down
	//       - contracts.ErrProviderRateLimited — hit rate limit
	//       - contracts.ErrProviderAuthFailed — invalid credentials
	//       - context.Canceled — caller cancelled the request
	//       - context.DeadlineExceeded — context timeout expired
	//
	// Note: Business-level "errors" (e.g., AI says "I can't do that") are NOT
	// errors — they are valid responses with Content explaining the refusal.
	Send(ctx context.Context, req *Request) (*Response, error)

	// Stream sends a request and returns a channel for streaming the response.
	//
	// The channel emits StreamChunk values as the AI generates content.
	// The channel is closed by the producer when:
	//   1. The AI finishes generating (last chunk has Done=true)
	//   2. An error occurs (chunk has Error set)
	//   3. The context is cancelled
	//
	// Parameters:
	//   - ctx: Must be respected. When ctx is cancelled, the producer must
	//          close the channel promptly (within 1 second).
	//   - req: The request to send.
	//
	// Returns:
	//   - <-chan StreamChunk: Read-only channel of response chunks.
	//     WHY read-only? → Consumer should not write to the channel.
	//     This is enforced by Go's type system.
	//   - error: Returned immediately if the stream cannot be established
	//     (e.g., provider is down, invalid credentials).
	//     Errors that occur DURING streaming are sent via StreamChunk.Error.
	//
	// Consumer MUST drain the channel:
	//   stream, err := provider.Stream(ctx, req)
	//   for chunk := range stream { ... } // range drains automatically
	//
	// If consumer stops reading before Done, the producer goroutine may leak.
	// Always use context cancellation to signal the producer to stop:
	//   ctx, cancel := context.WithCancel(parentCtx)
	//   defer cancel() // This signals the producer to stop and close channel
	Stream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

	// IsAvailable checks if the provider is ready to accept requests.
	//
	// For CLI providers: checks if the binary exists and is executable.
	// For API providers: performs a lightweight health check (e.g., list models).
	// For local providers: checks if the model server is running.
	//
	// This method must NOT block for more than 5 seconds.
	// Use the provided context for cancellation.
	//
	// Returns false if the provider is misconfigured, down, or unreachable.
	IsAvailable(ctx context.Context) bool

	// Models returns the list of models supported by this provider.
	//
	// Example return: ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"]
	//
	// This is used by:
	//   - CLI: `orchestrator providers list` command
	//   - Planner: to select the best model for a task
	//   - Validator: to check if the configured model exists
	Models(ctx context.Context) ([]string, error)
}
```

## Quy tắc
1. Interface CHỈ có 4 methods — minimal nhưng đủ cho mọi use case
2. KHÔNG có `Init()`, `Start()`, `Stop()` — lifecycle nằm trong Plugin interface
3. KHÔNG có `Close()` — resource cleanup nằm trong Plugin.Stop()
4. `Stream()` trả về `<-chan` (read-only) — type safety, consumer không thể ghi
5. Mỗi method có extensive Godoc — AI sẽ đọc comments này để implement

## ⚠️ Pitfalls
1. **Mixing concerns**: Provider interface = runtime behavior ONLY. Lifecycle (init, stop) = Plugin interface. Configuration = Config struct. TÁCH RA.
2. **Stream channel ownership**: Producer (provider) tạo channel, producer close channel. Consumer (orchestrator) chỉ ĐỌC. Nếu consumer close → producer write to closed channel → PANIC.
3. **Context propagation**: Provider PHẢI listen `ctx.Done()` trong Send() và Stream(). Nếu không → user nhấn Ctrl+C nhưng provider vẫn chạy.
4. **Error categorization**: Network/system errors → return error. AI refusal → return Response with Content. KHÔNG trộn lẫn.

## Checklist
- [ ] File `contracts/provider/provider.go` tồn tại
- [ ] Provider interface với đúng 4 methods
- [ ] `Send()` nhận `context.Context` là param đầu tiên
- [ ] `Stream()` trả về `<-chan StreamChunk` (read-only)
- [ ] KHÔNG có Init/Start/Stop methods
- [ ] KHÔNG có Close method
- [ ] Godoc comments chi tiết cho mỗi method
- [ ] Error documentation trong Send() comment
- [ ] `go build ./contracts/...` không lỗi

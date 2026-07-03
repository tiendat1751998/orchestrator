package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

type accumulatedToolCall struct {
	id   string
	name string
	args strings.Builder
}

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

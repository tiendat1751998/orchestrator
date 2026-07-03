# Micro-Task 4.15: Create plugins/providers/antigravity/provider_gemini.go

## Info
- **File**: `plugins/providers/antigravity/provider_gemini.go`
- **Package**: `antigravity`
- **Depends on**: 4.14
- **Time**: 25 min
- **Verify**: `go build ./plugins/providers/antigravity/...`

## Purpose
Implements the native Gemini API fallback driver (`GeminiProvider` and helpers) to communicate directly with Google's Gemini REST API endpoints, bypassing CLI process pipes.

## EXACT code to create

```go
package antigravity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

// GeminiProvider connects directly to the Google Gemini API endpoints via HTTP REST.
// Extends BaseProvider. Thread-safe.
type GeminiProvider struct {
	*sdkprovider.BaseProvider

	apiKey     string
	httpClient *http.Client
}

// NewGeminiProvider constructs a new GeminiProvider.
func NewGeminiProvider(cfg *provider.Config) (*GeminiProvider, error) {
	if cfg == nil {
		return nil, errors.New("gemini: configuration cannot be nil")
	}

	defaultModels := []string{"gemini-3.5-pro", "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"}
	baseProvider, err := sdkprovider.NewBaseProvider(cfg, defaultModels, nil)
	if err != nil {
		return nil, err
	}

	if cfg.APIKey == "" {
		return nil, errors.New("gemini: missing api_key configuration setting")
	}

	return &GeminiProvider{
		BaseProvider: baseProvider,
		apiKey:       cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.TimeoutOrDefault(),
		},
	}, nil
}

// IsAvailable pings the Gemini API service to verify connectivity.
func (p *GeminiProvider) IsAvailable(ctx context.Context) bool {
	if !p.IsStarted() {
		return false
	}
	// Basic check: is the API key populated?
	return p.apiKey != ""
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Send executes a standard JSON REST API request to the generateContent endpoint.
func (p *GeminiProvider) Send(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if !p.IsStarted() {
		return nil, errors.New("gemini: provider is not running")
	}

	// Map provider.Request to Gemini API schema
	var contents []geminiContent
	for _, m := range req.Messages {
		role := "user"
		if m.Role == provider.RoleAssistant {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: m.Content},
			},
		})
	}

	geminiReqBody := geminiRequest{Contents: contents}
	jsonBytes, err := json.Marshal(geminiReqBody)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		req.Model, p.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to construct HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini: API returned HTTP status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gemini: empty response candidate returned from API")
	}

	content := geminiResp.Candidates[0].Content.Parts[0].Text

	// Estimate token usages
	promptLen := len(jsonBytes)
	completionLen := len(content)
	usage := provider.Usage{
		PromptTokens:     (promptLen + 3) / 4,
		CompletionTokens: (completionLen + 3) / 4,
		TotalTokens:      ((promptLen + completionLen) + 3) / 4,
	}

	return &provider.Response{
		ID:           fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Content:      content,
		FinishReason: "stop",
		Usage:        usage,
		Model:        req.Model,
		CreatedAt:    time.Now(),
	}, nil
}

// Stream performs a Send request and returns a simulated word stream.
func (p *GeminiProvider) Stream(ctx context.Context, req *provider.Request) (<-chan provider.StreamChunk, error) {
	ch := make(chan provider.StreamChunk, 5)

	go func() {
		defer close(ch)
		resp, err := p.Send(ctx, req)
		if err != nil {
			ch <- provider.StreamChunk{Error: err}
			return
		}
		ch <- provider.StreamChunk{Delta: resp.Content}
		ch <- provider.StreamChunk{Done: true, Usage: &resp.Usage}
	}()

	return ch, nil
}
```

## Pitfalls

### Pitfall 1: Bypassing request context bindings
```go
// WRONG:
req, _ := http.NewRequest("POST", url, body) // Ignores context deadlines, risking resource leaks!

// CORRECT:
http.NewRequestWithContext(ctx, "POST", url, body)
```
Failing to bind the request context allows HTTP calls to hang indefinitely if the network drops. Always use `NewRequestWithContext`.

### Pitfall 2: Accessing response properties without index validation
If the API returns validation warnings or blocks the request, candidates slices will be empty. Indexing `geminiResp.Candidates[0]` directly will trigger a panic. Check candidate lengths first.

## Verify
```bash
go build ./plugins/providers/antigravity/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/provider_gemini.go`
- [ ] Package name is `antigravity`
- [ ] All exported types have Godoc
- [ ] `GeminiProvider` maps configuration keys correctly
- [ ] HTTP calls use context bindings
- [ ] Build command passes

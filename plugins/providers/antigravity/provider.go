package antigravity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/parser"
	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/prompt"
	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/session"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

// AntigravityProvider wraps process pooling and parsing engines.
type AntigravityProvider struct {
	*sdkprovider.BaseProvider

	logger *slog.Logger
	binary string
	sm     *session.SessionManager
	mu     sync.Mutex
}

// NewProvider creates a new AntigravityProvider.
func NewProvider(cfg *provider.Config) (*AntigravityProvider, error) {
	if cfg == nil {
		return nil, errors.New("antigravity: config cannot be nil")
	}

	defaultModels := []string{"gemini-3.5-pro", "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"}
	baseProvider, err := sdkprovider.NewBaseProvider(cfg, defaultModels, nil)
	if err != nil {
		return nil, err
	}

	binaryPath := cfg.Binary
	if binaryPath == "" {
		binaryPath = "antigravity"
	}

	// 5 maximum concurrent CLI processes, 5 minutes idle timeout
	sm := session.NewSessionManager(binaryPath, 5, 5*time.Minute)

	return &AntigravityProvider{
		BaseProvider: baseProvider,
		binary:       binaryPath,
		sm:           sm,
	}, nil
}

// IsAvailable verifies if the CLI binary exists and is executable.
func (p *AntigravityProvider) IsAvailable(ctx context.Context) bool {
	if !p.IsStarted() {
		return false
	}
	_, err := exec.LookPath(p.binary)
	return err == nil
}

// Send formats the request, writes to the CLI process, and parses the complete response.
func (p *AntigravityProvider) Send(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if !p.IsStarted() {
		return nil, errors.New("antigravity: provider is not running")
	}
	if req == nil {
		return nil, errors.New("antigravity: request cannot be nil")
	}

	sessionID, _ := ctx.Value("session_id").(string)
	if sessionID == "" {
		sessionID = "default-session"
	}

	s, err := p.sm.GetOrCreate(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("antigravity: failed to get session: %w", err)
	}

	formatted, err := prompt.BuildCLIPrompt(req)
	if err != nil {
		return nil, fmt.Errorf("antigravity: failed to build prompt: %w", err)
	}

	if err := s.Adapter.WritePrompt(formatted); err != nil {
		return nil, fmt.Errorf("antigravity: failed to write prompt: %w", err)
	}

	delimiter := "---END-OF-PROMPT---"
	rawResp, err := s.Adapter.ReadResponse(ctx, delimiter)
	if err != nil {
		return nil, parser.ParseError(err.Error())
	}

	toolCalls, _ := parser.ParseToolCalls(rawResp)
	cleanedContent, _ := parser.ParseMarkdown(rawResp)

	promptLen := len(formatted)
	completionLen := len(rawResp)
	usage := provider.Usage{
		PromptTokens:     (promptLen + 3) / 4,
		CompletionTokens: (completionLen + 3) / 4,
		TotalTokens:      ((promptLen + completionLen) + 3) / 4,
	}

	return &provider.Response{
		ID:           fmt.Sprintf("antigravity-%d", time.Now().UnixNano()),
		Content:      cleanedContent,
		ToolCalls:    toolCalls,
		FinishReason: "stop",
		Usage:        usage,
		Model:        req.Model,
		CreatedAt:    time.Now(),
	}, nil
}

// Stream simulates response streaming by reading lines from stdout and piping chunks.
func (p *AntigravityProvider) Stream(ctx context.Context, req *provider.Request) (<-chan provider.StreamChunk, error) {
	if !p.IsStarted() {
		return nil, errors.New("antigravity: provider is not running")
	}

	ch := make(chan provider.StreamChunk, 5)

	go func() {
		defer close(ch)

		resp, err := p.Send(ctx, req)
		if err != nil {
			select {
			case <-ctx.Done():
			case ch <- provider.StreamChunk{Error: err}:
			}
			return
		}

		// ponytail: simulated response streaming by splitting full response content into words, real streaming should read chunks progressively from terminal stream
		words := strings.Fields(resp.Content)
		for i, word := range words {
			delta := word
			if i > 0 {
				delta = " " + word
			}

			select {
			case <-ctx.Done():
				return
			case ch <- provider.StreamChunk{Delta: delta}:
			}

			timer := time.NewTimer(10 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}

		select {
		case <-ctx.Done():
		case ch <- provider.StreamChunk{
			Done:         true,
			FinishReason: "stop",
			Usage:        &resp.Usage,
		}:
		}
	}()

	return ch, nil
}

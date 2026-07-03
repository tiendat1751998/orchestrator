package agent

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

func TestBuildPrompt_SystemMessage(t *testing.T) {
	// Case 1: System prompt is set
	manifest := &agent.Manifest{
		SystemPrompt: "You are a helpful coding assistant.",
	}
	task := &agent.Task{
		ID:          "tsk-12345",
		Name:        "Test Task",
		Type:        "test",
		Description: "A simple task description",
	}

	messages := BuildPrompt(manifest, task, nil)
	if len(messages) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(messages))
	}

	systemMsg := messages[0]
	if systemMsg.Role != provider.RoleSystem {
		t.Errorf("expected first message to have role system, got %s", systemMsg.Role)
	}
	if systemMsg.Content != "You are a helpful coding assistant." {
		t.Errorf("expected system prompt content %q, got %q", "You are a helpful coding assistant.", systemMsg.Content)
	}

	// Case 2: System prompt is empty
	manifestEmpty := &agent.Manifest{
		SystemPrompt: "",
	}
	messagesEmpty := BuildPrompt(manifestEmpty, task, nil)
	if len(messagesEmpty) != 1 {
		t.Fatalf("expected exactly 1 message (task header user message only), got %d", len(messagesEmpty))
	}
	if messagesEmpty[0].Role != provider.RoleUser {
		t.Errorf("expected role user, got %s", messagesEmpty[0].Role)
	}
}

func TestBuildPrompt_TaskDetails(t *testing.T) {
	task := &agent.Task{
		ID:          contracts.TaskID("tsk-abcde"),
		Name:        "Build Prompt",
		Type:        "codegen",
		Description: "Write prompt.go file.",
		Input: map[string]any{
			"param1": "value1",
			"param2": float64(42),
		},
	}
	manifest := &agent.Manifest{}

	messages := BuildPrompt(manifest, task, nil)
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	userMsg := messages[0]
	if userMsg.Role != provider.RoleUser {
		t.Errorf("expected role user, got %s", userMsg.Role)
	}

	content := userMsg.Content

	// Check headers
	if !strings.Contains(content, "Task ID: tsk-abcde") {
		t.Errorf("expected task ID to be formatted, got: %s", content)
	}
	if !strings.Contains(content, "Task Name: Build Prompt") {
		t.Errorf("expected task name to be formatted, got: %s", content)
	}
	if !strings.Contains(content, "Task Type: codegen") {
		t.Errorf("expected task type to be formatted, got: %s", content)
	}
	if !strings.Contains(content, "Description:\nWrite prompt.go file.") {
		t.Errorf("expected description to be formatted, got: %s", content)
	}

	// Check pretty-printed JSON parameters
	if !strings.Contains(content, "Input Parameters:") {
		t.Errorf("expected 'Input Parameters:' header, got: %s", content)
	}
	if !strings.Contains(content, "```json") {
		t.Errorf("expected JSON block starting with ```json, got: %s", content)
	}

	// Unmarshal the JSON portion to ensure it is valid JSON representing the input map
	jsonStart := strings.Index(content, "```json\n")
	if jsonStart == -1 {
		t.Fatalf("could not find JSON start marker")
	}
	jsonEnd := strings.Index(content[jsonStart+8:], "\n```")
	if jsonEnd == -1 {
		t.Fatalf("could not find JSON end marker")
	}
	jsonStr := content[jsonStart+8 : jsonStart+8+jsonEnd]

	var parsedInput map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsedInput); err != nil {
		t.Fatalf("failed to parse formatted JSON inputs: %v", err)
	}

	if parsedInput["param1"] != "value1" || parsedInput["param2"] != float64(42) {
		t.Errorf("unexpected values in formatted JSON inputs: %v", parsedInput)
	}
}

func TestBuildPrompt_ContextItems(t *testing.T) {
	task := &agent.Task{
		ID:          "tsk-999",
		Name:        "Test Context Items",
		Type:        "testing",
		Description: "Verify context formatting",
	}
	task.AddContext("file", "content of file.go", "src/file.go")
	task.AddContext("empty", "", "ignored.txt") // should be skipped
	task.AddContext("history", "previous CLI output", "cli")

	manifest := &agent.Manifest{}

	messages := BuildPrompt(manifest, task, nil)
	// Expected messages: Task Details UserMsg + Context Item 1 UserMsg + Context Item 2 UserMsg = 3 messages
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	// Check first context item (index 1)
	ctxMsg1 := messages[1]
	if ctxMsg1.Role != provider.RoleUser {
		t.Errorf("expected role user, got %s", ctxMsg1.Role)
	}
	expectedCtx1 := "### Context Item (file) from \"src/file.go\":\ncontent of file.go\n---"
	if ctxMsg1.Content != expectedCtx1 {
		t.Errorf("expected context content:\n%q\ngot:\n%q", expectedCtx1, ctxMsg1.Content)
	}

	// Check second context item (index 2)
	ctxMsg2 := messages[2]
	if ctxMsg2.Role != provider.RoleUser {
		t.Errorf("expected role user, got %s", ctxMsg2.Role)
	}
	expectedCtx2 := "### Context Item (history) from \"cli\":\nprevious CLI output\n---"
	if ctxMsg2.Content != expectedCtx2 {
		t.Errorf("expected context content:\n%q\ngot:\n%q", expectedCtx2, ctxMsg2.Content)
	}
}

func TestBuildPrompt_TokenEstimationWarning(t *testing.T) {
	// A long prompt that will exceed limit
	sysPrompt := strings.Repeat("system", 50) // 300 chars
	description := strings.Repeat("desc", 50) // 200 chars
	ctxContent := strings.Repeat("ctx", 50)   // 150 chars

	task := &agent.Task{
		ID:          "tsk-limit-check",
		Name:        "Limit Checker",
		Type:        "check",
		Description: description,
	}
	task.AddContext("file", ctxContent, "file.txt")

	// Total characters estimation:
	// manifest.SystemPrompt (300) + taskDetails (len("You have been assigned...") + len(description) (200)) + len(context prompt) (~200)
	// estimatedTokens will be at least (300+250+200)/4 ≈ 187 tokens.

	manifestExceeds := &agent.Manifest{
		SystemPrompt: sysPrompt,
		MaxTokens:    100, // less than estimated tokens
	}

	manifestOk := &agent.Manifest{
		SystemPrompt: sysPrompt,
		MaxTokens:    1000, // greater than estimated tokens
	}

	manifestZero := &agent.Manifest{
		SystemPrompt: sysPrompt,
		MaxTokens:    0, // ignored
	}

	// Capture slog outputs
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Case 1: Exceeds limit -> should log warning
	buf.Reset()
	_ = BuildPrompt(manifestExceeds, task, logger)
	logOutput := buf.String()

	if !strings.Contains(logOutput, "prompt size estimation exceeds agent max_tokens limit") {
		t.Errorf("expected warning log on token limit overflow, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"task_id":"tsk-limit-check"`) {
		t.Errorf("expected task_id in warning log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"max_tokens_limit":100`) {
		t.Errorf("expected max_tokens_limit in warning log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"estimated_tokens"`) {
		t.Errorf("expected estimated_tokens in warning log, got: %s", logOutput)
	}

	// Case 2: Under limit -> should not log warning
	buf.Reset()
	_ = BuildPrompt(manifestOk, task, logger)
	logOutput = buf.String()
	if strings.Contains(logOutput, "prompt size estimation exceeds agent max_tokens limit") {
		t.Errorf("unexpected warning log when under token limit: %s", logOutput)
	}

	// Case 3: Zero limit -> should not log warning
	buf.Reset()
	_ = BuildPrompt(manifestZero, task, logger)
	logOutput = buf.String()
	if strings.Contains(logOutput, "prompt size estimation exceeds agent max_tokens limit") {
		t.Errorf("unexpected warning log when max_tokens is zero: %s", logOutput)
	}
}

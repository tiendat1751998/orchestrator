package agent

import (
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
)

func TestResult(t *testing.T) {
	taskID := contracts.TaskID("tsk-12345678")
	agentName := "test-agent"
	output := "execution output"
	errMsg := "some error occurred"

	// SuccessResult
	resSucc := SuccessResult(taskID, agentName, output)
	if resSucc.TaskID != taskID {
		t.Errorf("expected TaskID %q, got %q", taskID, resSucc.TaskID)
	}
	if resSucc.AgentName != agentName {
		t.Errorf("expected AgentName %q, got %q", agentName, resSucc.AgentName)
	}
	if !resSucc.IsSuccess() {
		t.Error("expected IsSuccess() to be true")
	}
	if resSucc.IsFailed() {
		t.Error("expected IsFailed() to be false")
	}
	if resSucc.Output != output {
		t.Errorf("expected Output %q, got %q", output, resSucc.Output)
	}

	// FailedResult
	resFail := FailedResult(taskID, agentName, errMsg)
	if resFail.TaskID != taskID {
		t.Errorf("expected TaskID %q, got %q", taskID, resFail.TaskID)
	}
	if resFail.AgentName != agentName {
		t.Errorf("expected AgentName %q, got %q", agentName, resFail.AgentName)
	}
	if resFail.IsSuccess() {
		t.Error("expected IsSuccess() to be false")
	}
	if !resFail.IsFailed() {
		t.Error("expected IsFailed() to be true")
	}
	if resFail.Error != errMsg {
		t.Errorf("expected Error %q, got %q", errMsg, resFail.Error)
	}
}

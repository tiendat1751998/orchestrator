package orchestrator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/workspace"
	"github.com/tiendat1751998/orchestrator/kernel/orchestrator"
)

type mockPlanner struct {
	planFunc  func(ctx context.Context, g goal.Goal) ([]fsm.DAG, error)
	scoreFunc func(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error)
}

func (m *mockPlanner) Plan(ctx context.Context, g goal.Goal) ([]fsm.DAG, error) {
	if m.planFunc != nil {
		return m.planFunc(ctx, g)
	}
	return []fsm.DAG{{Nodes: make(map[string]*fsm.DAGNode)}}, nil
}

func (m *mockPlanner) Score(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error) {
	if m.scoreFunc != nil {
		return m.scoreFunc(ctx, candidates)
	}
	return fsm.DAG{Nodes: make(map[string]*fsm.DAGNode)}, nil
}

func (m *mockPlanner) Explain(ctx context.Context, chosen fsm.DAG, candidates []fsm.DAG) (string, error) {
	return "mock explanation", nil
}

func (m *mockPlanner) Learn(ctx context.Context, history fsm.TransitionRecord) error {
	return nil
}

type mockTxEngine struct {
	beginFunc    func(ctx context.Context, missionID string) (workspace.TransactionID, error)
	commitFunc   func(ctx context.Context, txID workspace.TransactionID) error
	rollbackFunc func(ctx context.Context, txID workspace.TransactionID) error
}

func (m *mockTxEngine) Begin(ctx context.Context, missionID string) (workspace.TransactionID, error) {
	if m.beginFunc != nil {
		return m.beginFunc(ctx, missionID)
	}
	return workspace.TransactionID("mock-tx"), nil
}

func (m *mockTxEngine) Commit(ctx context.Context, txID workspace.TransactionID) error {
	if m.commitFunc != nil {
		return m.commitFunc(ctx, txID)
	}
	return nil
}

func (m *mockTxEngine) Rollback(ctx context.Context, txID workspace.TransactionID) error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc(ctx, txID)
	}
	return nil
}

func TestPipelineManager_Transitions(t *testing.T) {
	pm := orchestrator.NewPipelineManager()

	if pm.GetState() != orchestrator.StatePlanning {
		t.Errorf("expected initial state Planning, got %q", pm.GetState())
	}

	err := pm.Transition(orchestrator.StateScheduling)
	if err != nil {
		t.Fatalf("failed scheduling transition: %v", err)
	}

	err = pm.Transition(orchestrator.StateExecuting)
	if err != nil {
		t.Fatalf("failed executing transition: %v", err)
	}

	err = pm.Transition(orchestrator.StatePlanning)
	if err == nil {
		t.Error("expected error transitioning backwards Executing -> Planning, got nil")
	}
}

func TestSupervisor_TimeoutScans(t *testing.T) {
	sup := orchestrator.NewSupervisor(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sup.Register("task-timeout", 10*time.Millisecond)
	sup.Start(ctx, 5*time.Millisecond)

	time.Sleep(30 * time.Millisecond)
	sup.Deregister("task-timeout")
}

func TestCoordinator_DependencyInjection(t *testing.T) {
	coord := orchestrator.NewCoordinator(nil)

	task := &contractsagent.Task{
		ID:           contracts.TaskID("task-B"),
		Name:         "B",
		Dependencies: []contracts.TaskID{contracts.TaskID("task-A")},
		Input:        map[string]any{"instruction": "generate code"},
	}

	results := map[string]*contractsagent.Result{
		"task-A": {
			TaskID: contracts.TaskID("task-A"),
			Status: contracts.StatusSuccess,
			Output: "package main\n",
		},
	}

	err := coord.InjectDependencyResults(task, results)
	if err != nil {
		t.Fatalf("failed to inject dependencies: %v", err)
	}

	if task.Input["instruction"] != "generate code" {
		t.Errorf("expected instruction key to be preserved, got %v", task.Input["instruction"])
	}

	depResults, ok := task.Input["_dependency_results"].(map[string]any)
	if !ok {
		t.Fatal("missing _dependency_results key in Input")
	}

	taskARes, ok := depResults["task-A"].(map[string]any)
	if !ok {
		t.Fatal("missing task-A results inside dependencies results map")
	}

	if taskARes["output"] != "package main\n" {
		t.Errorf("expected injected output 'package main\n', got %v", taskARes["output"])
	}
}

func TestFeedbackCollector_Records(t *testing.T) {
	fc := orchestrator.NewFeedbackCollector()

	fc.RecordSuccess("mock-agent", 100, 50*time.Millisecond)
	fc.RecordFailure("mock-agent")

	metrics := fc.GetMetrics()
	m, ok := metrics["mock-agent"]
	if !ok {
		t.Fatal("expected metrics to contain 'mock-agent'")
	}

	if m.SuccessCount != 1 || m.FailureCount != 1 {
		t.Errorf("incorrect counts: success=%d, failure=%d", m.SuccessCount, m.FailureCount)
	}

	if m.TotalTokens != 100 {
		t.Errorf("expected 100 tokens, got %d", m.TotalTokens)
	}
}

func TestOrchestrator_Execute_Success(t *testing.T) {
	var beginCalled, scoreCalled, planCalled, commitCalled, rollbackCalled bool
	p := &mockPlanner{
		planFunc: func(ctx context.Context, g goal.Goal) ([]fsm.DAG, error) {
			planCalled = true
			return []fsm.DAG{{Nodes: make(map[string]*fsm.DAGNode)}}, nil
		},
		scoreFunc: func(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error) {
			scoreCalled = true
			return fsm.DAG{Nodes: make(map[string]*fsm.DAGNode)}, nil
		},
	}
	tx := &mockTxEngine{
		beginFunc: func(ctx context.Context, missionID string) (workspace.TransactionID, error) {
			beginCalled = true
			return workspace.TransactionID("tx-123"), nil
		},
		commitFunc: func(ctx context.Context, txID workspace.TransactionID) error {
			commitCalled = true
			return nil
		},
		rollbackFunc: func(ctx context.Context, txID workspace.TransactionID) error {
			rollbackCalled = true
			return nil
		},
	}

	o, err := orchestrator.NewOrchestrator(p, tx, nil)
	if err != nil {
		t.Fatalf("failed to create orchestrator: %v", err)
	}

	res, err := o.Execute(context.Background(), "mission-1", goal.Goal{})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if res == nil || res.MissionID != "mission-1" || res.Status != fsm.StateCompleted {
		t.Errorf("invalid result: %+v", res)
	}
	if !beginCalled || !planCalled || !scoreCalled || !commitCalled || rollbackCalled {
		t.Errorf("unexpected call states: begin=%t plan=%t score=%t commit=%t rollback=%t", beginCalled, planCalled, scoreCalled, commitCalled, rollbackCalled)
	}
}

func TestOrchestrator_Execute_BeginTransactionFailure(t *testing.T) {
	p := &mockPlanner{}
	tx := &mockTxEngine{
		beginFunc: func(ctx context.Context, missionID string) (workspace.TransactionID, error) {
			return "", errors.New("begin error")
		},
	}

	o, err := orchestrator.NewOrchestrator(p, tx, nil)
	if err != nil {
		t.Fatalf("failed to create orchestrator: %v", err)
	}

	res, err := o.Execute(context.Background(), "mission-1", goal.Goal{})
	if err == nil || res != nil {
		t.Errorf("expected error and nil result, got err=%v res=%v", err, res)
	}
}

func TestOrchestrator_Execute_PlanningFailure(t *testing.T) {
	var rollbackCalled bool
	p := &mockPlanner{
		planFunc: func(ctx context.Context, g goal.Goal) ([]fsm.DAG, error) {
			return nil, errors.New("plan error")
		},
	}
	tx := &mockTxEngine{
		beginFunc: func(ctx context.Context, missionID string) (workspace.TransactionID, error) {
			return workspace.TransactionID("tx-plan"), nil
		},
		rollbackFunc: func(ctx context.Context, txID workspace.TransactionID) error {
			rollbackCalled = true
			return nil
		},
	}

	o, err := orchestrator.NewOrchestrator(p, tx, nil)
	if err != nil {
		t.Fatalf("failed to create orchestrator: %v", err)
	}

	res, err := o.Execute(context.Background(), "mission-1", goal.Goal{})
	if err == nil || res != nil || !rollbackCalled {
		t.Errorf("expected error, nil result, and rollback called: err=%v res=%v rollback=%t", err, res, rollbackCalled)
	}
}

func TestOrchestrator_Execute_ScoringFailure(t *testing.T) {
	var rollbackCalled bool
	p := &mockPlanner{
		scoreFunc: func(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error) {
			return fsm.DAG{}, errors.New("score error")
		},
	}
	tx := &mockTxEngine{
		beginFunc: func(ctx context.Context, missionID string) (workspace.TransactionID, error) {
			return workspace.TransactionID("tx-score"), nil
		},
		rollbackFunc: func(ctx context.Context, txID workspace.TransactionID) error {
			rollbackCalled = true
			return nil
		},
	}

	o, err := orchestrator.NewOrchestrator(p, tx, nil)
	if err != nil {
		t.Fatalf("failed to create orchestrator: %v", err)
	}

	res, err := o.Execute(context.Background(), "mission-1", goal.Goal{})
	if err == nil || res != nil || !rollbackCalled {
		t.Errorf("expected error, nil result, and rollback called: err=%v res=%v rollback=%t", err, res, rollbackCalled)
	}
}

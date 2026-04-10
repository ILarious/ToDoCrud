package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestPlannerProcessesClaimedTasks(t *testing.T) {
	store := &plannerStoreStub{
		claimTasks: []Task{
			{ID: 1, Title: "Add worker pool"},
		},
		completeCh: make(chan int64, 1),
	}

	planner, err := NewPlanner(store, plannerGeneratorStub{
		plan: "1. Build\n2. Test",
	}, PlannerConfig{
		PoolSize:     1,
		PollInterval: time.Hour,
		StaleAfter:   time.Second,
		ClaimBatch:   1,
	})
	if err != nil {
		t.Fatalf("NewPlanner() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	planner.Start(ctx)

	select {
	case taskID := <-store.completeCh:
		if taskID != 1 {
			t.Fatalf("completed task id = %d, want 1", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("planner did not complete claimed task")
	}
}

func TestPlannerMarksTaskFailedOnGeneratorError(t *testing.T) {
	store := &plannerStoreStub{
		claimTasks: []Task{
			{ID: 7, Title: "Broken task"},
		},
		failCh: make(chan int64, 1),
	}

	planner, err := NewPlanner(store, plannerGeneratorStub{
		err: errors.New("llm unavailable"),
	}, PlannerConfig{
		PoolSize:     1,
		PollInterval: time.Hour,
		StaleAfter:   time.Second,
		ClaimBatch:   1,
	})
	if err != nil {
		t.Fatalf("NewPlanner() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	planner.Start(ctx)

	select {
	case taskID := <-store.failCh:
		if taskID != 7 {
			t.Fatalf("failed task id = %d, want 7", taskID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("planner did not fail task")
	}
}

type plannerStoreStub struct {
	mu         sync.Mutex
	claimTasks []Task
	completeCh chan int64
	failCh     chan int64
}

func (s *plannerStoreStub) ClaimTasks(_ context.Context, _ int, _ time.Duration) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := s.claimTasks
	s.claimTasks = nil
	return tasks, nil
}

func (s *plannerStoreStub) CompleteTask(_ context.Context, taskID int64, _ string) error {
	if s.completeCh != nil {
		s.completeCh <- taskID
	}
	return nil
}

func (s *plannerStoreStub) FailTask(_ context.Context, taskID int64, _ string) error {
	if s.failCh != nil {
		s.failCh <- taskID
	}
	return nil
}

type plannerGeneratorStub struct {
	plan string
	err  error
}

func (g plannerGeneratorStub) GeneratePlan(_ context.Context, _ Task) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.plan, nil
}

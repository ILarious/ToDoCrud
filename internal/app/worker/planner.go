package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"todo_crud/pkg/worker_pool"
)

type Task struct {
	ID     int64
	ListID int64
	Title  string
}

type TaskStore interface {
	ClaimTasks(ctx context.Context, limit int, staleAfter time.Duration) ([]Task, error)
	CompleteTask(ctx context.Context, taskID int64, plan string) error
	FailTask(ctx context.Context, taskID int64, reason string) error
}

type PlanGenerator interface {
	GeneratePlan(ctx context.Context, task Task) (string, error)
}

type Planner struct {
	store        TaskStore
	generator    PlanGenerator
	pool         *worker_pool.WorkerPool
	pollInterval time.Duration
	staleAfter   time.Duration
	claimBatch   int
	logger       *log.Logger
	startOnce    sync.Once
}

type PlannerConfig struct {
	PoolSize     int
	PollInterval time.Duration
	StaleAfter   time.Duration
	ClaimBatch   int
	Logger       *log.Logger
}

func NewPlanner(store TaskStore, generator PlanGenerator, cfg PlannerConfig) (*Planner, error) {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 4
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 3 * time.Second
	}
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = 30 * time.Second
	}
	if cfg.ClaimBatch <= 0 {
		cfg.ClaimBatch = cfg.PoolSize
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}

	pool, err := worker_pool.NewWorkerPool(cfg.PoolSize)
	if err != nil {
		return nil, err
	}

	return &Planner{
		store:        store,
		generator:    generator,
		pool:         pool,
		pollInterval: cfg.PollInterval,
		staleAfter:   cfg.StaleAfter,
		claimBatch:   cfg.ClaimBatch,
		logger:       cfg.Logger,
	}, nil
}

func (p *Planner) Start(ctx context.Context) {
	p.startOnce.Do(func() {
		go p.loop(ctx)
	})
}

func (p *Planner) loop(ctx context.Context) {
	defer func() {
		p.pool.Close()
		p.pool.Wait()
	}()

	p.dispatch(ctx)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.dispatch(ctx)
		}
	}
}

func (p *Planner) dispatch(ctx context.Context) {
	tasks, err := p.store.ClaimTasks(ctx, p.claimBatch, p.staleAfter)
	if err != nil {
		if ctx.Err() == nil {
			p.logger.Printf("planner claim tasks: %v", err)
		}
		return
	}

	for _, task := range tasks {
		task := task
		if err := p.pool.Submit(func() {
			p.processTask(ctx, task)
		}); err != nil {
			if ctx.Err() != nil {
				return
			}
			p.logger.Printf("planner submit task %d: %v", task.ID, err)
			_ = p.store.FailTask(context.Background(), task.ID, err.Error())
		}
	}
}

func (p *Planner) processTask(parentCtx context.Context, task Task) {
	ctx, cancel := context.WithTimeout(parentCtx, 45*time.Second)
	defer cancel()

	plan, err := p.generator.GeneratePlan(ctx, task)
	if err != nil {
		p.failTask(task.ID, err)
		return
	}

	if err := p.store.CompleteTask(ctx, task.ID, strings.TrimSpace(plan)); err != nil {
		p.failTask(task.ID, err)
		return
	}
}

func (p *Planner) failTask(taskID int64, err error) {
	if err == nil {
		return
	}

	if failErr := p.store.FailTask(context.Background(), taskID, err.Error()); failErr != nil {
		p.logger.Printf("planner fail task %d: %v (original error: %v)", taskID, failErr, err)
		return
	}

	p.logger.Printf("planner task %d failed: %v", taskID, err)
}

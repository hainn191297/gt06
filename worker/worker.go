package worker

import (
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

type Task struct {
	TraceID string
	Action  func() error
}

// WorkerPool manages a pool of workers to process tasks.
type WorkerPool struct {
	tasks      chan Task
	rg         *threading.RoutineGroup
	numWorkers int
	shutdownCh chan struct{}
	mu         *sync.Mutex
	cond       *sync.Cond
	isShutdown bool
}

// NewWorkerPool creates a new worker pool. Use when handling implicit/explicit tasks.
func NewWorkerPool(numWorkers int, bufferSize int) *WorkerPool {
	logx.Info("[WORKER] Starting...")
	mu := new(sync.Mutex)
	wp := WorkerPool{
		tasks:      make(chan Task, bufferSize),
		rg:         threading.NewRoutineGroup(),
		numWorkers: numWorkers,
		shutdownCh: make(chan struct{}),
		isShutdown: false,
		cond:       sync.NewCond(mu),
		mu:         mu,
	}
	return &wp
}

// Start initializes the worker pool and starts the workers.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.rg.RunSafe(func() {
			for {
				select {
				case task, ok := <-wp.tasks:
					if ok {
						wp.rg.RunSafe(func() {
							// Execute the task's action
							err := task.Action()
							if err != nil {
								logx.Errorf("[WORKER] Error task for reason %s with TraceID: %s", err.Error(), task.TraceID)
							} else {
								logx.Infof("[WORKER] Completed task with TraceID: %s", task.TraceID)
							}
							wp.cond.Signal()
						})
					}
				case <-wp.shutdownCh: // Check for shutdown signal
					logx.Infof("[WORKER] Worker received shutdown signal.")
					return // Exit the worker
				}
			}
		})
	}
}

// Submit adds a new task to the worker pool.
func (wp *WorkerPool) Submit(task Task) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Do not add tasks if the pool is shutting down
	if wp.isShutdown {
		logx.Info("[WORKER] Cannot submit task; worker pool is shutting down.")
		return
	}

	// Wait until there's space in the task queue (blocking behavior)
	for len(wp.tasks) == cap(wp.tasks) {
		logx.Infof("[WORKER] Queue is full. Waiting for space to submit task with TraceID: %v", task.TraceID)
		wp.cond.Wait() // Wait until signaled that space is available
	}

	// Submit the task since there is space available
	wp.tasks <- task
	logx.Infof("[WORKER] Task submitted successfully with TraceID: %v", task.TraceID)
}

// Shutdown waits for all workers to finish processing and stops the pool
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isShutdown {
		return
	}
	logx.Info("[WORKER] Received termination signal, shutting down...")
	wp.isShutdown = true

	close(wp.shutdownCh)
	wp.rg.Wait()
	logx.Info("[WORKER] All workers have finished. Worker pool shutdown complete.")
}

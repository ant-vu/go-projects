package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Priority represents the priority of a task
type Priority int

const (
	// Low represents a low-priority task
	Low Priority = iota
	// Medium represents a medium-priority task
	Medium
	// High represents a high-priority task
	High
)

// TaskFunc represents a task function
type TaskFunc func(ctx context.Context) error

// Task represents a task to be executed
type Task struct {
	id         int
	timeout    time.Duration
	priority   Priority
	fn         TaskFunc
	resultChan chan error
	retryCount int
}

// WorkerPoolConfig represents worker pool configuration
type WorkerPoolConfig struct {
	NumWorkers         int
	TaskQueueSize      int
	RetryDelay         time.Duration
	MonitoringInterval time.Duration
}

// WorkerPool represents a pool of worker goroutines
type WorkerPool struct {
	taskQueue       *priorityQueue
	wg              sync.WaitGroup
	stopChan        chan struct{}
	stats           Stats
	workerCount     int32
	retryDelay      time.Duration
	monitoringInterval time.Duration
}

// Stats represents worker pool statistics
type Stats struct {
	TasksExecuted      uint64
	TotalExecutionTime time.Duration
	TaskFailures       uint64
}

// NewWorkerPool returns a new worker pool
func NewWorkerPool(cfg WorkerPoolConfig) *WorkerPool {
	wp := &WorkerPool{
		taskQueue:       newPriorityQueue(cfg.TaskQueueSize),
		stopChan:        make(chan struct{}),
		retryDelay:      cfg.RetryDelay,
		monitoringInterval: cfg.MonitoringInterval,
	}
	wp.setWorkerCount(cfg.NumWorkers)
	go wp.monitor()
	for i := 0; i < cfg.NumWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
	return wp
}

// ExecuteTask executes a task
func (wp *WorkerPool) ExecuteTask(task Task) {
	wp.taskQueue.push(task)
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.stopChan)
	wp.wg.Wait()
}

type priorityQueue struct {
	tasks []*Task
	mu    sync.Mutex
	cond  *sync.Cond
}

func newPriorityQueue(size int) *priorityQueue {
	pq := &priorityQueue{
		tasks: make([]*Task, 0, size),
	}
	pq.cond = sync.NewCond(&pq.mu)
	return pq
}

func (pq *priorityQueue) push(task *Task) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.tasks = append(pq.tasks, task)
	pq.cond.Signal()
}

func (pq *priorityQueue) pop() *Task {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	for len(pq.tasks) == 0 {
		pq.cond.Wait()
	}
	task := pq.tasks[0]
	pq.tasks = pq.tasks[1:]
	return task
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.stopChan:
			return
		default:
			task := wp.taskQueue.pop()
			wp.executeTask(task)
		}
	}
}

func (wp *WorkerPool) executeTask(task *Task) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), task.timeout)
	defer cancel()
	err := task.fn(ctx)
	executionTime := time.Since(startTime)
	atomic.AddUint64(&wp.stats.TasksExecuted, 1)
	atomic.AddInt64(&wp.stats.TotalExecutionTime, int64(executionTime))
	if err != nil {
		atomic.AddUint64(&wp.stats.TaskFailures, 1)
		if task.retryCount > 0 {
			task.retryCount--
			time.Sleep(wp.retryDelay)
			wp.taskQueue.push(task)
		}
	}
}

func (wp *WorkerPool) monitor() {
	ticker := time.NewTicker(wp.monitoringInterval)
	defer ticker.Stop()
	for {
		select {
		case <-wp.stopChan:
			return
		case <-ticker.C:
			log.Printf("Tasks executed: %d, Total execution time: %s, Task failures: %d",
				atomic.LoadUint64(&wp.stats.TasksExecuted),
				time.Duration(atomic.LoadInt64(&wp.stats.TotalExecutionTime)),
				atomic.LoadUint64(&wp.stats.TaskFailures))
		}
	}
}

func (wp *WorkerPool) setWorkerCount(count int) {
	atomic.StoreInt32(&wp.workerCount, int32(count))
}

func (wp *WorkerPool) generateTasks() {
	taskID := 0
	for {
		select {
		case <-wp.stopChan:
			return
		default:
			task := &Task{
				id:      taskID,
				timeout: 2 * time.Second,
				priority: Priority(rand.Intn(3)),
				fn: func(ctx context.Context) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(rand.Intn(2)) * time.Second):
						return nil
					}
				},
				retryCount: 3,
			}
			wp.ExecuteTask(task)
			taskID++
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	cfg := WorkerPoolConfig{
		NumWorkers:         5,
		TaskQueueSize:      10,
		RetryDelay:         1 * time.Second,
		MonitoringInterval: 5 * time.Second,
	}
	wp := NewWorkerPool(cfg)
	go wp.generateTasks()
	time.Sleep(30 * time.Second) // Allow tasks to complete
	wp.Stop()
}
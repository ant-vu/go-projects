package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// TaskPriority represents the priority of a task
type TaskPriority int

const (
	// LowPriority represents a low-priority task
	LowPriority TaskPriority = iota
	// MediumPriority represents a medium-priority task
	MediumPriority
	// HighPriority represents a high-priority task
	HighPriority
)

// Task represents a task to be executed
type Task struct {
	id      int
	timeout time.Duration
	priority TaskPriority
	fn      func(ctx context.Context) error
	resultChan chan error
	retryCount int
}

// WorkerPool represents a pool of worker goroutines
type WorkerPool struct {
	taskQueue *priorityQueue
	wg       sync.WaitGroup
	stopChan chan struct{}
	stats    Stats
	workerCount int32
	retryDelay time.Duration
	monitoringInterval time.Duration
}

// Stats represents worker pool statistics
type Stats struct {
	TasksExecuted uint64
	TotalExecutionTime time.Duration
}

// NewWorkerPool returns a new worker pool
func NewWorkerPool(numWorkers int, taskQueueSize int, retryDelay time.Duration, monitoringInterval time.Duration) *WorkerPool {
	wp := &WorkerPool{
		taskQueue: newPriorityQueue(taskQueueSize),
		stopChan: make(chan struct{}),
		retryDelay: retryDelay,
		monitoringInterval: monitoringInterval,
	}
	wp.setWorkerCount(numWorkers)
	go wp.monitor()
	return wp
}

// setWorkerCount adjusts the worker count dynamically
func (wp *WorkerPool) setWorkerCount(numWorkers int) {
	currentCount := atomic.LoadInt32(&wp.workerCount)
	if numWorkers > int(currentCount) {
		wp.addWorkers(numWorkers - int(currentCount))
	} else if numWorkers < int(currentCount) {
		wp.removeWorkers(int(currentCount) - numWorkers)
	}
}

func (wp *WorkerPool) addWorkers(count int) {
	wp.wg.Add(count)
	atomic.AddInt32(&wp.workerCount, int32(count))
	for i := 0; i < count; i++ {
		go wp.worker()
	}
}

func (wp *WorkerPool) removeWorkers(count int) {
	for i := 0; i < count; i++ {
		wp.stopChan <- struct{}{}
	}
	atomic.AddInt32(&wp.workerCount, int32(-count))
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case task := <-wp.taskQueue.highPriority:
			wp.executeTask(task)
		case task := <-wp.taskQueue.mediumPriority:
			wp.executeTask(task)
		case task := <-wp.taskQueue.lowPriority:
			wp.executeTask(task)
		case <-wp.stopChan:
			return
		}
	}
}

func (wp *WorkerPool) executeTask(task Task) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), task.timeout)
	err := task.fn(ctx)
	cancel()
	if task.resultChan != nil {
		task.resultChan <- err
	}
	if err != nil {
		log.Printf("Task %d failed with error: %v", task.id, err)
		if task.retryCount > 0 {
			task.retryCount--
			time.Sleep(wp.retryDelay)
			wp.taskQueue.push(task)
		}
	} else {
		atomic.AddUint64(&wp.stats.TasksExecuted, 1)
		atomic.AddInt64((*int64)(&wp.stats.TotalExecutionTime), int64(time.Since(startTime)))
		log.Printf("Worker executed task %d in %v", task.id, time.Since(startTime))
	}
}

// ExecuteTask sends a task to the worker pool
func (wp *WorkerPool) ExecuteTask(task Task) {
	wp.taskQueue.push(task)
}

// ExecuteTaskWithResult sends a task to the worker pool and returns the result
func (wp *WorkerPool) ExecuteTaskWithResult(task Task) error {
	task.resultChan = make(chan error, 1)
	wp.taskQueue.push(task)
	return <-task.resultChan
}

// AdjustWorkerCount adjusts the worker count dynamically
func (wp *WorkerPool) AdjustWorkerCount(numWorkers int) {
	wp.setWorkerCount(numWorkers)
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.taskQueue.highPriority)
	close(wp.taskQueue.mediumPriority)
	close(wp.taskQueue.lowPriority)
	wp.wg.Wait()
}

func (wp *WorkerPool) GetStats() Stats {
	return Stats{
		TasksExecuted: atomic.LoadUint64(&wp.stats.TasksExecuted),
		TotalExecutionTime: time.Duration(atomic.LoadInt64((*int64)(&wp.stats.TotalExecutionTime))),
	}
}

func (wp *WorkerPool) monitor() {
	ticker := time.NewTicker(wp.monitoringInterval)
	defer ticker.Stop()
	prevTasksExecuted := wp.GetStats().TasksExecuted
	for range ticker.C {
		stats := wp.GetStats()
		taskExecutionRate := float64(stats.TasksExecuted-prevTasksExecuted) / wp.monitoringInterval.Seconds()
		prevTasksExecuted = stats.TasksExecuted
		queueSize := len(wp.taskQueue.highPriority) + len(wp.taskQueue.mediumPriority) + len(wp.taskQueue.lowPriority)
		if taskExecutionRate < float64(queueSize) {
			wp.AdjustWorkerCount(int(atomic.LoadInt32(&wp.workerCount)) + 1)
		} else if taskExecutionRate > float64(queueSize)*2 {
			wp.AdjustWorkerCount(int(atomic.LoadInt32(&wp.workerCount)) - 1)
		}
		log.Printf("Worker pool stats: %+v", wp.GetStats())
	}
}

func newPriorityQueue(taskQueueSize int) *priorityQueue {
	return &priorityQueue{
		highPriority: make(chan Task, taskQueueSize),
		mediumPriority: make(chan Task, taskQueueSize),
		lowPriority: make(chan Task, taskQueueSize),
	}
}

type priorityQueue struct {
	highPriority chan Task
	mediumPriority chan Task
	lowPriority chan Task
}

func (pq *priorityQueue) push(task Task) {
	switch task.priority {
	case HighPriority:
		pq.highPriority <- task
	case MediumPriority:
		pq.mediumPriority <- task
	case LowPriority:
		pq.lowPriority <- task
	}
}

func main() {
	wp := NewWorkerPool(5, 10, 1*time.Second, 5*time.Second)
	for i := 0; i < 10; i++ {
		task := Task{
			id:      i,
			timeout: 2 * time.Second,
			priority: MediumPriority,
			fn: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
					return nil
				}
			},
			retryCount: 3,
		}
		err := wp.ExecuteTaskWithResult(task)
		if err != nil {
			log.Printf("Task %d failed with error: %v", task.id, err)
		}
	}
	time.Sleep(30 * time.Second) // Allow tasks to complete
	wp.Stop()
}
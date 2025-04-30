package main

import (
	"context"
	"fmt"
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
	taskQueue chan Task
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
		taskQueue: make(chan Task, taskQueueSize),
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
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}
			startTime := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), task.timeout)
			err := task.fn(ctx)
			cancel()
			if task.resultChan != nil {
				task.resultChan <- err
			}
			if err != nil {
				if task.retryCount > 0 {
					task.retryCount--
					time.Sleep(wp.retryDelay)
					wp.taskQueue <- task
				} else {
					fmt.Printf("Task %d failed with error: %v\n", task.id, err)
				}
			} else {
				atomic.AddUint64(&wp.stats.TasksExecuted, 1)
				atomic.AddInt64((*int64)(&wp.stats.TotalExecutionTime), int64(time.Since(startTime)))
				fmt.Printf("Worker executed task %d in %v\n", task.id, time.Since(startTime))
			}
		case <-wp.stopChan:
			return
		}
	}
}

// ExecuteTask sends a task to the worker pool
func (wp *WorkerPool) ExecuteTask(task Task) {
	wp.taskQueue <- task
}

// ExecuteTaskWithResult sends a task to the worker pool and returns the result
func (wp *WorkerPool) ExecuteTaskWithResult(task Task) error {
	task.resultChan = make(chan error, 1)
	wp.ExecuteTask(task)
	return <-task.resultChan
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.taskQueue)
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
	for range ticker.C {
		fmt.Printf("Worker pool stats: %+v\n", wp.GetStats())
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
			fmt.Printf("Task %d failed with error: %v\n", task.id, err)
		}
	}
	time.Sleep(30 * time.Second) // Allow tasks to complete
	wp.Stop()
}
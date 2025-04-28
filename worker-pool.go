package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a task to be executed
type Task struct {
	id      int
	timeout time.Duration
	fn      func(ctx context.Context) error
}

// WorkerPool represents a pool of worker goroutines
type WorkerPool struct {
	taskChan chan Task
	wg       sync.WaitGroup
	stopChan chan struct{}
	stats    Stats
	workerCount int32
}

// Stats represents worker pool statistics
type Stats struct {
	TasksExecuted uint64
	TotalExecutionTime time.Duration
}

// NewWorkerPool returns a new worker pool
func NewWorkerPool(numWorkers int, taskQueueSize int) *WorkerPool {
	wp := &WorkerPool{
		taskChan: make(chan Task, taskQueueSize),
		stopChan: make(chan struct{}),
	}
	wp.setWorkerCount(numWorkers)
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
		case task, ok := <-wp.taskChan:
			if !ok {
				return
			}
			startTime := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), task.timeout)
			err := task.fn(ctx)
			cancel()
			if err != nil {
				fmt.Printf("Task %d failed with error: %v\n", task.id, err)
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
	wp.taskChan <- task
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.taskChan)
	wp.wg.Wait()
}

func (wp *WorkerPool) GetStats() Stats {
	return Stats{
		TasksExecuted: atomic.LoadUint64(&wp.stats.TasksExecuted),
		TotalExecutionTime: time.Duration(atomic.LoadInt64((*int64)(&wp.stats.TotalExecutionTime))),
	}
}

func main() {
	wp := NewWorkerPool(5, 10)
	for i := 0; i < 10; i++ {
		task := Task{
			id:      i,
			timeout: 2 * time.Second,
			fn: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
					return nil
				}
			},
		}
		wp.ExecuteTask(task)
	}
	time.Sleep(5 * time.Second) // Allow tasks to complete
	fmt.Printf("Worker pool stats: %+v\n", wp.GetStats())
	wp.Stop()
}
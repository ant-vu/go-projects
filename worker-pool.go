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
	go wp.generateTasks()
	return wp
}

func (wp *WorkerPool) generateTasks() {
	taskID := 0
	for {
		task := Task{
			id:      taskID,
			timeout: 2 * time.Second,
			priority: TaskPriority(rand.Intn(3)),
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

// ... (rest of the code remains the same)

func main() {
	wp := NewWorkerPool(5, 10, 1*time.Second, 5*time.Second)
	time.Sleep(30 * time.Second) // Allow tasks to complete
	wp.Stop()
}
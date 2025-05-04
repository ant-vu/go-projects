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

// generateTasks generates tasks for demonstration purposes
func (wp *WorkerPool) generateTasks() {
	taskID := 0
	for {
		select {
		case <-wp.stopChan:
			return
		default:
			task := Task{
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
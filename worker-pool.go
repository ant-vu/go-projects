package main

import (
	"fmt"
	"sync"
	"time"
)

// Task represents a task to be executed
type Task struct {
	id int
}

// WorkerPool represents a pool of worker goroutines
type WorkerPool struct {
	taskChan chan Task
	wg       sync.WaitGroup
}

// NewWorkerPool returns a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	wp := &WorkerPool{
		taskChan: make(chan Task),
	}
	wp.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wp.wg.Done()
			for task := range wp.taskChan {
				fmt.Printf("Worker executing task %d\n", task.id)
				time.Sleep(1 * time.Second)
			}
		}()
	}
	return wp
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

func main() {
	wp := NewWorkerPool(5)
	for i := 0; i < 10; i++ {
		wp.ExecuteTask(Task{id: i})
	}
	wp.Stop()
}
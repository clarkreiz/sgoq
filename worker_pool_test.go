package main

import (
	"sync"
	"testing"
	"time"
)

// MockQueue implements Queue interface for testing
type MockQueue struct {
	tasks  []*Task
	mu     sync.Mutex
	stop   bool
	closed bool
}

func (mq *MockQueue) Enqueue(task *Task) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if mq.closed {
		return nil
	}
	mq.tasks = append(mq.tasks, task)
	return nil
}

func (mq *MockQueue) Dequeue() *Task {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	if len(mq.tasks) == 0 {
		return nil
	}
	task := mq.tasks[0]
	mq.tasks = mq.tasks[1:]
	return task
}

func (mq *MockQueue) Close() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.closed = true
}

func (mq *MockQueue) IsEmpty() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return len(mq.tasks) == 0
}

func (mq *MockQueue) Stop() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.stop = true
}

func (mq *MockQueue) IsStopped() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return mq.stop
}

func TestNewWorkerPool(t *testing.T) {
	queue := &MockQueue{}
	initialSize := 5

	wp := NewWorkerPool(queue, initialSize)

	if wp == nil {
		t.Error("Expected non-nil WorkerPool")
	}

	if wp.Current() != initialSize {
		t.Errorf("Expected %d workers, got %d", initialSize, wp.Current())
	}
}

func TestWorkerPool_Scale(t *testing.T) {
	tests := []struct {
		name        string
		initialSize int
		delta       int
		wantSize    int
	}{
		{"Scale up", 2, 3, 5},
		{"Scale down", 5, -2, 3},
		{"Scale down to zero", 3, -3, 0},
		{"Scale down more than exists", 3, -5, 0},
		{"No change", 3, 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &MockQueue{}
			wp := NewWorkerPool(queue, tt.initialSize)

			wp.Scale(tt.delta)

			// Give time for goroutines to complete
			time.Sleep(100 * time.Millisecond)

			if got := wp.Current(); got != tt.wantSize {
				t.Errorf("WorkerPool.Scale() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

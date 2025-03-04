package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func NewMockTask(priority int, d time.Duration) *Task {
	task := &Task{
		ID:        rand.Intn(100),
		Priority:  priority,
		Payload:   struct{}{},
		CreatedAt: time.Now(),
		executed:  false,
	}

	task.Exe = func() {
		time.Sleep(d)
		task.executed = true
	}

	return task
}

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

func TestStartProcessing(t *testing.T) {
	mq := &MockQueue{}
	concurrency := 2
	worker := NewWorker(mq, concurrency)
	var wg sync.WaitGroup

	task1 := NewMockTask(0, time.Millisecond*250)
	task2 := NewMockTask(0, time.Millisecond*230)
	mq.Enqueue(task1)
	mq.Enqueue(task2)

	worker.StartProcessing(&wg)

	time.Sleep(time.Millisecond * 500)

	if !task1.executed || !task2.executed {
		t.Error("Expected tasks to be executed")
	}

	worker.cancel()
	wg.Wait()
}

func TestShutdownTimeout(t *testing.T) {
	mq := &MockQueue{}
	concurrency := 2
	worker := NewWorker(mq, concurrency)
	var wg sync.WaitGroup

	worker.StartProcessing(&wg)

	longTask := NewMockTask(0, time.Hour)
	mq.Enqueue(longTask)

	// start Shutdown with small timeout
	shutdownTimeout := time.Millisecond * 10
	worker.Shutdown(&wg, shutdownTimeout)

	if !mq.IsStopped() {
		t.Error("Expected queue to be stopped")
	}

	if longTask.executed {
		t.Error("Expected task to not be executed due to timeout")
	}
}

package main

import (
	"sync"
	"testing"
)

func TestPriorityQueue_BasicOperations(t *testing.T) {
	pq := NewPriorityQueue(10)
	task := &Task{ID: 1, Priority: 2, Exe: func() {}}
	pq.Enqueue(task)
	deqTask := pq.Dequeue()

	if deqTask == nil {
		t.Fatalf("Expected a task, got nil")
	}
	if deqTask.ID != task.ID {
		t.Fatalf("Dequeued task ID mismatch: got %d, want %d", deqTask.ID, task.ID)
	}
}

func TestPriorityQueue_PriorityOrder(t *testing.T) {
	pq := NewPriorityQueue(10)
	tasks := []*Task{
		{ID: 1, Priority: 2},
		{ID: 2, Priority: 1},
		{ID: 3, Priority: 0},
	}

	for _, task := range tasks {
		pq.Enqueue(task)
	}

	if pq.Dequeue().ID != 3 {
		t.Fatalf("Expected task ID 3 (priority 0), got different")
	}
	if pq.Dequeue().ID != 2 {
		t.Fatalf("Expected task ID 2 (priority 1), got different")
	}
	if pq.Dequeue().ID != 1 {
		t.Fatalf("Expected task ID 1 (priority 2), got different")
	}
}

func TestPriorityQueue_Buffer(t *testing.T) {
	qsize := 50
	pq := NewPriorityQueue(qsize)
	var wg sync.WaitGroup
	totalTasks := int32(100)
	numPriorities := 5

	for i := int32(0); i < totalTasks; i++ {
		wg.Add(1)
		go func(id int32) {
			defer wg.Done()
			task := &Task{ID: int(id), Priority: int(id) % numPriorities}
			pq.Enqueue(task)
		}(i)
	}

	wg.Wait()

	if pq.totalTask.Load() != int32(qsize) {
		t.Fatalf("Expected total tasks %d, got %d", qsize, pq.totalTask.Load())
	}
}

func TestPriorityQueue_StopBehavior(t *testing.T) {
	pq := NewPriorityQueue(10)
	task := &Task{ID: 1, Priority: 1}
	pq.Stop()
	pq.Enqueue(task)
	if !pq.IsStopped() {
		t.Fatalf("Expected queue to be stopped")
	}
	if !pq.IsEmpty() {
		t.Fatalf("Expected queue to be empty after stopping")
	}
}

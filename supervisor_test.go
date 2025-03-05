package main

import (
	"sync"
	"testing"
	"time"
)

type MockWorkerManager struct {
	mu          sync.RWMutex
	concurrency int
	t           *testing.T
}

func (m *MockWorkerManager) Scale(delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.concurrency += delta
}

func (m *MockWorkerManager) Current() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.concurrency
}

type MockQueueManager struct {
	capacity   int
	totalTasks int
}

func (m *MockQueueManager) GetCapacity() int   { return m.capacity }
func (m *MockQueueManager) GetTotalTasks() int { return m.totalTasks }

func TestSupervisor(t *testing.T) {
	tests := []struct {
		name            string
		queueCapacity   int
		totalTasks      int
		minWorkers      int
		maxWorkers      int
		initWorkers     int
		expectedWorkers int
	}{
		{
			name:            "Scale up when utilization is high",
			queueCapacity:   100,
			totalTasks:      80, // 80% utilization
			minWorkers:      1,
			maxWorkers:      10,
			initWorkers:     5,
			expectedWorkers: 6, // 5 * 1.2 = 6
		},
		{
			name:            "Scale down when utilization is low",
			queueCapacity:   100,
			totalTasks:      20, // 20% utilization
			minWorkers:      1,
			maxWorkers:      10,
			initWorkers:     5,
			expectedWorkers: 4, // 5 * 0.8 = 4
		},
		{
			name:            "Don't scale up beyond max workers",
			queueCapacity:   100,
			totalTasks:      90, // 90% utilization
			minWorkers:      1,
			maxWorkers:      5,
			initWorkers:     5,
			expectedWorkers: 5,
		},
		{
			name:            "Don't scale down below min workers",
			queueCapacity:   100,
			totalTasks:      10, // 10% utilization
			minWorkers:      3,
			maxWorkers:      10,
			initWorkers:     3,
			expectedWorkers: 3,
		},
		{
			name:            "No change when utilization is within thresholds",
			queueCapacity:   100,
			totalTasks:      50, // 50% utilization
			minWorkers:      1,
			maxWorkers:      10,
			initWorkers:     5,
			expectedWorkers: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := &MockWorkerManager{concurrency: tt.initWorkers, t: t}
			queue := &MockQueueManager{
				capacity:   tt.queueCapacity,
				totalTasks: tt.totalTasks,
			}

			supervisor := NewSupervisor(worker, queue, tt.minWorkers, tt.maxWorkers)

			supervisor.Start()
			time.Sleep(2 * time.Second)
			supervisor.Stop()

			if worker.Current() != tt.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tt.expectedWorkers, worker.Current())
			}
		})
	}
}

func TestSupervisorEdgeCases(t *testing.T) {
	t.Run("Zero capacity queue", func(t *testing.T) {
		worker := &MockWorkerManager{concurrency: 5, t: t}
		queue := &MockQueueManager{capacity: 0, totalTasks: 0}

		supervisor := NewSupervisor(worker, queue, 1, 10)

		supervisor.Start()
		time.Sleep(2 * time.Second)
		supervisor.Stop()

		if worker.Current() != 5 {
			t.Errorf("Expected workers to remain at 5, got %d", worker.Current())
		}
	})

	t.Run("Equal min and max workers", func(t *testing.T) {
		worker := &MockWorkerManager{concurrency: 5, t: t}
		queue := &MockQueueManager{capacity: 100, totalTasks: 80}

		supervisor := NewSupervisor(worker, queue, 5, 5)

		supervisor.Start()
		time.Sleep(2 * time.Second)
		supervisor.Stop()

		if worker.Current() != 5 {
			t.Errorf("Expected workers to remain at 5, got %d", worker.Current())
		}
	})
}

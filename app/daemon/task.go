package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusPaused    TaskStatus = "paused"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // "download", "upload", "forward"
	Status    TaskStatus  `json:"status"`
	Name      string      `json:"name"`
	Size      int64       `json:"size"`
	Downloaded int64      `json:"downloaded"`
	Speed     int64       `json:"speed"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`

	// internal context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

type TaskManager struct {
	tasks map[string]*Task
	mu    sync.RWMutex
	hub   *Hub
}

func NewTaskManager(hub *Hub) *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
		hub:   hub,
	}
}

func (m *TaskManager) Run(ctx context.Context) {
	// periodic broadcast of all task progresses
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.broadcastProgress()
		}
	}
}

func (m *TaskManager) broadcastProgress() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeTasks := make([]*Task, 0)
	for _, task := range m.tasks {
		if task.Status == StatusRunning {
			activeTasks = append(activeTasks, task)
		}
	}

	if len(activeTasks) > 0 {
		m.hub.Broadcast("tasks_progress", activeTasks)
	}
}

func (m *TaskManager) AddTask(t *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.Status = StatusPending
	
	m.tasks[t.ID] = t
	m.hub.Broadcast("task_added", t)
}

func (m *TaskManager) UpdateTaskStatus(id string, status TaskStatus, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, ok := m.tasks[id]; ok {
		task.Status = status
		if err != nil {
			task.Error = err.Error()
		}
		m.hub.Broadcast("task_updated", task)
	}
}

func (m *TaskManager) GetTask(id string) *Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[id]
}

func (m *TaskManager) DeleteTask(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, ok := m.tasks[id]; ok {
		if task.cancel != nil {
			task.cancel()
		}
		delete(m.tasks, id)
		m.hub.Broadcast("task_deleted", id)
	}
}
func (m *TaskManager) GetAllTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		list = append(list, task)
	}
	return list
}

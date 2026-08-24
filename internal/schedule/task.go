package schedule

import "reconcilesvc/internal/model"

// TaskState 表示对账任务状态。
type TaskState int

const (
	StatePending TaskState = iota
	StateRunning
	StatePartial
	StateDone
)

// String 返回任务状态的稳定名称。
func (s TaskState) String() string {
	switch s {
	case StatePending:
		return "pending"
	case StateRunning:
		return "running"
	case StatePartial:
		return "partial"
	case StateDone:
		return "done"
	default:
		return "unknown"
	}
}

// Task 表示一次对账任务。
type Task struct {
	ID      string
	State   TaskState
	Windows []model.Window
	Current int
	Retries int
	Err     error
}

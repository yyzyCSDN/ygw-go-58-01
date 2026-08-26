package schedule

import "fmt"

// Progress 返回任务进度文本。
func (t *Task) Progress() string {
	if t == nil || len(t.Windows) == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", t.Current+1, len(t.Windows))
}

// Summary 返回任务摘要。
func (t *Task) Summary() string {
	if t == nil {
		return "task state=unknown"
	}
	return fmt.Sprintf("task %s state=%s progress=%s retries=%d", t.ID, t.State.String(), t.Progress(), t.Retries)
}

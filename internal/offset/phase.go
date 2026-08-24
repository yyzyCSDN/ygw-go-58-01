package offset

import (
	"context"
	"fmt"

	"reconcilesvc/internal/model"
)

// SwitchToIncremental 将位点阶段切换为增量，并以全量最后一窗为锚点。
func (m *Manager) SwitchToIncremental(ctx context.Context, anchor model.Window) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if anchor.End == "" || anchor.Start > anchor.End {
		return fmt.Errorf("invalid incremental anchor: %+v", anchor)
	}
	if anchor.Phase != model.PhaseFull {
		return fmt.Errorf("anchor must belong to the full phase, got %s", anchor.Phase)
	}
	if m.pos.Key != "" && anchor.End < m.pos.Key {
		return fmt.Errorf("incremental anchor %s is behind position %s", anchor.End, m.pos.Key)
	}
	next := model.Position{
		Phase:       model.PhaseIncremental,
		WindowIndex: anchor.Index,
		Key:         anchor.End,
		Completed:   true,
	}
	if err := m.store.Save(ctx, next); err != nil {
		return fmt.Errorf("persist phase switch: %w", err)
	}
	m.pos = next
	if m.journal != nil {
		m.journal.Append(next)
	}
	return nil
}

// AnchorWindow 返回全量最后一窗，作为切换增量的锚点。
func AnchorWindow(windows []model.Window) (model.Window, error) {
	if len(windows) == 0 {
		return model.Window{}, fmt.Errorf("no windows to anchor")
	}
	return windows[len(windows)-1], nil
}

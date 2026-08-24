package schedule

import (
	"context"
	"testing"

	"reconcilesvc/internal/model"
)

func TestPlannerProducesWindows(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}
	windows := NewPlanner(2).Plan(keys, model.PhaseFull)
	if len(windows) == 0 {
		t.Fatalf("planner produced no windows")
	}
	for _, w := range windows {
		if w.Start > w.End {
			t.Fatalf("window reversed: %+v", w)
		}
	}
}

func TestResumeWindowsFilters(t *testing.T) {
	planner := NewPlanner(1)
	keys := []string{"a", "b", "c"}
	planned := planner.Plan(keys, model.PhaseIncremental)
	all := planner.ResumeWindows(keys, model.PhaseIncremental, "")
	if len(all) != len(planned) {
		t.Fatalf("empty resume key should keep all planned windows: %d != %d", len(all), len(planned))
	}
	filtered := planner.ResumeWindows(keys, model.PhaseIncremental, "a")
	if len(filtered) >= len(planned) {
		t.Fatalf("resume filter should drop windows at or before the resume key")
	}
}

func TestRunnerRunsAll(t *testing.T) {
	var calls int
	err := NewRunner(2).Run(context.Background(), 4, func(ctx context.Context, index int) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 4 {
		t.Fatalf("calls = %d, want 4", calls)
	}
}

func TestRetryPolicy(t *testing.T) {
	policy := NewRetryPolicy(2)
	if !policy.ShouldRetry(0) || !policy.ShouldRetry(1) || policy.ShouldRetry(2) {
		t.Fatalf("retry policy boundaries wrong")
	}
	if NewRetryPolicy(0).MaxAttempts != 1 {
		t.Fatalf("retry policy should default to 1")
	}
}

func TestTaskStateNames(t *testing.T) {
	if StatePending.String() != "pending" || StateRunning.String() != "running" ||
		StatePartial.String() != "partial" || StateDone.String() != "done" {
		t.Fatalf("task state names wrong")
	}
}

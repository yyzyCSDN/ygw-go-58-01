package schedule

import (
	"sort"
	"testing"

	"reconcilesvc/internal/model"
)

// TestPlanCoversAllKeysNoGaps verifies the core reconciliation invariant:
// every window in a run must be contiguous with its neighbours, and the
// union of all windows must cover every key exactly with no gap.
//
// This guards against the regression where the trailing window was dropped,
// leaving a fixed window of data never reconciled while the run still
// reported "all passed".
func TestPlanCoversAllKeysNoGaps(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"} // 5 keys, size 2 -> 3 windows
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)

	windows := NewPlanner(2).Plan(keys, model.PhaseFull)

	// Sanity: we expect exactly ceil(5/2) = 3 windows. The trailing window
	// must not be silently dropped.
	if got, want := len(windows), 3; got != want {
		t.Fatalf("got %d windows, want %d (trailing window dropped?)", got, want)
	}

	// Indices must be contiguous 0..n-1 and Total consistent.
	for i, w := range windows {
		if w.Index != i {
			t.Fatalf("window %d has Index %d, windows not contiguous", i, w.Index)
		}
		if w.Total != len(windows) {
			t.Fatalf("window %d Total %d != %d", i, w.Total, len(windows))
		}
	}

	// Adjacent windows must be contiguous: window[i].End is the key just
	// before window[i+1].Start in sorted order (no gap between them).
	for i := 0; i < len(windows)-1; i++ {
		cur, next := windows[i], windows[i+1]
		if cur.End >= next.Start {
			// End >= Start would mean overlap or touch; for a key-range
			// [Start, End] plan, the next Start must strictly follow the
			// current End in sorted order.
			t.Fatalf("gap between window %d (end %s) and window %d (start %s)",
				cur.Index, cur.End, next.Index, next.Start)
		}
	}

	// Every key must fall into exactly one window.
	covered := 0
	for _, w := range windows {
		for _, k := range sorted {
			if w.Contains(k) {
				covered++
			}
		}
	}
	if covered != len(sorted) {
		t.Fatalf("covered %d keys, want %d (some keys never reconciled)", covered, len(sorted))
	}
}

// TestPlanTailWindowNotDropped checks the exact symptom from the bug report:
// the last window's data was never reconciled. With size that does not evenly
// divide total, the final (partial) window must still be emitted.
func TestPlanTailWindowNotDropped(t *testing.T) {
	cases := []struct {
		name    string
		keys    []string
		size    int
		wantWin int
	}{
		{"odd keys size 2", []string{"a", "b", "c", "d", "e"}, 2, 3},
		{"odd keys size 3", []string{"a", "b", "c", "d", "e"}, 3, 2},
		{"even keys size 2", []string{"a", "b", "c", "d"}, 2, 2},
		{"single key", []string{"a"}, 2, 1},
		{"size larger than total", []string{"a", "b"}, 5, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := NewPlanner(tc.size).Plan(tc.keys, model.PhaseFull)
			if len(ws) != tc.wantWin {
				t.Fatalf("got %d windows, want %d", len(ws), tc.wantWin)
			}
			// The last key must be covered by the last window's End.
			sorted := append([]string(nil), tc.keys...)
			sort.Strings(sorted)
			last := ws[len(ws)-1]
			if last.End != sorted[len(sorted)-1] {
				t.Fatalf("last window End = %s, want %s (tail key not covered)",
					last.End, sorted[len(sorted)-1])
			}
		})
	}
}

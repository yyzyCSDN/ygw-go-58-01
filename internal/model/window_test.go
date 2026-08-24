package model

import "testing"

func TestWindowContains(t *testing.T) {
	w := Window{ID: "w0", Start: "b", End: "d"}
	for key, want := range map[string]bool{"b": true, "c": true, "d": true, "a": false, "e": false} {
		if got := w.Contains(key); got != want {
			t.Fatalf("Contains(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestPhaseString(t *testing.T) {
	if PhaseFull.String() != "full" || PhaseIncremental.String() != "incremental" {
		t.Fatalf("unexpected phase names")
	}
}

func TestEntryDedupKey(t *testing.T) {
	e := Entry{Key: "k1", Kind: DiffMismatch}
	if got, want := e.DedupKey(), "k1@mismatch"; got != want {
		t.Fatalf("DedupKey() = %q, want %q", got, want)
	}
}

func TestSnapshotGet(t *testing.T) {
	s := &Snapshot{Version: 3, Records: map[string]*Record{"a": {Key: "a", Version: 1}}}
	if rec, ok := s.Get("a"); !ok || rec.Version != 1 {
		t.Fatalf("Get(a) = %v, %v", rec, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatalf("Get(missing) should not be found")
	}
}

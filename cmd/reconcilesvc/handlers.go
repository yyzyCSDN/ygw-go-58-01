package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/offset"
	"reconcilesvc/internal/schedule"
	"reconcilesvc/internal/source"
	"reconcilesvc/internal/target"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := source.Probe(r.Context(), a.src); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "source": a.src.Name()})
}

func (a *app) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	a.mu.Lock()
	task := &schedule.Task{ID: "task-" + time.Now().Format("150405"), State: schedule.StatePending}
	a.tasks[task.ID] = task
	a.mu.Unlock()
	result, err := a.scheduler.Run(r.Context(), task)
	a.mu.Lock()
	a.lastResult = result
	a.mu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"task_id": task.ID, "state": task.State.String(), "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":  task.ID,
		"state":    task.State.String(),
		"total":    result.Total,
		"missing":  result.Missing,
		"mismatch": result.Mismatch,
		"extra":    result.Extra,
	})
}

func (a *app) handleStatus(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	tasks := make([]map[string]any, 0, len(a.tasks))
	for id, task := range a.tasks {
		tasks = append(tasks, map[string]any{
			"id":       id,
			"state":    task.State.String(),
			"current":  task.Current,
			"progress": task.Progress(),
			"retries":  task.Retries,
			"error":    errString(task.Err),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (a *app) handleDiffs(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	entries, err := diff.ListFiltered(r.Context(), a.diffs, diff.Filter{Kind: kind, Limit: limit})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(entries),
		"entries": entries,
		"counts":  diff.Counts(entries),
	})
}

func (a *app) handleOffsets(w http.ResponseWriter, r *http.Request) {
	pos := a.offsets.Position()
	writeJSON(w, http.StatusOK, map[string]any{
		"phase":        pos.Phase.String(),
		"window_index": pos.WindowIndex,
		"key":          pos.Key,
		"completed":    pos.Completed,
	})
}

func (a *app) handleSwitchIncremental(w http.ResponseWriter, r *http.Request) {
	windows := a.scheduler.FullWindows()
	anchor, err := offset.AnchorWindow(windows)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := a.offsets.SwitchToIncremental(r.Context(), anchor); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	pos := a.offsets.Position()
	writeJSON(w, http.StatusOK, map[string]any{
		"phase":        pos.Phase.String(),
		"window_index": pos.WindowIndex,
		"key":          pos.Key,
	})
}

func (a *app) handleSourceStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"active_conns": a.pool.Active(),
		"total_conns":  a.pool.Total(),
	})
}

func (a *app) handleSourceReads(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"reads": a.tracker.Stats()})
}

func (a *app) handleNotifySummary(w http.ResponseWriter, r *http.Request) {
	summary := a.counting.Snapshot()
	select {
	case msg := <-a.channel.Chan():
		writeJSON(w, http.StatusOK, map[string]any{"summary": summary, "latest": msg})
	default:
		writeJSON(w, http.StatusOK, map[string]any{"summary": summary, "latest": nil})
	}
}

func (a *app) handleNotifyHistory(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"messages": a.history.Messages()})
}

func (a *app) handleTargetStatus(w http.ResponseWriter, r *http.Request) {
	snapshot, err := a.snapshots.Latest(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": snapshot.Version,
		"keys":    len(snapshot.Records),
		"history": a.snapshots.History(),
	})
}

func (a *app) handleTargetUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	snapshot, err := a.snapshots.Latest(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	next := nextTargetRecords(snapshot.Records)
	a.targetStore.Update(snapshot.Version+1, next)
	a.targetLog.Append(target.VersionRecord{Version: snapshot.Version + 1, Keys: len(next)})
	writeJSON(w, http.StatusOK, map[string]any{"version": snapshot.Version + 1, "keys": len(next)})
}

func (a *app) handleOffsetsHistory(w http.ResponseWriter, r *http.Request) {
	positions := a.journal.Entries()
	check := "ok"
	if err := offset.ValidateProgress(positions); err != nil {
		check = err.Error()
	}
	writeJSON(w, http.StatusOK, map[string]any{"positions": positions, "check": check})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

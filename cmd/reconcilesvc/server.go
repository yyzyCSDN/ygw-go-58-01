package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"reconcilesvc/internal/compare"
	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/notify"
	"reconcilesvc/internal/offset"
	"reconcilesvc/internal/schedule"
	"reconcilesvc/internal/source"
	"reconcilesvc/internal/target"
)

// app 聚合对账引擎的各个组件。
type app struct {
	scheduler   *schedule.Scheduler
	diffs       *diff.MemoryStore
	offsets     *offset.Manager
	notifier    *notify.Notifier
	counting    *notify.CountingSubscriber
	channel     *notify.ChannelSubscriber
	pool        *source.ConnPool
	keys        []string
	targetStore *target.MemoryStore
	snapshots   *target.SnapshotService
	targetLog   *target.UpdateLog
	journal     *offset.Journal
	history     *notify.History
	tracker     *source.ReadTracker
	src         source.Source

	mu         sync.Mutex
	tasks      map[string]*schedule.Task
	lastResult *model.Result
}

// buildApp 组装全部组件并准备演示数据。
func buildApp() (*app, error) {
	pool := source.NewConnPool()
	sourceRecords := seedSourceRecords()
	keys := make([]string, 0, len(sourceRecords))
	for _, rec := range sourceRecords {
		keys = append(keys, rec.Key)
	}
	targetStore := target.NewMemoryStore(1, seedTargetRecords())
	targetLog := target.NewUpdateLog()
	baseline := target.NewBaselineManager(targetStore)
	diffStore := diff.NewMemoryStore()
	sink := diff.NewSink(diffStore, 1)
	engine := compare.NewEngine(baseline, sink, compare.DefaultOptions())
	offsetStore := offset.NewMemoryStore()
	journal := offset.NewJournal()
	offsets := offset.NewManagerWithJournal(offsetStore, journal)
	counting := notify.NewCountingSubscriber()
	channel := notify.NewChannelSubscriber(64)
	history := notify.NewHistory(32)
	notifier := notify.NewNotifier(counting, channel, notify.NewHistorySubscriber(history))
	planner := schedule.NewPlanner(64)
	tracker := source.NewReadTracker()
	src := source.NewMemorySource("demo-source", sourceRecords, pool)
	scheduler := schedule.NewScheduler(
		planner,
		src,
		engine,
		offsets,
		notifier,
		16,
		5*time.Second,
		schedule.NewRetryPolicy(2),
		keys,
		tracker,
	)
	return &app{
		scheduler:   scheduler,
		diffs:       diffStore,
		offsets:     offsets,
		notifier:    notifier,
		counting:    counting,
		channel:     channel,
		pool:        pool,
		keys:        keys,
		targetStore: targetStore,
		snapshots:   target.NewSnapshotService(targetStore).WithLog(targetLog),
		targetLog:   targetLog,
		journal:     journal,
		history:     history,
		tracker:     tracker,
		src:         src,
		tasks:       make(map[string]*schedule.Task),
	}, nil
}

// runDemo 启动后执行一次批量对账作为演示。
func (a *app) runDemo(ctx context.Context) {
	tasks := []*schedule.Task{
		{ID: "demo-full-1", State: schedule.StatePending},
		{ID: "demo-full-2", State: schedule.StatePending},
	}
	batch, err := schedule.RunBatch(ctx, schedule.NewRunner(2), a.scheduler, tasks)
	if err != nil {
		fmt.Printf("demo reconcile failed: %v\n", err)
		return
	}
	fmt.Printf("demo reconcile done: %s\n", compare.Describe(batch.Total))
}

// routes 注册全部 HTTP 路由。
func (a *app) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleMonitor)
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/reconcile/run", a.handleRun)
	mux.HandleFunc("/api/reconcile/status", a.handleStatus)
	mux.HandleFunc("/api/diffs", a.handleDiffs)
	mux.HandleFunc("/api/offsets", a.handleOffsets)
	mux.HandleFunc("/api/offsets/switch-incremental", a.handleSwitchIncremental)
	mux.HandleFunc("/api/source/stats", a.handleSourceStats)
	mux.HandleFunc("/api/source/reads", a.handleSourceReads)
	mux.HandleFunc("/api/notify/summary", a.handleNotifySummary)
	mux.HandleFunc("/api/notify/history", a.handleNotifyHistory)
	mux.HandleFunc("/api/target/status", a.handleTargetStatus)
	mux.HandleFunc("/api/target/update", a.handleTargetUpdate)
	mux.HandleFunc("/api/offsets/history", a.handleOffsetsHistory)
	return mux
}

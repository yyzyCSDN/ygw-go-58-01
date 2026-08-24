package model

import "time"

// DiffKind 表示差异类型。
type DiffKind int

const (
	DiffMissing DiffKind = iota
	DiffMismatch
	DiffExtra
)

// String 返回差异类型的稳定名称。
func (k DiffKind) String() string {
	switch k {
	case DiffMissing:
		return "missing"
	case DiffMismatch:
		return "mismatch"
	case DiffExtra:
		return "extra"
	default:
		return "unknown"
	}
}

// Entry 表示一条差异记录。
type Entry struct {
	Key           string
	Kind          DiffKind
	SourceVersion int64
	TargetVersion int64
	SourceData    string
	TargetData    string
}

// DedupKey 返回差异的唯一键。
func (e Entry) DedupKey() string {
	return e.Key + "@" + e.Kind.String()
}

// Result 表示一个窗口的对账结果。
type Result struct {
	Window     Window
	Total      int
	Missing    int
	Mismatch   int
	Extra      int
	Entries    []Entry
	StartedAt  time.Time
	FinishedAt time.Time
}

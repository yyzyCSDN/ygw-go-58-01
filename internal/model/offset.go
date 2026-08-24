package model

// Position 表示对账位点。
type Position struct {
	Phase       Phase
	WindowIndex int
	Key         string
	Completed   bool
}

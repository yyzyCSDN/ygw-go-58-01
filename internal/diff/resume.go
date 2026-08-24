package diff

import "reconcilesvc/internal/model"

// ResumeInfo 描述重试续对时的已提交位点。
type ResumeInfo struct {
	CommittedKey string
	Phase        model.Phase
}

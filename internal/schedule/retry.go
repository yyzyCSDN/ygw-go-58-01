package schedule

// RetryPolicy 描述对账失败后的重试策略。
type RetryPolicy struct {
	MaxAttempts int
}

// NewRetryPolicy 创建重试策略。
func NewRetryPolicy(maxAttempts int) RetryPolicy {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return RetryPolicy{MaxAttempts: maxAttempts}
}

// ShouldRetry 判断当前尝试次数是否还能重试。
func (p RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < p.MaxAttempts
}

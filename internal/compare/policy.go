package compare

// Options 控制比对行为。
type Options struct {
	CompareData bool
}

// DefaultOptions 返回默认比对选项：同时比较版本与数据。
func DefaultOptions() Options {
	return Options{CompareData: true}
}

package model

// Record 表示对账数据中的一条记录。
type Record struct {
	Key     string
	Version int64
	Data    string
}

// Snapshot 表示目标侧某个版本的数据快照。
type Snapshot struct {
	Version int64
	Records map[string]*Record
}

// Get 返回键对应的记录；记录不存在或为空时返回 nil, false。
func (s *Snapshot) Get(key string) (*Record, bool) {
	if s == nil || s.Records == nil {
		return nil, false
	}
	rec, ok := s.Records[key]
	if !ok || rec == nil {
		return nil, false
	}
	return rec, true
}

// Keys 返回快照中的全部键。
func (s *Snapshot) Keys() []string {
	if s == nil || s.Records == nil {
		return nil
	}
	keys := make([]string, 0, len(s.Records))
	for key := range s.Records {
		keys = append(keys, key)
	}
	return keys
}

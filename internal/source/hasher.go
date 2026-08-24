package source

import "github.com/cespare/xxhash/v2"

// Bucket 将键散列到 [0, buckets) 的桶，用于窗口切分。
func Bucket(key string, buckets int) int {
	if buckets <= 0 {
		return 0
	}
	return int(xxhash.Sum64String(key) % uint64(buckets))
}

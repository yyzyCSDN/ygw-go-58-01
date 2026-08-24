package main

import (
	"fmt"
	"strconv"

	"reconcilesvc/internal/model"
)

// seedSourceRecords 构造源侧演示数据。
func seedSourceRecords() []*model.Record {
	records := make([]*model.Record, 0, 260)
	for i := 0; i < 260; i++ {
		key := fmt.Sprintf("rec-%04d", i)
		version := int64(1 + i%5)
		data := fmt.Sprintf("payload-%03d", i)
		if i%17 == 0 {
			version++
		}
		records = append(records, &model.Record{Key: key, Version: version, Data: data})
	}
	return records
}

// seedTargetRecords 构造目标侧演示数据，与源侧存在少量差异。
func seedTargetRecords() map[string]*model.Record {
	records := make(map[string]*model.Record, 260)
	for i := 0; i < 260; i++ {
		key := fmt.Sprintf("rec-%04d", i)
		version := int64(1 + i%5)
		data := fmt.Sprintf("payload-%03d", i)
		if i%13 == 0 {
			version--
		}
		if i%29 == 0 {
			data = data + "-stale"
		}
		records[key] = &model.Record{Key: key, Version: version, Data: data}
	}
	return records
}

// nextTargetRecords 构造目标更新后的数据，模拟目标侧版本提升。
func nextTargetRecords(prev map[string]*model.Record) map[string]*model.Record {
	next := make(map[string]*model.Record, len(prev)+1)
	index := 0
	for key, rec := range prev {
		cp := *rec
		if index%7 == 0 {
			cp.Version++
			cp.Data = cp.Data + "-v" + strconv.FormatInt(cp.Version, 10)
		}
		next[key] = &cp
		index++
	}
	next["rec-new-01"] = &model.Record{Key: "rec-new-01", Version: 1, Data: "added-after-update"}
	return next
}

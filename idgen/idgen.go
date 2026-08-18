// Package idgen 提供进程级雪花 ID（sonyflake），对外一律以十进制字符串暴露（docs/04 §1.2：ID 一律字符串）。
package idgen

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var (
	once sync.Once
	sf   *sonyflake.Sonyflake
)

func instance() *sonyflake.Sonyflake {
	once.Do(func() {
		sf = sonyflake.NewSonyflake(sonyflake.Settings{
			StartTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		})
	})
	return sf
}

// NextID 返回全局递增趋势的雪花 ID。
func NextID() (int64, error) {
	id, err := instance().NextID()
	if err != nil {
		return 0, fmt.Errorf("idgen: %w", err)
	}
	return int64(id), nil
}

// NextString 返回字符串形式 ID；生成失败时 panic（仅在时钟严重回拨时发生，属不可恢复故障）。
func NextString() string {
	id, err := NextID()
	if err != nil {
		panic(err)
	}
	return strconv.FormatInt(id, 10)
}

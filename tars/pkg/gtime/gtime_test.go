package gtime

import (
	"testing"
	"time"
)

func BenchmarkGtime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CurrUnixTime
		_ = CurrDateTime
		_ = CurrDateHour
		_ = CurrDateDay
	}
}

func BenchmarkTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now := time.Now()
		_ = now.Format("2006-01-02 15:04:05")
		_ = now.Format("2006010215")
		_ = now.Format("20060102")
	}
}

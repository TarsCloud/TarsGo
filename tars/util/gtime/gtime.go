package gtime

import "time"

var (
	// CurrUnixTime is current unix time
	CurrUnixTime int64
	CurrDateTime string
	CurrDateHour string
	CurrDateDay  string
)

func init() {
	now := time.Now()
	CurrUnixTime = now.Unix()
	CurrDateTime = now.Format("2006-01-02 15:04:05")
	CurrDateHour = now.Format("2006010215")
	CurrDateDay = now.Format("20060102")
	go func() {
		tm := time.NewTimer(time.Second)
		for {
			now := time.Now()
			d := time.Second - time.Duration(now.Nanosecond())
			tm.Reset(d)
			<-tm.C
			now = time.Now()
			CurrUnixTime = now.Unix()
			CurrDateTime = now.Format("2006-01-02 15:04:05")
			CurrDateHour = now.Format("2006010215")
			CurrDateDay = now.Format("20060102")
		}
	}()
}

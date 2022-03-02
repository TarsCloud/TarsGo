package trace

import (
	"github.com/google/uuid"
	"strconv"
	"strings"
)

type TraceContext struct {
	traceType    int    // 取值范围0-15， 0 不用打参数， 其他情况按位做控制开关，从低位到高位分别控制CS、CR、SR、SS，为1则输出对应参数
	paramMaxLen  uint   // 业务接口参数最大长度，如果超过该值，那么不输出参数，用特殊串标记 {"trace_param_over_max_len":true}
	traceID      string // traceID
	spanID       string // spanID
	parentSpanID string // 父spanID
}

type TraceData struct {
	TraceCall    bool
	TraceContext *TraceContext
}

type ESpanType uint8

type ENeedParam int

var (
	EstCS ESpanType = 1
	EstCR ESpanType = 2
	EstSR ESpanType = 4
	EstSS ESpanType = 8
	EstTS ESpanType = 9
	EstTE ESpanType = 10

	EnpNo            ENeedParam = 0
	EnpNormal        ENeedParam = 1
	EnpOverMaxLen    ENeedParam = 2
	traceParamMaxLen uint       = 1 // 默认1K
)

const (
	TraceAnnotationTS = "ts"
	TraceAnnotationTE = "te"
	TraceAnnotationCS = "cs"
	TraceAnnotationCR = "cr"
	TraceAnnotationSR = "sr"
	TraceAnnotationSS = "ss"
)

func NewTraceData() *TraceData {
	return &TraceData{
		TraceContext: newTraceContext(),
	}
}

func newTraceContext() *TraceContext {
	return &TraceContext{}
}

// Init
// key 分两种情况，1.rpc调用； 2.异步回调
// eg: f.2-ee824ad0eb4dacf56b29d230a229c584|030019ac000010796162bc5900000021|030019ac000010796162bc5900000021
func (c *TraceContext) Init(traceKey string) bool {
	traceKeys := strings.Split(traceKey, "|")
	if len(traceKeys) == 2 {
		c.traceID = traceKeys[0]
		c.parentSpanID = traceKeys[1]
		c.spanID = ""
		c.traceType, c.paramMaxLen = TraceContextInitType(c.traceID)
		return true
	} else if len(traceKeys) == 3 {
		c.traceID = traceKeys[0]
		c.parentSpanID = traceKeys[1]
		c.spanID = traceKeys[2]
		c.traceType, c.paramMaxLen = TraceContextInitType(c.traceID)
		return true
	} else {
		c.Reset()
	}
	return false
}

func (c *TraceContext) open(traceID string) bool {
	if len(traceID) > 0 {
		c.traceID = traceID
		c.parentSpanID = ""
		c.spanID = ""
		c.traceType, c.paramMaxLen = TraceContextInitType(c.traceID)
		return true
	}
	return false
}

// TraceContextInitType parse type and maxLen
func TraceContextInitType(tid string) (tType int, maxLen uint) {
	maxLen = GetTraceParamMaxLen()
	pos := strings.Index(tid, "-")
	if pos != -1 {
		flags := strings.Split(tid[:pos], ".")
		if len(flags) >= 1 {
			int64Type, _ := strconv.ParseInt(flags[0], 16, 32)
			tType = int(int64Type)
		}
		if len(flags) >= 2 {
			uint64Len, _ := strconv.ParseUint(flags[1], 10, 32)
			if maxLen < uint(uint64Len) {
				maxLen = uint(uint64Len)
			}
		}
	}
	if tType < 0 || tType > 15 {
		tType = 0
	}
	return tType, maxLen
}

func (c *TraceContext) Reset() {
	c.traceID = ""
	c.spanID = ""
	c.parentSpanID = ""
	c.traceType = 0
	c.paramMaxLen = 1
}

// newSpan 生成spanId
func (c *TraceContext) newSpan() {
	c.spanID = uuid.NewString()
	if len(c.parentSpanID) == 0 {
		c.parentSpanID = c.spanID
	}
}

func (c *TraceContext) getKey(es ESpanType) string {
	switch es {
	case EstCS, EstCR, EstTS, EstTE:
		return c.traceID + "|" + c.spanID + "|" + c.parentSpanID
	case EstSR, EstSS:
		return c.traceID + "|" + c.parentSpanID + "|*"
	}
	return ""
}

func (c *TraceContext) getKeyFull(full bool) string {
	if full {
		return c.traceID + "|" + c.spanID + "|" + c.parentSpanID
	}
	return c.traceID + "|" + c.spanID
}

// TraceContextNeedParam
// return: 0 不需要参数， 1：正常打印参数， 2：参数太长返回默认串
func TraceContextNeedParam(es ESpanType, tType int, len uint, maxLen uint) ENeedParam {
	if es == EstTS {
		es = EstCS
	} else if es == EstTE {
		es = EstCR
	}
	if (int(es) & tType) == 0 {
		return EnpNo
	} else if len > maxLen*1024 {
		return EnpOverMaxLen
	}
	return EnpNormal
}

// GetTraceKey 获取 traceKey
func (t *TraceData) GetTraceKey(es ESpanType) string {
	return t.TraceContext.getKey(es)
}

// GetTraceKeyFull 获取 traceKey
func (t *TraceData) GetTraceKeyFull(full bool) string {
	return t.TraceContext.getKeyFull(full)
}

// NewSpan 获取 traceKey
func (t *TraceData) NewSpan() {
	t.TraceContext.newSpan()
}

// InitTrace 获取 traceKey
func (t *TraceData) InitTrace(traceKey string) bool {
	return t.TraceContext.Init(traceKey)
}

// 获取 trace type
func (t *TraceData) getTraceType() int {
	return t.TraceContext.traceType
}

// NeedTraceParam 控制参数打印
func (t *TraceData) NeedTraceParam(es ESpanType, len uint) ENeedParam {
	return TraceContextNeedParam(es, t.TraceContext.traceType, len, t.TraceContext.paramMaxLen)
}

// OpenTrace 业务主动打开调用链
// @param traceFlag: 调用链日志输出参数控制，取值范围0-15， 0 不用打参数， 其他情况按位做控制开关，从低位到高位分别控制CS、CR、SR、SS，为1则输出对应参数
// @param maxLen: 参数输出最大长度， 不传或者默认0， 则按服务模板默认取值
func (t *TraceData) OpenTrace(traceFlag int, maxLen uint) bool {
	traceID := uuid.NewString()
	if maxLen > 0 {
		traceID = strconv.FormatInt(int64(traceFlag), 16) + "." + strconv.Itoa(int(maxLen)) + "-" + traceID
	} else {
		traceID = strconv.FormatInt(int64(traceFlag), 16) + "-" + traceID
	}
	t.TraceCall = t.TraceContext.open(traceID)
	return t.TraceCall
}

// NeedTraceParam 控制参数打印
func NeedTraceParam(es ESpanType, traceID string, len uint) ENeedParam {
	tType, maxLen := TraceContextInitType(traceID)
	return TraceContextNeedParam(es, tType, len, maxLen)
}

// SetTraceParamMaxLen 设置控制参数长度
func SetTraceParamMaxLen(len uint) {
	// 最最大保护，不超过10M
	if len < 1024*10 {
		traceParamMaxLen = len
	}
}

// GetTraceParamMaxLen 获取控制参数长度
func GetTraceParamMaxLen() uint {
	return traceParamMaxLen
}

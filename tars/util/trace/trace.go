package trace

import (
	"strconv"
	"strings"
)

// SpanContext 调用链追踪信息
type SpanContext struct {
	traceType    int    // 取值范围0-15， 0 不用打参数， 其他情况按位做控制开关，从低位到高位分别控制CS、CR、SR、SS，为1则输出对应参数
	paramMaxLen  uint   // 业务接口参数最大长度，如果超过该值，那么不输出参数，用特殊串标记 {"trace_param_over_max_len":true}
	traceID      string // traceID
	spanID       string // spanID
	parentSpanID string // 父spanID
}

type Trace struct {
	call bool         //标识当前调用是否需要调用链追踪，默认不打开
	sc   *SpanContext //调用链追踪信息
}

type SpanType uint8

// NeedParam 是否输出参数
type NeedParam int

const (
	EstCS SpanType = 1
	EstCR SpanType = 2
	EstSR SpanType = 4
	EstSS SpanType = 8
	EstTS SpanType = 9
	EstTE SpanType = 10

	EnpNo         NeedParam = 0
	EnpNormal     NeedParam = 1
	EnpOverMaxLen NeedParam = 2

	AnnotationTS = "ts"
	AnnotationTE = "te"
	AnnotationCS = "cs"
	AnnotationCR = "cr"
	AnnotationSR = "sr"
	AnnotationSS = "ss"
)

var (
	idGenerator           = newGenerator()
	traceParamMaxLen uint = 1 // 默认1K
)

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

type SpanContextOption func(*SpanContext)

func WithTraceKey(traceKey string) SpanContextOption {
	return func(sc *SpanContext) {
		sc.Init(traceKey)
	}
}

func NewSpanContext(opts ...SpanContextOption) *SpanContext {
	sc := &SpanContext{}
	for _, opt := range opts {
		opt(sc)
	}
	return sc
}

// Init
// key 分两种情况，1.rpc调用； 2.异步回调
// eg: f.2-ee824ad0eb4dacf56b29d230a229c584|030019ac000010796162bc5900000021|030019ac000010796162bc5900000021
func (c *SpanContext) Init(traceKey string) bool {
	traceKeys := strings.Split(traceKey, "|")
	if len(traceKeys) == 2 {
		c.traceID = traceKeys[0]
		c.parentSpanID = traceKeys[1]
		c.spanID = ""
		c.traceType, c.paramMaxLen = initType(c.traceID)
		return true
	} else if len(traceKeys) == 3 {
		c.traceID = traceKeys[0]
		c.parentSpanID = traceKeys[1]
		c.spanID = traceKeys[2]
		c.traceType, c.paramMaxLen = initType(c.traceID)
		return true
	} else {
		c.Reset()
	}
	return false
}

func (c *SpanContext) Open(traceID string) bool {
	if len(traceID) > 0 {
		c.traceID = traceID
		c.parentSpanID = ""
		c.spanID = ""
		c.traceType, c.paramMaxLen = initType(c.traceID)
		return true
	}
	return false
}

func (c *SpanContext) Reset() {
	c.traceID = ""
	c.spanID = ""
	c.parentSpanID = ""
	c.traceType = 0
	c.paramMaxLen = 1
}

// NewSpan 生成spanId
func (c *SpanContext) NewSpan() {
	c.spanID = idGenerator.NewSpanID()
	if len(c.parentSpanID) == 0 {
		c.parentSpanID = c.spanID
	}
}

func (c *SpanContext) Key(es SpanType) string {
	switch es {
	case EstCS, EstCR, EstTS, EstTE:
		return c.traceID + "|" + c.spanID + "|" + c.parentSpanID
	case EstSR, EstSS:
		return c.traceID + "|" + c.parentSpanID + "|*"
	}
	return ""
}

func (c *SpanContext) FullKey(full bool) string {
	if full {
		return c.traceID + "|" + c.spanID + "|" + c.parentSpanID
	}
	return c.traceID + "|" + c.spanID
}

func (c *SpanContext) TraceID() string {
	return c.traceID
}

func (c *SpanContext) SpanID() string {
	return c.spanID
}

func (c *SpanContext) ParentSpanID() string {
	return c.parentSpanID
}

func (c *SpanContext) TraceType() int {
	return c.traceType
}

func New() *Trace {
	return &Trace{
		sc: NewSpanContext(),
	}
}

// SetCall set call value
func (t *Trace) SetCall(call bool) {
	t.call = call
}

// Call return call
func (t *Trace) Call() bool {
	return t.call
}

// GetTraceKey 获取 traceKey
func (t *Trace) GetTraceKey(es SpanType) string {
	return t.sc.Key(es)
}

// GetTraceFullKey 获取 traceKey
func (t *Trace) GetTraceFullKey(full bool) string {
	return t.sc.FullKey(full)
}

// SpanContext 获取 SpanContext
func (t *Trace) SpanContext() *SpanContext {
	return t.sc
}

// NewSpan 生成 spanId
func (t *Trace) NewSpan() {
	t.sc.NewSpan()
}

// InitTrace 获取 traceKey
func (t *Trace) InitTrace(traceKey string) bool {
	return t.sc.Init(traceKey)
}

// GetTraceType 获取 trace type
func (t *Trace) GetTraceType() int {
	return t.sc.TraceType()
}

// NeedTraceParam 控制参数打印
func (t *Trace) NeedTraceParam(es SpanType, len uint) NeedParam {
	return getNeedParam(es, t.sc.traceType, len, t.sc.paramMaxLen)
}

// OpenTrace 业务主动打开调用链
// @param traceFlag: 调用链日志输出参数控制，取值范围0-15， 0 不用打参数， 其他情况按位做控制开关，从低位到高位分别控制CS、CR、SR、SS，为1则输出对应参数
// @param maxLen: 参数输出最大长度， 不传或者默认0， 则按服务模板默认取值
func (t *Trace) OpenTrace(traceFlag int, maxLen uint) bool {
	traceID := idGenerator.NewTraceID()
	if maxLen > 0 {
		traceID = strconv.FormatInt(int64(traceFlag), 16) + "." + strconv.Itoa(int(maxLen)) + "-" + traceID
	} else {
		traceID = strconv.FormatInt(int64(traceFlag), 16) + "-" + traceID
	}
	t.call = t.sc.Open(traceID)
	return t.call
}

// NeedTraceParam 控制参数打印
func NeedTraceParam(es SpanType, traceID string, len uint) NeedParam {
	typ, maxLen := initType(traceID)
	return getNeedParam(es, typ, len, maxLen)
}

// initType parse type and maxLen
func initType(tid string) (typ int, maxLen uint) {
	maxLen = GetTraceParamMaxLen()
	pos := strings.Index(tid, "-")
	if pos != -1 {
		flags := strings.Split(tid[:pos], ".")
		if len(flags) >= 1 {
			int64Type, err := strconv.ParseInt(flags[0], 16, 32)
			if err == nil {
				typ = int(int64Type)
			}
		}
		if len(flags) >= 2 {
			uint64Len, err := strconv.ParseUint(flags[1], 10, 32)
			if err == nil && maxLen < uint(uint64Len) {
				maxLen = uint(uint64Len)
			}
		}
	}
	if typ < 0 || typ > 15 {
		typ = 0
	}
	return typ, maxLen
}

// getNeedParam
// return: 0 不需要参数， 1：正常打印参数， 2：参数太长返回默认串
func getNeedParam(es SpanType, typ int, len uint, maxLen uint) NeedParam {
	if es == EstTS {
		es = EstCS
	} else if es == EstTE {
		es = EstCR
	}
	if (int(es) & typ) == 0 {
		return EnpNo
	} else if len > maxLen*1024 {
		return EnpOverMaxLen
	}
	return EnpNormal
}

package current

import (
	"context"
	"net"

	tarstrace "github.com/TarsCloud/TarsGo/tars/util/trace"
)

type tarsCurrentKey int64

var tcKey = tarsCurrentKey(0x484900)

// Current contains message for the specify request.
// This current is used for server side.
type Current struct {
	uuid        string
	clientIP    string
	clientPort  string
	recvPkgTs   int64
	cPacketType int8
	reqStatus   map[string]string
	resStatus   map[string]string
	reqContext  map[string]string
	resContext  map[string]string
	needDyeing  bool
	dyeingUser  string
	trace       *tarstrace.Trace

	rawConn net.Conn
	udpAddr *net.UDPAddr
}

// NewCurrent return a Current point.
func newCurrent() *Current {
	return &Current{}
}

// ContextWithTarsCurrent set TarsCurrent
func ContextWithTarsCurrent(ctx context.Context) context.Context {
	tc := newCurrent()
	ctx = context.WithValue(ctx, tcKey, tc)
	return ctx
}

// GetClientIPFromContext gets the client ip from the context.
func GetClientIPFromContext(ctx context.Context) (string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.clientIP, ok
	}
	return "", ok
}

// SetClientIPWithContext set Client IP to the tars current.
func SetClientIPWithContext(ctx context.Context, IP string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.clientIP = IP
	}
	return ok
}

// GetClientPortFromContext gets the client ip from the context.
func GetClientPortFromContext(ctx context.Context) (string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.clientPort, ok
	}
	return "", ok
}

// SetClientPortWithContext set client port to the tars current.
func SetClientPortWithContext(ctx context.Context, port string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.clientPort = port
	}
	return ok
}

// currentFromContext gets current from the context
func currentFromContext(ctx context.Context) (*Current, bool) {
	tc, ok := ctx.Value(tcKey).(*Current)
	return tc, ok
}

func SetUUIDWithContext(ctx context.Context, uid string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.uuid = uid
	}
	return ok
}

// GetUUID get current uuid using for connect uniqe mark
func GetUUID(ctx context.Context) (uuid string, ret bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.uuid, true
	}
	return
}

// SetResponseStatus set the response package' status .
func SetResponseStatus(ctx context.Context, s map[string]string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.resStatus = s
	}
	return ok
}

// GetResponseStatus get response status set by user.
func GetResponseStatus(ctx context.Context) (map[string]string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.resStatus, ok
	}
	return nil, ok
}

// SetResponseContext set the response package' context .
func SetResponseContext(ctx context.Context, c map[string]string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.resContext = c
	}
	return ok
}

// GetResponseContext get response context set by user.
func GetResponseContext(ctx context.Context) (map[string]string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.resContext, ok
	}
	return nil, ok
}

// SetRequestStatus set the request package' status .
func SetRequestStatus(ctx context.Context, s map[string]string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.reqStatus = s
	}
	return ok
}

// GetRequestStatus get request status set by user.
func GetRequestStatus(ctx context.Context) (map[string]string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.reqStatus, ok
	}
	return nil, ok
}

// SetRequestContext set the request package' context .
func SetRequestContext(ctx context.Context, c map[string]string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.reqContext = c
	}
	return ok
}

// GetRequestContext get request context set by user.
func GetRequestContext(ctx context.Context) (map[string]string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.reqContext, ok
	}
	return nil, ok
}

// GetRecvPkgTsFromContext gets the recvTs from the context.
func GetRecvPkgTsFromContext(ctx context.Context) (int64, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.recvPkgTs, ok
	}
	return 0, ok
}

// SetRecvPkgTsFromContext set recv Ts to the tars current.
func SetRecvPkgTsFromContext(ctx context.Context, recvPkgTs int64) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.recvPkgTs = recvPkgTs
	}
	return ok
}

// GetPacketTypeFromContext gets the PacketType from the context.
func GetPacketTypeFromContext(ctx context.Context) (int8, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.cPacketType, ok
	}
	return 0, ok
}

// SetPacketTypeFromContext set PacketType to the tars current.
func SetPacketTypeFromContext(ctx context.Context, cPacketType int8) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.cPacketType = cPacketType
	}
	return ok
}

// GetReqStatusValue get req status from current in context
func GetReqStatusValue(ctx context.Context, key string) (string, bool) {
	reqStatus, ok := GetRequestStatus(ctx)
	if ok && reqStatus != nil {
		value, reqOk := reqStatus[key]
		return value, reqOk
	}
	return "", ok
}

// SetReqStatusValue set req status of current of context
func SetReqStatusValue(ctx context.Context, key string, value string) bool {
	reqStatus, ok := GetRequestStatus(ctx)
	if ok {
		if reqStatus == nil {
			reqStatus = make(map[string]string)
		}
		reqStatus[key] = value

		ok := SetRequestStatus(ctx, reqStatus)
		return ok
	}
	return ok
}

const StatusDyedKey = "STATUS_DYED_KEY"

// GetDyeingKey gets dyeing key from the context.
func GetDyeingKey(ctx context.Context) (string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		if tc.reqStatus != nil {
			if dyeingKey, exists := tc.reqStatus[StatusDyedKey]; exists {
				return dyeingKey, true
			}
		}
	}

	return "", false
}

// SetDyeingKey set dyeing key to the tars current.
func SetDyeingKey(ctx context.Context, dyeingKey string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		if tc.reqStatus == nil {
			tc.reqStatus = make(map[string]string)
		}
		tc.reqStatus[StatusDyedKey] = dyeingKey
		tc.needDyeing = true
	}
	return ok
}

// GetDyeingUser gets dyeing user from the context.
func GetDyeingUser(ctx context.Context) (string, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.dyeingUser, ok
	}
	return "", ok
}

// SetDyeingUser set dyeing user to the tars current.
func SetDyeingUser(ctx context.Context, user string) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.dyeingUser = user
	}
	return ok
}

const StatusTraceKey = "STATUS_TRACE_KEY"

// OpenTarsTrace 开启trace
func OpenTarsTrace(ctx context.Context, traceParams bool) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		trace := tarstrace.New()
		if traceParams {
			trace.OpenTrace(15, 0)
		} else {
			trace.OpenTrace(0, 0)
		}
		tc.trace = trace
	}
	return ok
}

// InitTarsTrace init trace data from the trace key.
func InitTarsTrace(ctx context.Context, traceKey string) bool {
	_, ok := currentFromContext(ctx)
	if ok {
		trace := tarstrace.New()
		if !trace.InitTrace(traceKey) {
			return false
		}
		trace.SetCall(true)
		return SetTarsTrace(ctx, trace)
	}
	return ok
}

// GetTarsTrace get trace data from the context.
func GetTarsTrace(ctx context.Context) (*tarstrace.Trace, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		if tc.trace != nil {
			return tc.trace, true
		}
	}
	return nil, false
}

// SetTarsTrace set trace data to the tars current.
func SetTarsTrace(ctx context.Context, trace *tarstrace.Trace) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.trace = trace
	}
	return ok
}

// GetRawConn get the raw tcp/udp connection from the context.
func GetRawConn(ctx context.Context) (net.Conn, *net.UDPAddr, bool) {
	tc, ok := currentFromContext(ctx)
	if ok {
		return tc.rawConn, tc.udpAddr, true
	}
	return nil, nil, false
}

// SetRawConnWithContext set tcp/udp connection to the tars current.
func SetRawConnWithContext(ctx context.Context, conn net.Conn, udpAddr *net.UDPAddr) bool {
	tc, ok := currentFromContext(ctx)
	if ok {
		tc.rawConn = conn
		tc.udpAddr = udpAddr
	}
	return ok
}

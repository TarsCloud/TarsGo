package current

import "context"

// ctckey is key for tars client current.
var ctcKey = tarsCurrentKey(0x484901)

// ClientCurrent for passing client side info for tars framework.
type ClientCurrent struct {
	isHash    bool
	hashCode  uint32
	hashType  int
	isTimeout bool
	timeout   int //in ms

	serverIP   string
	serverPort string
}

func newCientCurrent() *ClientCurrent {
	return &ClientCurrent{
		isHash:    false,
		isTimeout: false,
	}
}

// clientCurrentFromContext gets  client current from the context
func clientCurrentFromContext(ctx context.Context) (*ClientCurrent, bool) {
	tc, ok := ctx.Value(ctcKey).(*ClientCurrent)
	return tc, ok
}

// ContextWithClientCurrent set ClientCurrent to the context.
func ContextWithClientCurrent(ctx context.Context) context.Context {
	cc := newCientCurrent()
	ctx = context.WithValue(ctx, ctcKey, cc)
	return ctx
}

// SetClientHash sets the client hash code and hash type
func SetClientHash(ctx context.Context, hashType int, hashCode uint32) bool {
	cc, ok := clientCurrentFromContext(ctx)
	if ok {
		cc.isHash = true
		cc.hashType = hashType
		cc.hashCode = hashCode
	}
	return ok
}

// GetClientHash returns the client hash info
func GetClientHash(ctx context.Context) (isOk bool, hashType int, hashCode uint32, isHash bool) {
	cc, ok := clientCurrentFromContext(ctx)
	if ok {
		return ok, cc.hashType, cc.hashCode, cc.isHash
	}
	return ok, 0, 0, false
}

// SetClientTimeout sets the client timeout
func SetClientTimeout(ctx context.Context, timeout int) bool {
	cc, ok := clientCurrentFromContext(ctx)
	if ok {
		cc.isTimeout = true
		cc.timeout = timeout
	}
	return ok
}

// GetClientTimeout returns the timeout sets for the client side.
func GetClientTimeout(ctx context.Context) (isOk bool, timeout int, isTimeout bool) {
	cc, ok := clientCurrentFromContext(ctx)
	if ok {
		return ok, cc.timeout, cc.isTimeout
	}
	return ok, 0, false
}

// GetServerIPFromContext gets the server ip from the context.
func GetServerIPFromContext(ctx context.Context) (string, bool) {
	tc, ok := clientCurrentFromContext(ctx)
	if ok {
		return tc.serverIP, ok
	}
	return "", ok
}

// SetServerIPWithContext set Server IP to the tars current.
func SetServerIPWithContext(ctx context.Context, IP string) bool {
	tc, ok := clientCurrentFromContext(ctx)
	if ok {
		tc.serverIP = IP
	}
	return ok
}

// GetServerPortFromContext gets the server ip from the context.
func GetServerPortFromContext(ctx context.Context) (string, bool) {
	tc, ok := clientCurrentFromContext(ctx)
	if ok {
		return tc.serverPort, ok
	}
	return "", ok
}

// SetServerPortWithContext set server port to the tars current.
func SetServerPortWithContext(ctx context.Context, port string) bool {
	tc, ok := clientCurrentFromContext(ctx)
	if ok {
		tc.serverPort = port
	}
	return ok
}

package tars

import (
	"context"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)

// Invoke is used for Invoke tars server service
type Invoke func(ctx context.Context, msg *Message, timeout time.Duration) (err error)

// DispatchReporter is the reporter in server-side dispatch, and will be used in logging
type DispatchReporter func(ctx context.Context, req []interface{}, rsp []interface{}, returns []interface{})

// Dispatch server side Dispatch
type Dispatch func(context.Context, interface{}, *requestf.RequestPacket, *requestf.ResponsePacket, bool) error

// ServerFilter is used for add Filter for server dispatcher ,for implementing plugins like opentracing.
type ServerFilter func(ctx context.Context, d Dispatch, f interface{},
	req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error)

// ClientFilter is used for filter request & response for client, for implementing plugins like opentracing
type ClientFilter func(ctx context.Context, msg *Message, invoke Invoke, timeout time.Duration) (err error)

// ServerFilterMiddleware is used for add multiple filter middlewares for dispatcher, for using multiple filter such as breaker, rate limit and trace.
type ServerFilterMiddleware func(next ServerFilter) ServerFilter

// ClientFilterMiddleware is used for add multiple filter middleware for client, for using multiple filter such as breaker, rate limit and trace.
type ClientFilterMiddleware func(next ClientFilter) ClientFilter

type filters struct {
	// client
	cf     ClientFilter
	cfOnce sync.Once
	cfms   []ClientFilterMiddleware

	// server
	sf     ServerFilter
	sfOnce sync.Once
	sfms   []ServerFilterMiddleware
}

// UseClientFilterMiddleware uses the client filter middleware.
func (f *filters) UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) {
	f.cfms = append(f.cfms, cfm...)
}

func (f *filters) getMiddlewareClientFilter() ClientFilter {
	if len(f.cfms) <= 0 {
		return nil
	}

	f.cfOnce.Do(func() {
		cf := func(ctx context.Context, msg *Message, invoke Invoke, timeout time.Duration) (err error) {
			return invoke(ctx, msg, timeout)
		}

		for i := len(f.cfms) - 1; i >= 0; i-- {
			cf = f.cfms[i](cf)
		}
		f.cf = cf
	})
	return f.cf
}

// UseServerFilterMiddleware uses the server filter middleware.
func (f *filters) UseServerFilterMiddleware(sfm ...ServerFilterMiddleware) {
	f.sfms = append(f.sfms, sfm...)
}

func (f *filters) getMiddlewareServerFilter() ServerFilter {
	if len(f.sfms) <= 0 {
		return nil
	}

	f.sfOnce.Do(func() {
		sf := func(ctx context.Context, d Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
			return d(ctx, f, req, resp, withContext)
		}

		for i := len(f.sfms) - 1; i >= 0; i-- {
			sf = f.sfms[i](sf)
		}
		f.sf = sf
	})
	return f.sf
}

// RegisterDispatchReporter registers the server dispatch reporter
func RegisterDispatchReporter(f DispatchReporter) {
	defaultApp.RegisterDispatchReporter(f)
}

// GetDispatchReporter returns the dispatch reporter
func GetDispatchReporter() DispatchReporter {
	return defaultApp.GetDispatchReporter()
}

// UseClientFilterMiddleware uses the client filter middleware.
func UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) {
	defaultApp.UseClientFilterMiddleware(cfm...)
}

// UseServerFilterMiddleware uses the server filter middleware.
func UseServerFilterMiddleware(sfm ...ServerFilterMiddleware) {
	defaultApp.UseServerFilterMiddleware(sfm...)
}

// RegisterDispatchReporter registers the server dispatch reporter
func (a *application) RegisterDispatchReporter(f DispatchReporter) {
	a.dispatchReporter = f
}

// GetDispatchReporter returns the dispatch reporter
func (a *application) GetDispatchReporter() DispatchReporter {
	return a.dispatchReporter
}

// UseClientFilterMiddleware uses the client filter middleware.
func (a *application) UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) *application {
	a.allFilters.UseClientFilterMiddleware(cfm...)
	return a
}

func (a *application) getMiddlewareClientFilter() ClientFilter {
	return a.allFilters.getMiddlewareClientFilter()
}

// UseServerFilterMiddleware uses the server filter middleware.
func (a *application) UseServerFilterMiddleware(sfm ...ServerFilterMiddleware) *application {
	a.allFilters.UseServerFilterMiddleware(sfm...)
	return a
}

func (a *application) getMiddlewareServerFilter() ServerFilter {
	return a.allFilters.getMiddlewareServerFilter()
}

package tars

import (
	"context"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)

type filters struct {
	cf      ClientFilter
	preCfs  []ClientFilter
	postCfs []ClientFilter
	cfms    []ClientFilterMiddleware

	sf      ServerFilter
	preSfs  []ServerFilter
	postSfs []ServerFilter
	sfms    []ServerFilterMiddleware
}

var allFilters = filters{nil, nil, nil, nil, nil, nil, nil, nil}
var dispatchReporter DispatchReporter

// Invoke is used for Invoke tars server service
type Invoke func(ctx context.Context, msg *Message, timeout time.Duration) (err error)

// DispatchReporter is the reporter in server-side dispatch, and will be used in logging
type DispatchReporter func(ctx context.Context, req []interface{}, rsp []interface{}, returns []interface{})

// RegisterDispatchReporter registers the server dispatch reporter
func RegisterDispatchReporter(f DispatchReporter) {
	dispatchReporter = f
}

// RegisterClientFilter  registers the Client filter , and will be executed in every request.
func RegisterClientFilter(f ClientFilter) {
	allFilters.cf = f
}

// RegisterPreClientFilter registers the client filter, and will be executed in order before every request
func RegisterPreClientFilter(f ClientFilter) {
	allFilters.preCfs = append(allFilters.preCfs, f)
}

// RegisterPostClientFilter registers the client filter, and will be executed in order after every request
func RegisterPostClientFilter(f ClientFilter) {
	allFilters.postCfs = append(allFilters.postCfs, f)
}

// Dispatch server side Dispatch
type Dispatch func(context.Context, interface{}, *requestf.RequestPacket, *requestf.ResponsePacket, bool) error

// ServerFilter is used for add Filter for server dispatcher ,for implementing plugins like opentracing.
type ServerFilter func(ctx context.Context, d Dispatch, f interface{},
	req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error)

//ClientFilter is used for filter request & response for client, for implementing plugins like opentracing
type ClientFilter func(ctx context.Context, msg *Message, invoke Invoke, timeout time.Duration) (err error)

//ServerFilterMiddleware is used for add multiple filter middlewares for dispatcher, for using multiple filter such as breaker, rate limit and trace.
type ServerFilterMiddleware func(next ServerFilter) ServerFilter

//ClientFilterMiddleware is used for add multiple filter middleware for client, for using multiple filter such as breaker, rate limit and trace.
type ClientFilterMiddleware func(next ClientFilter) ClientFilter

//UseClientFilterMiddleware uses the client filter middleware.
func UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) {
	allFilters.cfms = append(allFilters.cfms, cfm...)
}

// GetDispatchReporter returns the dispatch reporter
func GetDispatchReporter() DispatchReporter {
	return dispatchReporter
}

func getMiddlewareClientFilter() ClientFilter {
	if len(allFilters.cfms) <= 0 {
		return nil
	}

	cf := func(ctx context.Context, msg *Message, invoke Invoke, timeout time.Duration) (err error) {
		return invoke(ctx, msg, timeout)
	}

	for i := len(allFilters.cfms) - 1; i >= 0; i-- {
		cf = allFilters.cfms[i](cf)
	}

	return cf
}

// RegisterServerFilter register the server filter.
func RegisterServerFilter(f ServerFilter) {
	allFilters.sf = f
}

// RegisterPreServerFilter registers the server filter, executed in order before every request
func RegisterPreServerFilter(f ServerFilter) {
	allFilters.preSfs = append(allFilters.preSfs, f)
}

// RegisterPostServerFilter registers the server filter, executed in order after every request
func RegisterPostServerFilter(f ServerFilter) {
	allFilters.postSfs = append(allFilters.postSfs, f)
}

//UserServerFilterMiddleware uses the server filter middleware.
func UseServerFilterMiddleware(sfm ...ServerFilterMiddleware) {
	allFilters.sfms = append(allFilters.sfms, sfm...)
}

func getMiddlewareServerFilter() ServerFilter {
	if len(allFilters.sfms) <= 0 {
		return nil
	}

	sf := func(ctx context.Context, d Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
		return d(ctx, f, req, resp, withContext)
	}

	for i := len(allFilters.sfms) - 1; i >= 0; i-- {
		sf = allFilters.sfms[i](sf)
	}

	return sf
}

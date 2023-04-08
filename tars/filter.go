package tars

import (
	"context"
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
	cf      ClientFilter   // Deprecated: As of TarsGo 1.5
	preCfs  []ClientFilter // Deprecated: As of TarsGo 1.5
	postCfs []ClientFilter // Deprecated: As of TarsGo 1.5
	cfms    []ClientFilterMiddleware

	sf      ServerFilter   // Deprecated: As of TarsGo 1.5
	preSfs  []ServerFilter // Deprecated: As of TarsGo 1.5
	postSfs []ServerFilter // Deprecated: As of TarsGo 1.5
	sfms    []ServerFilterMiddleware
}

func (f *filters) registerClientFilter(cf ClientFilter) {
	f.cf = cf
}

// RegisterPreClientFilter registers the client filter, and will be executed in order before every request
func (f *filters) registerPreClientFilter(cf ClientFilter) {
	f.preCfs = append(f.preCfs, cf)
}

// RegisterPostClientFilter registers the client filter, and will be executed in order after every request
func (f *filters) registerPostClientFilter(cf ClientFilter) {
	f.postCfs = append(f.postCfs, cf)
}

// RegisterServerFilter register the server filter.
func (f *filters) registerServerFilter(sf ServerFilter) {
	f.sf = sf
}

// RegisterPreServerFilter registers the server filter, executed in order before every request
func (f *filters) registerPreServerFilter(sf ServerFilter) {
	f.preSfs = append(f.preSfs, sf)
}

// RegisterPostServerFilter registers the server filter, executed in order after every request
func (f *filters) registerPostServerFilter(sf ServerFilter) {
	f.postSfs = append(f.postSfs, sf)
}

// UseClientFilterMiddleware uses the client filter middleware.
func (f *filters) UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) {
	f.cfms = append(f.cfms, cfm...)
}

func (f *filters) getMiddlewareClientFilter() ClientFilter {
	if len(f.cfms) <= 0 {
		return nil
	}

	cf := func(ctx context.Context, msg *Message, invoke Invoke, timeout time.Duration) (err error) {
		return invoke(ctx, msg, timeout)
	}

	for i := len(f.cfms) - 1; i >= 0; i-- {
		cf = f.cfms[i](cf)
	}

	return cf
}

// UseServerFilterMiddleware uses the server filter middleware.
func (f *filters) UseServerFilterMiddleware(sfm ...ServerFilterMiddleware) {
	f.sfms = append(f.sfms, sfm...)
}

func (f *filters) getMiddlewareServerFilter() ServerFilter {
	if len(f.sfms) <= 0 {
		return nil
	}

	sf := func(ctx context.Context, d Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
		return d(ctx, f, req, resp, withContext)
	}

	for i := len(f.sfms) - 1; i >= 0; i-- {
		sf = f.sfms[i](sf)
	}

	return sf
}

// RegisterDispatchReporter registers the server dispatch reporter
func RegisterDispatchReporter(f DispatchReporter) {
	defaultApp.RegisterDispatchReporter(f)
}

// GetDispatchReporter returns the dispatch reporter
func GetDispatchReporter() DispatchReporter {
	return defaultApp.GetDispatchReporter()
}

// RegisterClientFilter  registers the Client filter , and will be executed in every request.
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseClientFilterMiddleware for the same functionality。
func RegisterClientFilter(f ClientFilter) {
	defaultApp.registerClientFilter(f)
}

// RegisterPreClientFilter registers the client filter, and will be executed in order before every request
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseClientFilterMiddleware for the same functionality。
func RegisterPreClientFilter(f ClientFilter) {
	defaultApp.registerPreClientFilter(f)
}

// RegisterPostClientFilter registers the client filter, and will be executed in order after every request
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseClientFilterMiddleware for the same functionality。
func RegisterPostClientFilter(f ClientFilter) {
	defaultApp.registerPostClientFilter(f)
}

// UseClientFilterMiddleware uses the client filter middleware.
func UseClientFilterMiddleware(cfm ...ClientFilterMiddleware) {
	defaultApp.UseClientFilterMiddleware(cfm...)
}

// RegisterServerFilter register the server filter.
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseServerFilterMiddleware for the same functionality.
func RegisterServerFilter(f ServerFilter) {
	defaultApp.registerServerFilter(f)
}

// RegisterPreServerFilter registers the server filter, executed in order before every request
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseServerFilterMiddleware for the same functionality.
func RegisterPreServerFilter(f ServerFilter) {
	defaultApp.registerPreServerFilter(f)
}

// RegisterPostServerFilter registers the server filter, executed in order after every request
// Deprecated: As of TarsGo 1.5, it is recommended to use tars.UseServerFilterMiddleware for the same functionality.
func RegisterPostServerFilter(f ServerFilter) {
	defaultApp.registerPostServerFilter(f)
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

// registerClientFilter  registers the Client filter , and will be executed in every request.
func (a *application) registerClientFilter(f ClientFilter) {
	a.allFilters.registerClientFilter(f)
}

// registerPreClientFilter registers the client filter, and will be executed in order before every request
func (a *application) registerPreClientFilter(f ClientFilter) {
	a.allFilters.registerPreClientFilter(f)
}

// registerPostClientFilter registers the client filter, and will be executed in order after every request
func (a *application) registerPostClientFilter(f ClientFilter) {
	a.allFilters.registerPostClientFilter(f)
}

// registerServerFilter register the server filter.
func (a *application) registerServerFilter(f ServerFilter) {
	a.allFilters.registerServerFilter(f)
}

// registerPreServerFilter registers the server filter, executed in order before every request
func (a *application) registerPreServerFilter(f ServerFilter) {
	a.allFilters.registerPreServerFilter(f)
}

// registerPostServerFilter registers the server filter, executed in order after every request
func (a *application) registerPostServerFilter(f ServerFilter) {
	a.allFilters.registerPostServerFilter(f)
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

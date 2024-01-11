package tars

import "time"

// MaxInvoke number of worker routine to handle client request
// zero means no control, just one goroutine for a client request.
// runtime.NumCPU() usually the best performance in the benchmark.
var MaxInvoke int32

const (
	// for now ,some option should update from remote config

	// Version is tars version
	Version string = "1.4.5"

	// server

	// AcceptTimeout accept timeout,default value is 500 milliseconds
	AcceptTimeout = 500
	// ReadTimeout zero millisecond for not set read deadline for Conn (better  performance)
	ReadTimeout = 0
	// WriteTimeout zero millisecond for not set write deadline for Conn (better performance)
	WriteTimeout = 0
	// HandleTimeout zero millisecond for not set deadline for invoke user interface (better performance)
	HandleTimeout = 0
	// IdleTimeout idle timeout,default value is 600000 milliseconds
	IdleTimeout = 600000
	// ZombieTimeout zombie timeout,default value is 10000 milliseconds
	ZombieTimeout = 10000
	// QueueCap queue gap
	QueueCap int = 10000000

	// client

	// Stat default value
	Stat = "tars.tarsstat.StatObj"
	// Property default value
	Property = "tars.tarsproperty.PropertyObj"
	// ModuleName default value
	ModuleName = "tup_client"
	// ClientQueueLen client queue length
	ClientQueueLen int = 10000
	// ClientIdleTimeout client idle timeout,default value is 600000 milliseconds
	ClientIdleTimeout = 600000
	// ClientReadTimeout client read timeout,default value is 100 milliseconds
	ClientReadTimeout = 100
	// ClientWriteTimeout client write timeout,default value is 3000 milliseconds
	ClientWriteTimeout = 3000
	// ReqDefaultTimeout request default timeout
	ReqDefaultTimeout int32 = 3000
	// ClientDialTimeout connection dial timeout
	ClientDialTimeout = 3000
	// ObjQueueMax obj queue max number
	ObjQueueMax int32 = 100000

	// log
	defaultRotateN      = 10
	defaultRotateSizeMB = 100

	// remote log

	// remoteLogQueueSize remote log queue size
	remoteLogQueueSize int = 500000
	// remoteLogMaxNumOneTime is the max logs for reporting in one time.
	remoteLogMaxNumOneTime int = 2000
	// remoteLogInterval log report interval, default value is 1000 milliseconds
	remoteLogInterval = time.Second

	// report

	// PropertyReportInterval property report interval,default value is 10000 milliseconds
	PropertyReportInterval = 10000
	// StatReportInterval stat report interval,default value is 10000 milliseconds
	StatReportInterval = 10000
	// StatReportChannelBufLen stat report channel len
	StatReportChannelBufLen = 100000

	// mainloop

	// MainLoopTicker main loop ticker,default value is 10000 milliseconds
	MainLoopTicker = 10000

	// adapter

	// communicator default ,update from remote config
	refreshEndpointInterval int = 60000
	reportInterval          int = 5000
	// AsyncInvokeTimeout async invoke timeout
	AsyncInvokeTimeout int = 3000

	// check endpoint status every 1000 ms
	checkStatusInterval int = 1000

	// adapter proxy keepAlive with server ,default value is 0 means close keepAlive. milliseconds
	keepAliveInterval int = 0

	// try interval after every 30s
	tryTimeInterval int64 = 30
	// failN & failInterval shows how many times fail in the failInterval second,the server will be blocked.
	fainN        int32 = 5
	failInterval int64 = 5

	// default check every 60 second , and over 2 is failed ,
	// and timeout ratio over 0.5 ,the server will be blocked.
	checkTime int64   = 60
	overN     int32   = 2
	failRatio float32 = 0.5

	// tcp network config

	// TCPReadBuffer tcp read buffer length
	TCPReadBuffer = 128 * 1024 * 1024
	// TCPWriteBuffer tcp write buffer length
	TCPWriteBuffer = 128 * 1024 * 1024
	// TCPNoDelay set tcp no delay
	TCPNoDelay = false

	// GracedownTimeout set timeout (milliseconds) for grace shutdown
	GracedownTimeout = 60000

	// MaxPackageLength maximum length of the request
	MaxPackageLength = 10485760
)

package tars

import "time"

//MaxInvoke number of worker routine to handle client request
//zero means no control, just one goroutine for a client request.
//runtime.NumCPU() usually best performance in the benchmark.
var MaxInvoke int32

const (
	//for now ,some option should update from remote config

	//TarsVersion is tars version
	TarsVersion string = "1.1.2"

	//server

	//AcceptTimeout accept timeout,defaultvalue is 500 milliseconds
	AcceptTimeout = 500
	//ReadTimeout zero millisecond for not set read deadline for Conn (better  performance)
	ReadTimeout = 0
	//WriteTimeout zero millisecond for not set write deadline for Conn (better performance)
	WriteTimeout = 0
	//HandleTimeout zero millisecond for not set deadline for invoke user interface (better performance)
	HandleTimeout = 0
	//IdleTimeout idle timeout,defaultvalue is 600000 milliseconds
	IdleTimeout = 600000
	//ZombileTimeout zombile timeout,defaultvalue is 10000 milliseconds
	ZombileTimeout = 10000
	//QueueCap queue gap
	QueueCap int = 10000000

	//client

	//ClientQueueLen client queue length
	ClientQueueLen int = 10000
	//ClientIdleTimeout client idle timeout,defaultvalue is 600000 milliseconds
	ClientIdleTimeout = 600000
	//ClientReadTimeout client read timeout,defaultvalue is 100 milliseconds
	ClientReadTimeout = 100
	//ClientWriteTimeout client write timeout,defaultvalue is 3000 milliseconds
	ClientWriteTimeout = 3000
	//ReqDefaultTimeout request default timeout
	ReqDefaultTimeout int32 = 3000
	//ClientDialTimeout connection dial timeout
	ClientDialTimeout = 3000
	//ObjQueueMax obj queue max number
	ObjQueueMax int32 = 100000

	//log
	defaultRotateN      = 10
	defaultRotateSizeMB = 100

	//remotelog

	//remoteLogQueueSize remote log queue size
	remoteLogQueueSize int = 500000
	//remoteLogMaxNumOneTime is the max logs for reporting in one time.
	remoteLogMaxNumOneTime int = 2000
	//remoteLogInterval log report interval, defaultvalue is 1000 milliseconds
	remoteLogInterval time.Duration = 1000 * time.Millisecond

	//report

	//PropertyReportInterval property report interval,defaultvalue is 10000 milliseconds
	PropertyReportInterval = 10000
	//StatReportInterval stat report interval,defaultvalue is 10000 milliseconds
	StatReportInterval = 10000
	// StatReportChannelBufLen stat report channel len
	StatReportChannelBufLen = 100000

	//mainloop

	//MainLoopTicker main loop ticker,defaultvalue is 10000 milliseconds
	MainLoopTicker = 10000

	//adapter

	//AdapterProxyTicker adapter proxy ticker,defaultvalue is 10000 milliseconds
	AdapterProxyTicker = 10000
	//AdapterProxyResetCount adapter proxy reset count
	AdapterProxyResetCount int = 5

	//communicator default ,update from remote config
	refreshEndpointInterval int = 60000
	reportInterval          int = 5000
	//AsyncInvokeTimeout async invoke timeout
	AsyncInvokeTimeout int = 3000

	//check endpoint status every 1000 ms
	checkStatusInterval int = 1000

	//try interval after every 30s
	tryTimeInterval int64 = 30
	//failN & failInterval shows how many times fail in the failInterval second,the server will be blocked.
	fainN        int32 = 5
	failInterval int64 = 5

	//default check every 60 second , and over 2 is failed ,
	//and timeout ratio over 0.5 ,the server will be blocked.
	checkTime int64   = 60
	overN     int32   = 2
	failRatio float32 = 0.5

	//tcp network config

	//TCPReadBuffer tcp read buffer length
	TCPReadBuffer = 128 * 1024 * 1024
	//TCPWriteBuffer tcp write buffer length
	TCPWriteBuffer = 128 * 1024 * 1024
	//TCPNoDelay set tcp no delay
	TCPNoDelay = false

	//GracedownTimeout set timeout (milliseconds) for grace shutdown
	GracedownTimeout   = 60000
	graceCheckInterval = time.Millisecond * 500

	//MaxPackageLength maximum length of the request
	MaxPackageLength = 10485760
)

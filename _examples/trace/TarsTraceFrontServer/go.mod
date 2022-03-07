module trace/frontend

go 1.16

require (
	github.com/TarsCloud/TarsGo v1.2.0
	trace/backend v0.0.0-00010101000000-000000000000
)

replace (
	github.com/TarsCloud/TarsGo v1.2.0 => ../../../
	trace/backend => ../TarsTraceBackServer
)

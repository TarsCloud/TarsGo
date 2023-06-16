package model

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)

type Callback interface {
	Dispatch(context.Context, *requestf.RequestPacket, *requestf.ResponsePacket, error) (int32, error)
}

// Servant is interface for call the remote server.
type Servant interface {
	Name() string
	TarsInvoke(ctx context.Context, cType byte,
		sFuncName string,
		buf []byte,
		status map[string]string,
		context map[string]string,
		resp *requestf.ResponsePacket) error

	TarsInvokeAsync(ctx context.Context, cType byte,
		sFuncName string,
		buf []byte,
		status map[string]string,
		context map[string]string,
		resp *requestf.ResponsePacket,
		callback Callback) error

	TarsSetTimeout(t int)
	TarsSetProtocol(Protocol)
	Endpoints() []*endpoint.Endpoint
	SetPushCallback(callback func([]byte))
}

type Protocol interface {
	RequestPack(*requestf.RequestPacket) ([]byte, error)
	ResponseUnpack([]byte) (*requestf.ResponsePacket, error)
	ParsePackage([]byte) (int, int)
}

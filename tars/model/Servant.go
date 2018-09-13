package model

import "github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"

type Servant interface {
	Tars_invoke(ctype byte,
		sFuncName string,
		buf []byte,
		status map[string]string,
		context map[string]string,
		Resp *requestf.ResponsePacket) error
	TarsSetTimeout(t int)
}

package push

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// Client is the pushing client
type Client struct {
	servant  model.Servant
	callback func(data []byte)
}

// SetServant implements client servant
func (c *Client) SetServant(s model.Servant) {
	s.SetPushCallback(c.callback)
	c.servant = s
}

// NewClient returns the client for pushing message
func NewClient(callback func(data []byte)) *Client {
	return &Client{callback: callback}
}

// Connect starts to connect to pushing server
func (c *Client) Connect(req []byte) ([]byte, error) {
	rsp := &requestf.ResponsePacket{}
	if err := c.servant.TarsInvoke(context.Background(), 0, "push", req, nil, nil, rsp); err != nil {
		return nil, err
	}
	return tools.Int8ToByte(rsp.SBuffer), nil
}

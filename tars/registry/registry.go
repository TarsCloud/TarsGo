package registry

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
)

type Endpoint = endpointf.EndpointF

// ServantInstance is an instance of a service in a discovery system.
type ServantInstance struct {
	TarsVersion string   `json:"tars_version"`
	App         string   `json:"app"`
	Server      string   `json:"server"`
	EnableSet   bool     `json:"enable_set"`
	SetDivision string   `json:"set_division"`
	Protocol    string   `json:"protocol"`
	Servant     string   `json:"servant"`
	Endpoint    Endpoint `json:"endpoint"`
}

// Registrar is service registrar.
type Registrar interface {
	Registry(ctx context.Context, servant *ServantInstance) error
	Deregister(ctx context.Context, servant *ServantInstance) error
	// QueryServant service discovery
	QueryServant(ctx context.Context, id string) (activeEp []Endpoint, inactiveEp []Endpoint, err error)
	QueryServantBySet(ctx context.Context, id, set string) (activeEp []Endpoint, inactiveEp []Endpoint, err error)
}

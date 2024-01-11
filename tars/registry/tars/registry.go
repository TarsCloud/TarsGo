package tars

import (
	"context"
	"fmt"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/registry"
)

type Context interface {
	Context() map[string]string
}

type tarsRegistry struct {
	query *queryf.QueryF
	ctx   Context
}

func New(query *queryf.QueryF, ctx Context) registry.Registrar {
	return &tarsRegistry{query: query, ctx: ctx}
}

func (t *tarsRegistry) Registry(_ context.Context, _ *registry.ServantInstance) error {
	return nil
}

func (t *tarsRegistry) Deregister(_ context.Context, _ *registry.ServantInstance) error {
	return nil
}

func (t *tarsRegistry) QueryServant(ctx context.Context, id string) (activeEp []registry.Endpoint, inactiveEp []registry.Endpoint, err error) {
	ret, err := t.query.FindObjectByIdInSameGroupWithContext(ctx, id, &activeEp, &inactiveEp, t.ctx.Context())
	if err != nil {
		return nil, nil, err
	}
	if ret != 0 {
		return nil, nil, fmt.Errorf("QueryServant id: %s fail, ret: %d", id, ret)
	}
	return activeEp, inactiveEp, nil
}

func (t *tarsRegistry) QueryServantBySet(ctx context.Context, id, set string) (activeEp []registry.Endpoint, inactiveEp []registry.Endpoint, err error) {
	ret, err := t.query.FindObjectByIdInSameSetWithContext(ctx, id, set, &activeEp, &inactiveEp, t.ctx.Context())
	if err != nil {
		return nil, nil, err
	}
	if ret != 0 {
		return nil, nil, fmt.Errorf("QueryServantBySet id: %s, setId: %s fail, ret: %d", id, set, ret)
	}
	return activeEp, inactiveEp, nil
}

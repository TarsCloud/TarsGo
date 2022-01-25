package random

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/selector"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

type Random struct {
	sync.RWMutex
	enableWeight            bool
	mapValues               map[string]struct{}
	endpoints               []endpoint.Endpoint
	staticWeightRouterCache []int
	rand                    *rand.Rand
}

var _ selector.Selector = (*Random)(nil)

func New(enableWeight bool) *Random {
	return &Random{
		enableWeight: enableWeight,
		mapValues:    make(map[string]struct{}),
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Random) Select(_ selector.Message) (endpoint.Endpoint, error) {
	r.RLock()
	defer r.RUnlock()
	var ep endpoint.Endpoint
	if len(r.endpoints) == 0 {
		return ep, errors.New("random: no such endpoint.Endpoint")
	}
	if len(r.staticWeightRouterCache) != 0 {
		idx := r.staticWeightRouterCache[r.rand.Intn(len(r.staticWeightRouterCache))]
		return r.endpoints[idx], nil
	}
	return r.endpoints[r.rand.Intn(len(r.endpoints))], nil
}

func (r *Random) Refresh(eps []endpoint.Endpoint) {
	r.Lock()
	defer r.Unlock()
	r.mapValues = make(map[string]struct{}, len(eps))
	r.endpoints = make([]endpoint.Endpoint, 0, len(eps))
	for _, ep := range eps {
		r.addLocked(ep)
	}
	r.reBuildLocked()
}

func (r *Random) Add(ep endpoint.Endpoint) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.mapValues[ep.HashKey()]; ok {
		return fmt.Errorf("random: endpoint %+v already exists", ep)
	}
	r.endpoints = append(r.endpoints, ep)
	r.mapValues[ep.HashKey()] = struct{}{}
	r.reBuildLocked()
	return nil
}

func (r *Random) addLocked(ep endpoint.Endpoint) error {
	if _, ok := r.mapValues[ep.HashKey()]; ok {
		return fmt.Errorf("random: endpoint %+v already exists", ep)
	}
	r.endpoints = append(r.endpoints, ep)
	r.mapValues[ep.HashKey()] = struct{}{}
	return nil
}

func (r *Random) Remove(ep endpoint.Endpoint) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.mapValues[ep.HashKey()]; !ok {
		return fmt.Errorf("random: endpoint %+v already removed", ep)
	}
	delete(r.mapValues, ep.HashKey())
	for i, n := range r.endpoints {
		if n.HashKey() == ep.HashKey() {
			r.endpoints = append(r.endpoints[:i], r.endpoints[i+1:]...)
			break
		}
	}
	r.reBuildLocked()
	return nil
}

func (r *Random) reBuildLocked() {
	r.staticWeightRouterCache = nil
	if r.enableWeight {
		r.staticWeightRouterCache = selector.BuildStaticWeightList(r.endpoints)
	}
}

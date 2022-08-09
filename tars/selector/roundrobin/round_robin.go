package roundrobin

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/selector"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

type RoundRobin struct {
	sync.RWMutex
	enableWeight             bool
	mapValues                map[string]struct{}
	endpoints                []endpoint.Endpoint
	lastPosition             uint64
	staticWeightRouterCache  []int
	lastStaticWeightPosition uint64
}

var _ selector.Selector = (*RoundRobin)(nil)

func New(enableWeight bool) *RoundRobin {
	return &RoundRobin{
		enableWeight: enableWeight,
		mapValues:    make(map[string]struct{}),
	}
}

func (r *RoundRobin) Select(_ selector.Message) (endpoint.Endpoint, error) {
	r.RLock()
	defer r.RUnlock()
	var ep endpoint.Endpoint
	if len(r.endpoints) == 0 {
		return ep, errors.New("round_robin: no such endpoint.Endpoint")
	}
	if len(r.staticWeightRouterCache) != 0 {
		idx := atomic.AddUint64(&r.lastStaticWeightPosition, 1)
		return r.endpoints[r.staticWeightRouterCache[idx%uint64(len(r.staticWeightRouterCache))]], nil
	}
	idx := atomic.AddUint64(&r.lastPosition, 1)
	ep = r.endpoints[idx%uint64(len(r.endpoints))]
	return ep, nil
}

func (r *RoundRobin) Refresh(eps []endpoint.Endpoint) {
	r.Lock()
	defer r.Unlock()
	r.mapValues = make(map[string]struct{}, len(eps))
	r.endpoints = make([]endpoint.Endpoint, 0, len(eps))
	for _, ep := range eps {
		r.addLocked(ep)
	}
	r.reBuildLocked()
}

func (r *RoundRobin) Add(ep endpoint.Endpoint) error {
	r.Lock()
	defer r.Unlock()
	if err := r.addLocked(ep); err != nil {
		return err
	}
	r.reBuildLocked()
	return nil
}

func (r *RoundRobin) addLocked(ep endpoint.Endpoint) error {
	if _, ok := r.mapValues[ep.HashKey()]; ok {
		return fmt.Errorf("round_robin: endpoint %+v already exists", ep)
	}
	r.endpoints = append(r.endpoints, ep)
	r.mapValues[ep.HashKey()] = struct{}{}
	return nil
}

func (r *RoundRobin) Remove(ep endpoint.Endpoint) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.mapValues[ep.HashKey()]; !ok {
		return fmt.Errorf("round_robin: endpoint %+v already removed", ep)
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

func (r *RoundRobin) reBuildLocked() {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.lastPosition, r.lastStaticWeightPosition = 0, 0
	if n := len(r.endpoints); n > 0 {
		r.lastPosition = uint64(rd.Intn(n))
	}
	r.staticWeightRouterCache = nil
	if r.enableWeight {
		r.staticWeightRouterCache = selector.BuildStaticWeightList(r.endpoints)
		if n := len(r.staticWeightRouterCache); n > 0 {
			r.lastStaticWeightPosition = uint64(rd.Intn(n))
		}
	}
}

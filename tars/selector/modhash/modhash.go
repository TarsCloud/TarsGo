package modhash

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TarsCloud/TarsGo/tars/selector"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

type ModHash struct {
	sync.RWMutex
	enableWeight            bool
	mapValues               map[string]struct{}
	endpoints               []endpoint.Endpoint
	staticWeightRouterCache []int
}

var _ selector.Selector = (*ModHash)(nil)

func New(enableWeight bool) *ModHash {
	return &ModHash{
		enableWeight: enableWeight,
		mapValues:    make(map[string]struct{}),
	}
}

func (m *ModHash) Select(msg selector.Message) (endpoint.Endpoint, error) {
	m.RLock()
	defer m.RUnlock()
	var ep endpoint.Endpoint
	if len(m.endpoints) == 0 {
		return ep, errors.New("modhash: no such endpoint.Endpoint")
	}
	hashCode := msg.HashCode()
	if len(m.staticWeightRouterCache) != 0 {
		idx := m.staticWeightRouterCache[hashCode%uint32(len(m.staticWeightRouterCache))]
		return m.endpoints[idx], nil
	}
	return m.endpoints[hashCode%uint32(len(m.endpoints))], nil
}

func (m *ModHash) Refresh(eps []endpoint.Endpoint) {
	m.Lock()
	defer m.Unlock()
	m.mapValues = make(map[string]struct{}, len(eps))
	m.endpoints = make([]endpoint.Endpoint, 0, len(eps))
	for _, ep := range eps {
		m.addLocked(ep)
	}
	m.reBuildLocked()
}

func (m *ModHash) Add(ep endpoint.Endpoint) error {
	m.Lock()
	defer m.Unlock()
	if err := m.addLocked(ep); err != nil {
		return err
	}
	m.reBuildLocked()
	return nil
}

func (m *ModHash) addLocked(ep endpoint.Endpoint) error {
	if _, ok := m.mapValues[ep.HashKey()]; ok {
		return fmt.Errorf("modhash: endpoint %+v already exists", ep)
	}
	m.endpoints = append(m.endpoints, ep)
	m.mapValues[ep.HashKey()] = struct{}{}
	return nil
}

func (m *ModHash) Remove(ep endpoint.Endpoint) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.mapValues[ep.HashKey()]; !ok {
		return fmt.Errorf("modhash: endpoint %+v already removed", ep)
	}
	delete(m.mapValues, ep.HashKey())
	for i, n := range m.endpoints {
		if n.HashKey() == ep.HashKey() {
			m.endpoints = append(m.endpoints[:i], m.endpoints[i+1:]...)
			break
		}
	}
	m.reBuildLocked()
	return nil
}

func (m *ModHash) reBuildLocked() {
	m.staticWeightRouterCache = nil
	if m.enableWeight {
		m.staticWeightRouterCache = selector.BuildStaticWeightList(m.endpoints)
	}
}

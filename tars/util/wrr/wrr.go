package wrr

import (
	"errors"
	"sync"
)

// KV is the key value type.
type KV interface {
	String() string
}

type NodeWeight struct {
	ep              string
	weight          int32
	node            KV
	currectWeight   int32
	effectiveWeight int32
}

type WrrBalance struct {
	lock *sync.RWMutex

	lastPos           int
	staticRouterCache []int
	activeNodeWeight  []*NodeWeight
}

func NewWrrBalance() *WrrBalance {
	return &WrrBalance{
		lock:    &sync.RWMutex{},
		lastPos: 0,
	}
}

// Add add the node to the rss
func (w *WrrBalance) Add(node KV, weight int32) error {

	w.lock.Lock()
	defer w.lock.Unlock()

	if weight <= 0 {
		return errors.New("weight less than or equal to 0")
	}

	var index int
	for index = 0; index < len(w.activeNodeWeight); index++ {
		if w.activeNodeWeight[index].ep == node.String() {
			break
		}
	}

	if index != len(w.activeNodeWeight) && w.activeNodeWeight[index].weight != weight {
		w.activeNodeWeight[index].weight = weight
		w.activeNodeWeight[index].effectiveWeight = weight
	} else {
		nw := &NodeWeight{ep: node.String(), node: node, weight: weight}
		nw.effectiveWeight = weight
		w.activeNodeWeight = append(w.activeNodeWeight, nw)
	}

	err := w.rebuildRouterCache()

	return err
}

func (w *WrrBalance) Remove(node string) error {

	w.lock.Lock()
	defer w.lock.Unlock()

	var index int
	for index = 0; index < len(w.activeNodeWeight); index++ {
		if w.activeNodeWeight[index].ep == node {
			break
		}
	}

	if index != len(w.activeNodeWeight) {
		w.activeNodeWeight = append(w.activeNodeWeight[:index], w.activeNodeWeight[index+1:]...)
	} else {
		return errors.New("node already removed")
	}

	err := w.rebuildRouterCache()

	return err
}

func (w *WrrBalance) Next() (KV, bool) {
	if len(w.staticRouterCache) == 0 {
		return nil, false
	}

	w.lock.RLock()
	defer w.lock.RUnlock()

	newPos := (w.lastPos + 1) % len(w.staticRouterCache)
	nw := w.activeNodeWeight[w.staticRouterCache[newPos]]
	w.lastPos = newPos

	return nw.node, true
}

func (w *WrrBalance) rebuildRouterCache() error {
	weights := make([]int32, 0)

	for _, nw := range w.activeNodeWeight {
		weights = append(weights, nw.weight)
	}

	if len(weights) <= 0 {
		return errors.New("no valid node")
	}

	var maxR, maxRouterR, maxWeight, minWeight int32
	w.staticRouterCache = make([]int, 0)

	maxWeight = weights[0]
	minWeight = weights[0]

	for _, weight := range weights {
		if weight > maxWeight {
			maxWeight = weight
		}
		if weight < minWeight {
			minWeight = weight
		}
	}

	if minWeight > 0 {
		maxR = maxWeight / minWeight
		if maxR < 10 {
			maxR = 10
		}
		if maxR > 100 {
			maxR = 100
		}
	} else {
		maxR = 1
		maxWeight = 1
	}

	for index, nw := range w.activeNodeWeight {
		newWeight := nw.weight * maxR / maxWeight
		if newWeight > 0 {
			nw.effectiveWeight = newWeight
			maxRouterR += newWeight
		} else {
			nw.effectiveWeight = 0
			w.staticRouterCache = append(w.staticRouterCache, index)
		}
	}

	var i int32
	for i = 0; i < maxRouterR; i++ {
		best := -1
		var total, tmpMax int32

		for index, nw := range w.activeNodeWeight {
			nw.currectWeight += nw.effectiveWeight
			total += nw.effectiveWeight

			if best == -1 || nw.currectWeight > tmpMax {
				best = index
				tmpMax = nw.currectWeight
			}
		}

		w.activeNodeWeight[best].currectWeight -= total
		w.staticRouterCache = append(w.staticRouterCache, best)
	}

	w.lastPos = 0

	return nil
}

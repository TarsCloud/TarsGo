package consistenthash

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

// ChMap consistent hash map
type ChMap struct {
	lock       *sync.RWMutex
	replicates int
	sortedKeys []uint32
	hashRing   map[uint32]KV
	mapValues  map[string]bool
}

// KV is the key value type.
type KV interface {
	String() string
}

// NewChMap  create a ChMap which has replicates of virtual nodes.
func NewChMap(replicates int) *ChMap {
	return &ChMap{
		lock:       &sync.RWMutex{},
		replicates: replicates,
		hashRing:   make(map[uint32]KV),
		mapValues:  make(map[string]bool),
	}
}

// Find finds a nodes to put the string key
func (c *ChMap) Find(key string) (KV, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if len(c.sortedKeys) == 0 {
		return nil, false
	}
	hashKey := crc32.ChecksumIEEE([]byte(key))
	index := sort.Search(len(c.sortedKeys), func(x int) bool {
		return c.sortedKeys[x] >= hashKey
	})
	if index >= len(c.sortedKeys) {
		index = 0
	}

	return c.hashRing[c.sortedKeys[index]], true
}

// FindUint32  finds a nodes to put the uint32 key
func (c *ChMap) FindUint32(key uint32) (KV, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if len(c.sortedKeys) == 0 {
		return nil, false
	}
	index := sort.Search(len(c.sortedKeys), func(x int) bool {
		return c.sortedKeys[x] >= key
	})
	if index >= len(c.sortedKeys) {
		index = 0
	}

	return c.hashRing[c.sortedKeys[index]], true
}

// Add add the node to the hash ring
func (c *ChMap) Add(node KV) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.mapValues[node.String()]; ok {
		return errors.New("node already exists")
	}
	for i := 0; i < c.replicates; i++ {
		virtualHost := fmt.Sprintf("%d#%s", i, node.String())
		virtualKey := crc32.ChecksumIEEE([]byte(virtualHost))
		c.hashRing[virtualKey] = node
		c.sortedKeys = append(c.sortedKeys, virtualKey)
	}
	sort.Slice(c.sortedKeys, func(x int, y int) bool {
		return c.sortedKeys[x] < c.sortedKeys[y]
	})
	c.mapValues[node.String()] = true
	return nil
}

// Remove remove the node and all the vatual nodes from the key
func (c *ChMap) Remove(node string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.mapValues[node]; !ok {
		return errors.New("host already removed")
	}
	delete(c.mapValues, node)
	for i := 0; i < c.replicates; i++ {
		virtualHost := fmt.Sprintf("%d#%s", i, node)
		virtualKey := crc32.ChecksumIEEE([]byte(virtualHost))
		delete(c.hashRing, virtualKey)
	}
	c.reBuildHashRing()
	return nil
}

func (c *ChMap) reBuildHashRing() {
	c.sortedKeys = nil
	for vk := range c.hashRing {
		c.sortedKeys = append(c.sortedKeys, vk)
	}
	sort.Slice(c.sortedKeys, func(x, y int) bool {
		return c.sortedKeys[x] < c.sortedKeys[y]
	})
}

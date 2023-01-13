package consistenthash

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"sync"
	"unsafe"

	"github.com/TarsCloud/TarsGo/tars/selector"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

// ConsistentHash consistent hash map
type ConsistentHash struct {
	sync.RWMutex
	hash         hash
	enableWeight bool
	replicates   int
	mapValues    map[string]struct{}
	hashRing     map[uint32]endpoint.Endpoint
	sortedKeys   []uint32
}

var _ selector.Selector = (*ConsistentHash)(nil)

type HashAlgorithmType int

const (
	KetamaHash  HashAlgorithmType = 0
	DefaultHash HashAlgorithmType = 1
)

type hash interface {
	Hash(key string) uint32
	GetHashType() HashAlgorithmType
}

type KetamaHashAlg struct {
}

func (k KetamaHashAlg) Hash(key string) uint32 {
	p := md5.Sum([]byte(key))
	h := uint32(p[3]&0xFF)<<24 | uint32(p[2]&0xFF)<<16 | uint32(p[1]&0xFF)<<8 | uint32(p[0]&0xFF)
	return h
}

func (k KetamaHashAlg) GetHashType() HashAlgorithmType {
	return KetamaHash
}

type DefaultHashAlg struct {
}

func (k DefaultHashAlg) Hash(key string) uint32 {
	p := md5.Sum([]byte(key))
	h := *((*uint32)(unsafe.Pointer(&p))) ^ *((*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + 4))) ^ *((*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + 8))) ^ *((*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + 12)))
	return h
}

func (k DefaultHashAlg) GetHashType() HashAlgorithmType {
	return DefaultHash
}

// New  create a ConsistentHash which has replicates of virtual endpoint.Endpoint.
func New(enableWeight bool, hashType HashAlgorithmType) *ConsistentHash {
	var h hash
	if hashType == KetamaHash {
		h = KetamaHashAlg{}
	} else {
		h = DefaultHashAlg{}
	}
	return &ConsistentHash{
		hash:         h,
		enableWeight: enableWeight,
		replicates:   selector.ConHashVirtualNodes,
		mapValues:    make(map[string]struct{}),
		hashRing:     make(map[uint32]endpoint.Endpoint),
	}
}

func (c *ConsistentHash) Select(msg selector.Message) (ep endpoint.Endpoint, err error) {
	var ok bool
	ep, ok = c.FindInt32(msg.HashCode())
	if !ok {
		return ep, errors.New("consistenthash: no such endpoint.Endpoint")
	}
	return ep, nil
}

// Find finds a endpoint.Endpoint to put the string key
func (c *ConsistentHash) Find(key string) (endpoint.Endpoint, bool) {
	c.RLock()
	defer c.RUnlock()
	var point endpoint.Endpoint
	if len(c.sortedKeys) == 0 {
		return point, false
	}
	hashKey := c.hash.Hash(key)
	index := sort.Search(len(c.sortedKeys), func(x int) bool {
		return c.sortedKeys[x] >= hashKey
	})
	if index >= len(c.sortedKeys) {
		index = 0
	}

	return c.hashRing[c.sortedKeys[index]], true
}

// FindInt32  finds a endpoint.Endpoint to put the uint32 key
func (c *ConsistentHash) FindInt32(key uint32) (endpoint.Endpoint, bool) {
	c.RLock()
	defer c.RUnlock()
	var point endpoint.Endpoint
	if len(c.sortedKeys) == 0 {
		return point, false
	}
	index := sort.Search(len(c.sortedKeys), func(x int) bool {
		return c.sortedKeys[x] >= key
	})

	if index >= len(c.sortedKeys) {
		index = 0
	}
	return c.hashRing[c.sortedKeys[index]], true
}

func (c *ConsistentHash) Refresh(eps []endpoint.Endpoint) {
	c.Lock()
	defer c.Unlock()
	c.mapValues = make(map[string]struct{}, len(eps))
	c.hashRing = make(map[uint32]endpoint.Endpoint, len(eps))
	c.sortedKeys = nil
	for _, ep := range eps {
		_ = c.addLocked(ep)
	}
	c.sort()
}

// Add the ep to the hash ring
func (c *ConsistentHash) Add(ep endpoint.Endpoint) error {
	c.Lock()
	defer c.Unlock()
	if err := c.addLocked(ep); err != nil {
		return err
	}
	c.sort()
	return nil
}

func (c *ConsistentHash) addLocked(ep endpoint.Endpoint) error {
	if _, ok := c.mapValues[ep.HashKey()]; ok {
		return fmt.Errorf("consistenthash: endpoint %+v already exists", ep)
	}
	weight := c.weight(ep.Weight)
	for i := 0; i < weight; i++ {
		virtualHost := fmt.Sprintf("%s_%d", ep.HashKey(), i)
		if c.hash.GetHashType() == KetamaHash {
			p := md5.Sum([]byte(virtualHost))
			for k := 0; k < 4; k++ {
				virtualKey := uint32(p[4*k+3]&0xFF)<<24 | uint32(p[4*k+2]&0xFF)<<16 | uint32(p[4*k+1]&0xFF)<<8 | uint32(p[4*k+0]&0xFF)
				c.hashRing[virtualKey] = ep
				c.sortedKeys = append(c.sortedKeys, virtualKey)
			}
		} else {
			virtualKey := c.hash.Hash(virtualHost)
			c.hashRing[virtualKey] = ep
			c.sortedKeys = append(c.sortedKeys, virtualKey)
		}
	}
	c.mapValues[ep.HashKey()] = struct{}{}
	return nil
}

// Remove the ep and all the virtual eps from the key
func (c *ConsistentHash) Remove(ep endpoint.Endpoint) error {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.mapValues[ep.HashKey()]; !ok {
		return fmt.Errorf("consistenthash: endpoint %+v already removed", ep)
	}
	delete(c.mapValues, ep.HashKey())
	weight := c.weight(ep.Weight)
	for i := 0; i < weight; i++ {
		virtualHost := fmt.Sprintf("%s_%d", ep.HashKey(), i)
		if c.hash.GetHashType() == KetamaHash {
			p := md5.Sum([]byte(virtualHost))
			for k := 0; k < 4; k++ {
				virtualKey := uint32(p[4*k+3]&0xFF)<<24 | uint32(p[4*k+2]&0xFF)<<16 | uint32(p[4*k+1]&0xFF)<<8 | uint32(p[4*k+0]&0xFF)
				delete(c.hashRing, virtualKey)
			}
		} else {
			virtualKey := c.hash.Hash(virtualHost)
			delete(c.hashRing, virtualKey)
		}
	}
	c.reBuildHashRingLocked()
	return nil
}

func (c *ConsistentHash) printNode() {
	mapNode := map[string]uint32{}
	// 打印哈希环
	for i, hashCode := range c.sortedKeys {
		var value uint32
		if i == 0 {
			value = 0xFFFFFFFF - c.sortedKeys[len(c.sortedKeys)-1] + c.sortedKeys[i]
			mapNode[c.hashRing[hashCode].Host] += value
		} else {
			value = c.sortedKeys[i] - c.sortedKeys[i-1]
			mapNode[c.hashRing[hashCode].Host] += value
		}
		fmt.Printf("printNode: %d|%s|%v\n", hashCode, c.hashRing[hashCode].Host, mapNode[c.hashRing[hashCode].Host])
	}
	avg, sum, n := float64(100), float64(0), float64(len(mapNode))
	// 打印各个区间比例
	for host, count := range mapNode {
		c := 100*float64(count)*n/0xFFFFFFFF - avg
		fmt.Printf("result: %s|%d|%f\n", host, count, c)
		sum += (float64(count)*100*n/0xFFFFFFFF - avg) * (float64(count)*100*n/0xFFFFFFFF - avg)
	}
	fmt.Printf("variance: %f, size: %d\n", sum/n, len(c.sortedKeys))
}

func (c *ConsistentHash) weight(w int32) int {
	weight := c.replicates
	if c.enableWeight {
		weight = int(w)
	}
	if weight > 0 {
		weight /= 4
		if weight == 0 {
			weight = 1
		}
	}
	return weight
}

func (c *ConsistentHash) reBuildHashRingLocked() {
	c.sortedKeys = make([]uint32, 0, len(c.hashRing))
	for vk := range c.hashRing {
		c.sortedKeys = append(c.sortedKeys, vk)
	}
	c.sort()
}

func (c *ConsistentHash) sort() {
	sort.Slice(c.sortedKeys, func(x, y int) bool {
		return c.sortedKeys[x] < c.sortedKeys[y]
	})
}

package selector

import (
	"math"
	"sort"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

// HashType is the hash type
type HashType int

func (h HashType) String() string {
	if h == ConsistentHash {
		return "ConsistentHash"
	} else {
		return "ModHash"
	}
}

// HashType enum
const (
	ModHash HashType = iota
	ConsistentHash
)

const (
	minStaticWeightLimit = 10
	maxStaticWeightLimit = 100
	ConHashVirtualNodes  = 100
)

type pair struct {
	first  int
	second int
}

type Message interface {
	HashCode() uint32
	HashType() HashType
	IsHash() bool
}

type Selector interface {
	Select(msg Message) (endpoint.Endpoint, error)
	Refresh(node []endpoint.Endpoint)
	Add(node endpoint.Endpoint) error
	Remove(node endpoint.Endpoint) error
}

func BuildStaticWeightList(endpoints []endpoint.Endpoint) []int {
	var maxRange, totalWeight, totalCapacity int
	minWeight, maxWeight := math.MaxInt32, math.MinInt32
	for _, node := range endpoints {
		if endpoint.WeightType(node.WeightType) != endpoint.EStaticWeight {
			return nil
		}
		weight := int(node.Weight)
		totalCapacity += weight
		if maxWeight < weight {
			maxWeight = weight
		}
		if minWeight > weight {
			minWeight = weight
		}
	}

	if minWeight > 0 {
		maxRange = maxWeight / minWeight
		if maxRange < minStaticWeightLimit {
			maxRange = minStaticWeightLimit
		}
		if maxRange > maxStaticWeightLimit {
			maxRange = maxStaticWeightLimit
		}
	} else {
		maxRange, totalWeight = 1, 1
	}

	var weightToId []pair
	idToWeight := map[int]int{}
	staticWeightRouterCache := make([]int, 0, totalCapacity+100)
	for idx, node := range endpoints {
		weight := int(node.Weight) * maxRange / maxWeight
		if weight > 0 {
			totalWeight += weight
			idToWeight[idx] = weight
			weightToId = append(weightToId, pair{weight, idx})
		} else {
			staticWeightRouterCache = append(staticWeightRouterCache, idx)
		}
	}

	for i := 0; i < totalWeight; i++ {
		// 升序
		sort.Slice(weightToId, func(i, j int) bool {
			if weightToId[i].first == weightToId[j].first {
				//return weightToId[i].second < weightToId[j].second
				return endpoints[weightToId[i].second].String() < endpoints[weightToId[j].second].String()
			}
			return weightToId[i].first < weightToId[j].first
		})
		var mulTemp []pair
		first := true
		// 倒序遍历
		for begin := len(weightToId) - 1; begin >= 0; begin-- {
			mIter := weightToId[begin]
			if first {
				first = false
				staticWeightRouterCache = append(staticWeightRouterCache, mIter.second)
				mulTemp = append(mulTemp, pair{mIter.first - totalWeight + idToWeight[mIter.second], mIter.second})
			} else {
				mulTemp = append(mulTemp, pair{mIter.first + idToWeight[mIter.second], mIter.second})
			}
		}
		weightToId = mulTemp
	}
	return staticWeightRouterCache
}

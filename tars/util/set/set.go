// Package set implement
package set

import (
	"sync"
	"sync/atomic"
)

// Set struct
type Set struct {
	el     sync.Map
	length int32
}

// NewSet news a set
func NewSet(opt ...any) *Set {
	s := new(Set)
	for _, o := range opt {
		s.Add(o)
	}
	return s
}

// Add : add an element to Set
func (s *Set) Add(e any) bool {
	_, ok := s.el.LoadOrStore(e, s.length)
	if !ok {
		atomic.AddInt32(&s.length, 1)
	}
	return ok
}

// Remove remove an element from set.
func (s *Set) Remove(e any) {
	atomic.AddInt32(&s.length, -1)
	atomic.CompareAndSwapInt32(&s.length, -1, 0)
	s.el.Delete(e)
}

// Clear : clear the set.
func (s *Set) Clear() {
	atomic.SwapInt32(&s.length, 0)
	s.el = sync.Map{}
}

// Slice init set from []interface.
func (s *Set) Slice() (list []any) {
	s.el.Range(func(k, v any) bool {
		list = append(list, k)
		return true
	})
	return list
}

// Len return the length of the set
func (s *Set) Len() int32 {
	return s.length
}

// Has indicates whether the set has the element.
func (s *Set) Has(e any) bool {
	if _, ok := s.el.Load(e); ok {
		return true
	}
	return false
}

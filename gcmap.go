package gcmap

import (
	"sync"
	"time"
)

const (
	defaultGCInterval = time.Hour
	defaultEntryTTL   = time.Hour * 24
)

// Storage is thread safe LRU storage backed by sync.Map
type Storage struct {
	sync.RWMutex
	data       map[interface{}]*storageItemContainer
	gcInterval time.Duration
	entryTTL   time.Duration
}

type storageItemContainer struct {
	lastUpdate time.Time
	data       interface{}
}

// NewStorage returns new instance of storage
func NewStorage(options ...OptionFn) *Storage {
	st := &Storage{
		data:       make(map[interface{}]*storageItemContainer, 0),
		gcInterval: defaultGCInterval,
		entryTTL:   defaultEntryTTL,
	}
	for _, opt := range options {
		opt(st)
	}
	if st.gcInterval > 0 && st.entryTTL > 0 {
		go st.gcLoop()
	}
	return st
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (s *Storage) Range(fn func(k, v interface{}) bool) {
	s.RLock()
	for k, v := range s.data {
		if exit := fn(k, v.data); !exit {
			break
		}
	}
	s.RUnlock()
}

// Store sets the value for a key.
func (s *Storage) Store(k, v interface{}) {
	s.Lock()
	s.data[k] = &storageItemContainer{
		lastUpdate: time.Now(),
		data:       v,
	}
	s.Unlock()
}

// StoreOrUpdate sets the value for a key if id does not exist
// or calls a callback function with update logic for a value of a key.
// Callback must RETURN modified value.
func (s *Storage) StoreOrUpdate(k, v interface{}, fn func(old, new interface{}) interface{}) {
	s.Lock()
	defer s.Unlock()
	container, exists := s.data[k]
	if !exists {
		s.data[k] = &storageItemContainer{
			lastUpdate: time.Now(),
			data:       v,
		}
		return
	}
	if fn != nil {
		container.data = fn(container.data, v)
		container.lastUpdate = time.Now()
	}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Storage) Load(k interface{}) (interface{}, bool) {
	s.RLock()
	container, found := s.data[k]
	if !found {
		return nil, false
	}
	s.RUnlock()
	return container.data, true
}

// Delete deletes the value for a key.
func (s *Storage) Delete(k interface{}) {
	s.Lock()
	delete(s.data, k)
	s.Unlock()
}

// gcLoop runs garbage collection loop
func (s *Storage) gcLoop() {
	ticker := time.NewTicker(s.gcInterval)
	for range ticker.C {
		s.Lock()
		deleteOlderThan := time.Now().Add(s.entryTTL * -1)
		for k, v := range s.data {
			if v.lastUpdate.Before(deleteOlderThan) {
				delete(s.data, k)
			}
		}
		s.Unlock()
	}
}

// OptionFn is a modification option for storage constructor
type OptionFn func(s *Storage)

// WithGCInterval sets custom GC interval for new storage
func WithGCInterval(gcInterval time.Duration) OptionFn {
	return func(s *Storage) {
		s.gcInterval = gcInterval
	}
}

// WithEntryTTL sets custom TTL for storage entries
func WithEntryTTL(entryTTL time.Duration) OptionFn {
	return func(s *Storage) {
		s.entryTTL = entryTTL
	}
}

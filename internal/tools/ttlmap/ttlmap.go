package ttlmap

import (
	"sync"
	"time"
)

// item is a struct that holds the value and the last access time
type item struct {
	value      any
	lastAccess int64
}

// You can have a single map for an application or few maps for different purposes
type TTLMap struct {
	m    map[string]*item
	mu   sync.Mutex
	stop chan struct{}
}

func New(size int, maxTTL int) *TTLMap {
	m := &TTLMap{
		m:    make(map[string]*item, size),
		stop: make(chan struct{}),
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				m.mu.Lock()
				for k, v := range m.m {
					if now.Unix()-v.lastAccess > int64(maxTTL) {
						delete(m.m, k)
					}
				}
				m.mu.Unlock()
			case <-m.stop:
				return
			}
		}
	}()

	return m
}

// Close stops the background cleanup goroutine.
func (m *TTLMap) Close() {
	close(m.stop)
}

// Put adds a new item to the map or updates the existing one
func (m *TTLMap) Put(k string, v any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	it, ok := m.m[k]
	if !ok {
		it = &item{
			value: v,
		}
	}
	it.value = v
	it.lastAccess = time.Now().Unix()
	m.m[k] = it
}

// Get returns the value of the given key if it exists
func (m *TTLMap) Get(k string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if it, ok := m.m[k]; ok {
		it.lastAccess = time.Now().Unix()
		return it.value, true
	}

	return nil, false
}

// Delete removes the item from the map
func (m *TTLMap) Delete(k string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.m, k)
}

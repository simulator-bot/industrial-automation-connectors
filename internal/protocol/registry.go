package protocol

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry { return &Registry{adapters: make(map[string]Adapter)} }

func (r *Registry) Register(adapter Adapter) error {
	name := adapter.Descriptor().Protocol
	if name == "" {
		return fmt.Errorf("protocol adapter name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("protocol adapter %q already registered", name)
	}
	r.adapters[name] = adapter
	return nil
}

func (r *Registry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

func (r *Registry) Descriptors() []Descriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Descriptor, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		items = append(items, adapter.Descriptor())
	}
	return items
}

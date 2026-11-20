package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
)

var ErrNotFound = errors.New("equipment not found")

// MemoryDevices is a local-development implementation; production storage will
// be introduced behind Devices when the persistence design is finalized.
type MemoryDevices struct {
	mu      sync.RWMutex
	devices map[string]model.Device
}

func NewMemoryDevices() *MemoryDevices {
	return &MemoryDevices{devices: make(map[string]model.Device)}
}

func (r *MemoryDevices) List(_ context.Context) ([]model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.Device, 0, len(r.devices))
	for _, item := range r.devices {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (r *MemoryDevices) Get(_ context.Context, id string) (model.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.devices[id]
	if !ok {
		return model.Device{}, ErrNotFound
	}
	return item, nil
}

func (r *MemoryDevices) Save(_ context.Context, item model.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[item.ID] = item
	return nil
}

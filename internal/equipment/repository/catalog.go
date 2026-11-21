package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
)

var ErrCatalogConflict = errors.New("设备配置已更新，请重新加载后再保存")

type Catalog struct {
	mu   sync.RWMutex
	path string
	data model.EquipmentCatalog
}

func OpenCatalog(path string) (*Catalog, error) {
	store := &Catalog{path: path}
	if path != "" {
		raw, err := os.ReadFile(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return nil, err
		default:
			if err := json.Unmarshal(raw, &store.data); err != nil {
				return nil, fmt.Errorf("decode equipment catalog: %w", err)
			}
		}
	}
	store.data = withCollections(store.data)
	return store, nil
}

// Snapshot returns a read-only revision. Replace swaps the complete revision.
func (s *Catalog) Snapshot() model.EquipmentCatalog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *Catalog) Replace(data model.EquipmentCatalog) (model.EquipmentCatalog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data.Revision != s.data.Revision {
		return model.EquipmentCatalog{}, ErrCatalogConflict
	}
	data = withCollections(data)
	data.Revision++
	if s.path != "" {
		if err := writeCatalog(s.path, data); err != nil {
			return model.EquipmentCatalog{}, err
		}
	}
	s.data = data
	return data, nil
}

func withCollections(data model.EquipmentCatalog) model.EquipmentCatalog {
	if data.Models == nil {
		data.Models = []model.DeviceModel{}
	}
	if data.Points == nil {
		data.Points = []model.ModelPoint{}
	}
	if data.Plans == nil {
		data.Plans = []model.CollectionPlan{}
	}
	if data.Instances == nil {
		data.Instances = []model.DeviceInstance{}
	}
	return data
}

func writeCatalog(path string, data model.EquipmentCatalog) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".equipment-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

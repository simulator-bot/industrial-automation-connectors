package repository

import (
	"context"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
)

// Devices is the persistence boundary. Process owns its implementation and data.
type Devices interface {
	List(ctx context.Context) ([]model.Device, error)
	Get(ctx context.Context, id string) (model.Device, error)
	Save(ctx context.Context, device model.Device) error
}

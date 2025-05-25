package repository

import (
	"context"

	"github.com/quaintdev/webshield/src/internal/entity"
)

// domain management interface
type DomainDataRepository interface {
	GetDomainCategory(domain string) string
	AddDomain(domain string, category string)
}

type SettingsRepository interface {
	GetConfig(ctx context.Context, id string) (*entity.Settings, error)
	UpdateConfig(ctx context.Context, config *entity.Settings) error
	DeleteConfig(ctx context.Context, config string) error
	GetAllConfigs(ctx context.Context) ([]*entity.Settings, error)
}

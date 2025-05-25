package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"

	"github.com/quaintdev/webshield/src/internal/apperrors"
	"github.com/quaintdev/webshield/src/internal/dto"
	"github.com/quaintdev/webshield/src/internal/entity"
	"github.com/quaintdev/webshield/src/internal/repository"
)

type DataMgmtService struct {
	settingsRepo  repository.SettingsRepository
	configService *ApplicationConfigService
}

func NewDataMgmtService(settingsRepo repository.SettingsRepository,
	configService *ApplicationConfigService) *DataMgmtService {
	return &DataMgmtService{

		settingsRepo:  settingsRepo,
		configService: configService,
	}
}

func (s *DataMgmtService) GetConfig(ctx context.Context, configId string) (*dto.PresetResponse, error) {
	config, err := s.settingsRepo.GetConfig(ctx, configId)
	if err != nil {
		slog.Error("failed to get config", "error", err)
		return nil, apperrors.ErrNotFound
	}
	return dto.MakePresetResponse(config), nil
}

func (s *DataMgmtService) AddConfig(ctx context.Context, req *dto.AddPresetRequest) (*dto.PresetResponse, error) {
	slog.Debug("adding new config", "req", req)

	//todo come up with better implementation
	var configId string
	for i := 0; i < 3; i++ {

		configId = generateConfigId()
		slog.Debug("generated configId", "configId", configId)
		_, err := s.settingsRepo.GetConfig(ctx, configId)
		if err != nil {
			if err == apperrors.ErrNotFound {
				break
			}
		}
	}

	updatePresetRequest := &dto.UpdatePresetRequest{
		PresetID: configId,
	}
	updatePresetRequest.PresetName = req.PresetName
	updatePresetRequest.Categories = make([]dto.Category, 0)
	for _, category := range s.configService.GetCategories() {
		updatePresetRequest.Categories = append(updatePresetRequest.Categories, dto.Category{
			Name:   category.Name,
			Status: "inactive",
		})
	}
	updatePresetRequest.Schedule = make([]dto.Schedule, 0)

	config := dto.MakeConfig(updatePresetRequest)
	err := s.settingsRepo.UpdateConfig(ctx, config)
	if err != nil {
		slog.Error("failed to update config", "err", err)
		return nil, err
	}
	return dto.MakePresetResponse(config), nil
}

func (s *DataMgmtService) UpdateConfig(ctx context.Context, req *dto.UpdatePresetRequest) (*dto.PresetResponse, error) {
	slog.Debug("updating config", "configId", req.PresetID)

	config := dto.MakeConfig(req)
	err := s.settingsRepo.UpdateConfig(ctx, config)
	if err != nil {
		slog.Error("failed to update preset", "err", err)
		return nil, err
	}
	return dto.MakePresetResponse(config), nil

}

func (s *DataMgmtService) DeleteConfig(ctx context.Context, configId string) error {

	err := s.settingsRepo.DeleteConfig(ctx, configId)
	if err != nil {
		slog.Error("failed to delete config", "err", err)
		return err
	}

	return nil
}

func (s *DataMgmtService) SetConfigState(ctx context.Context, configId string, enabled bool) error {

	slog.Debug("setting config state", "configId", configId, "enabled", enabled)
	config, err := s.settingsRepo.GetConfig(ctx, configId)
	if err != nil {
		slog.Error("failed to get config", "error", err)
		return err
	}
	if config == nil {
		slog.Error("config not found", "configId", configId)
		return errors.New("config not found")
	}
	config.Enabled = enabled
	err = s.settingsRepo.UpdateConfig(ctx, config)
	if err != nil {
		slog.Error("failed to update config state", "error", err)
		return err
	}
	return nil
}

func (s *DataMgmtService) GetAllConfigs(ctx context.Context) ([]*entity.Settings, error) {
	configs, err := s.settingsRepo.GetAllConfigs(ctx)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func generateConfigId() string {
	const (
		// Use characters that are safe for DNS labels
		charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		length  = 7
	)

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	result := make([]byte, length)
	for i := range b {
		result[i] = charset[int(b[i])%len(charset)]
	}
	return string(result)
}

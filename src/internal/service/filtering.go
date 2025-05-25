package service

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/quaintdev/webshield/src/internal/entity"
	"github.com/quaintdev/webshield/src/internal/repository"
)

type FilteringService struct {
	settingsRepo repository.SettingsRepository
	dnsRepo      repository.DomainDataRepository
}

func NewFilteringService(settings repository.SettingsRepository, dnsRepo repository.DomainDataRepository) *FilteringService {
	return &FilteringService{
		settingsRepo: settings,
		dnsRepo:      dnsRepo,
	}
}

func (s *FilteringService) IsDomainBlocked(ctx context.Context, settingId string, domainName string) (bool, error) {
	domainName = removeLastPeriod(domainName)
	config, err := s.settingsRepo.GetConfig(ctx, settingId)
	if err != nil {
		log.Println("dnsService.IsDomainBlocked: ", err)
		return false, err
	}

	if !config.Enabled {
		return false, nil
	}

	category := s.dnsRepo.GetDomainCategory(domainName)
	if category == "" {
		slog.Debug("domain not found in repository", "domainName", domainName)
		return false, nil
	}

	switch config.Categories[category] {
	case entity.Black:
		slog.Debug("domain is blocked", "domainName", domainName)
		return true, nil
	case entity.Blue:
		// config has allowed access to domain as per schedule
		now := time.Now().UTC()
		startTime := config.WeekDayScheduleMap[now.Weekday()].StartTime
		endTime := config.WeekDayScheduleMap[now.Weekday()].EndTime
		allowedStartTime := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.Local)
		allowedEndTime := time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.Local)

		// when time is outside allowed window
		if !(now.After(allowedStartTime) && now.Before(allowedEndTime)) {
			slog.Debug("domain is blocked as per schedule", "domainName", domainName)
			return true, nil
		}
	}

	return false, nil
}

func removeLastPeriod(s string) string {
	// Check if the string is empty or doesn't end with a period
	if len(s) == 0 || s[len(s)-1] != '.' {
		return s
	}

	// Return the string without the last character (the period)
	return s[:len(s)-1]
}

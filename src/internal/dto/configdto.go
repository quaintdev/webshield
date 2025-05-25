package dto

import (
	"log/slog"
	"time"

	"github.com/quaintdev/webshield/src/internal/entity"
)

type Category struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "allowed", "blocked", or "inactive"
}

type Schedule struct {
	Day       string `json:"day"`
	StartTime string `json:"startTime"` // Format: "HH:MM"
	EndTime   string `json:"endTime"`   // Format: "HH:MM"
}

type AddPresetRequest struct {
	Email string `json:"-"`
	ConfigFields
}
type PresetResponse struct {
	PresetID string `json:"id"`
	ConfigFields
}

type ConfigFields struct {
	PresetName string     `json:"name"`
	Enabled    bool       `json:"enabled"`
	UTCOffset  int        `json:"offset"`
	Categories []Category `json:"categories"`
	Schedule   []Schedule `json:"schedule"`
}

func MakePresetResponse(config *entity.Settings) *PresetResponse {
	response := new(PresetResponse)
	response.PresetName = config.Name
	response.PresetID = config.ID
	response.Enabled = config.Enabled
	for k, v := range config.Categories {
		var category Category
		category.Name = k
		switch v {
		case entity.White:
			category.Status = "inactive"
		case entity.Black:
			category.Status = "blocked"
		case entity.Blue:
			category.Status = "active"
		}
		response.Categories = append(response.Categories, category)
	}
	response.UTCOffset = config.UTCOffset
	for k, v := range config.WeekDayScheduleMap {
		var schedule Schedule
		schedule.Day = k.String()
		schedule.StartTime = v.StartTime.Add(-time.Duration(config.UTCOffset) * time.Minute).Format("15:04")
		schedule.EndTime = v.EndTime.Add(-time.Duration(config.UTCOffset) * time.Minute).Format("15:04")
		response.Schedule = append(response.Schedule, schedule)
	}
	return response
}

type UpdatePresetRequest struct {
	PresetID string `json:"-"`
	ConfigFields
}

func MakeConfig(req *UpdatePresetRequest) *entity.Settings {
	config := new(entity.Settings)
	config.Name = req.PresetName
	config.ID = req.PresetID
	config.Enabled = req.Enabled
	config.UTCOffset = req.UTCOffset
	config.Categories = make(map[string]entity.Category)
	config.WeekDayScheduleMap = make(map[time.Weekday]entity.Schedule)
	for _, v := range req.Categories {
		switch v.Status {
		case "inactive":
			config.Categories[v.Name] = entity.White
		case "blocked":
			config.Categories[v.Name] = entity.Black
		case "active":
			config.Categories[v.Name] = entity.Blue
		}
	}
	now := time.Now().UTC()
	for _, v := range req.Schedule {
		startHrMin, err := time.Parse("15:04", v.StartTime)
		if err != nil {
			slog.Error("Failed to parse start time", "startTime", v.StartTime)
			return nil
		}

		endHrMin, err := time.Parse("15:04", v.EndTime)
		if err != nil {
			slog.Error("Failed to parse end time", "endTime", v.EndTime)
			return nil
		}

		startTime := time.Date(now.Year(), now.Month(), now.Day(), startHrMin.Hour(), startHrMin.Minute(), 0, 0, time.UTC)
		startTime = startTime.Add(time.Duration(req.UTCOffset) * time.Minute)
		endTime := time.Date(now.Year(), now.Month(), now.Day(), endHrMin.Hour(), endHrMin.Minute(), 0, 0, time.UTC)
		endTime = endTime.Add(time.Duration(req.UTCOffset) * time.Minute)

		if endTime.Before(startTime) {
			endTime = endTime.AddDate(0, 0, 1)
		}

		config.WeekDayScheduleMap[convertDayStrToWeekday(v.Day)] = entity.Schedule{StartTime: startTime, EndTime: endTime}
	}
	return config
}

func convertDayStrToWeekday(day string) time.Weekday {
	var weekday time.Weekday
	switch day {
	case "Monday":
		weekday = time.Monday
	case "Tuesday":
		weekday = time.Tuesday
	case "Wednesday":
		weekday = time.Wednesday
	case "Thursday":
		weekday = time.Thursday
	case "Friday":
		weekday = time.Friday
	case "Saturday":
		weekday = time.Saturday
	case "Sunday":
		weekday = time.Sunday
	}
	return weekday
}

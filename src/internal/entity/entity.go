package entity

import "time"

type Category string

const (
	White Category = "white"
	Blue  Category = "blue"
	Black Category = "black"
)

type User struct {
	Email     string
	FirstName string
	LastName  string
	Configs   []string
}

type Settings struct {
	ID      string
	Name    string
	Enabled bool

	Categories map[string]Category

	WeekDayScheduleMap map[time.Weekday]Schedule
	UTCOffset          int
}

type Schedule struct {
	StartTime time.Time
	EndTime   time.Time
}

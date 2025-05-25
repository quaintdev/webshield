package dto

import (
	"github.com/quaintdev/webshield/src/internal/entity"
	"reflect"
	"testing"
	"time"
)

func TestMakePresetResponse(t *testing.T) {
	type args struct {
		config *entity.Settings
	}
	tests := []struct {
		name string
		args args
		want *PresetResponse
	}{
		{
			name: "",
			args: args{
				config: &entity.Settings{
					WeekDayScheduleMap: make(map[time.Weekday]entity.Schedule),
					Categories:         make(map[string]entity.Category),
					Name:               "test",
					ID:                 "test",
					UTCOffset:          0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakePresetResponse(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakePresetResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

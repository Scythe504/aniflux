package utils

import (
	"testing"
	"time"
)

func TestGet30HourSchedule(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)

	tests := []struct {
		name       string
		inputTime  time.Time
		wantDay    string
		wantTiming string
	}{
		{
			name:       "Tuesday 01:30 AM JST -> Monday 25:30",
			inputTime:  time.Date(2026, 8, 11, 1, 30, 0, 0, jst),
			wantDay:    "Monday",
			wantTiming: "25:30",
		},
		{
			name:       "Tuesday 05:59 AM JST -> Monday 29:59",
			inputTime:  time.Date(2026, 8, 11, 5, 59, 59, 0, jst),
			wantDay:    "Monday",
			wantTiming: "29:59",
		},
		{
			name:       "Tuesday 06:00 AM JST -> Tuesday 06:00",
			inputTime:  time.Date(2026, 8, 11, 6, 0, 0, 0, jst),
			wantDay:    "Tuesday",
			wantTiming: "06:00",
		},
		{
			name:       "Monday 00:00 AM JST -> Sunday 24:00",
			inputTime:  time.Date(2026, 8, 10, 0, 0, 0, 0, jst),
			wantDay:    "Sunday",
			wantTiming: "24:00",
		},
		{
			name:       "Sunday 12:00 PM JST -> Sunday 12:00",
			inputTime:  time.Date(2026, 8, 9, 12, 0, 0, 0, jst),
			wantDay:    "Sunday",
			wantTiming: "12:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDay, gotTiming := Get30HourSchedule(tt.inputTime.Unix())
			if gotDay != tt.wantDay {
				t.Errorf("Get30HourSchedule() gotDay = %v, want %v", gotDay, tt.wantDay)
			}
			if gotTiming != tt.wantTiming {
				t.Errorf("Get30HourSchedule() gotTiming = %v, want %v", gotTiming, tt.wantTiming)
			}
		})
	}
}

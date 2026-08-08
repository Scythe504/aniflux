package utils

import (
	"fmt"
	"time"
)

// Get30HourSchedule converts a Unix timestamp (seconds) into its 30-hour schedule representation.
// It uses JST (Japan Standard Time) as the reference timezone.
// If the JST hour is between 00:00:00 and 05:59:59, it belongs to the previous calendar day
// (e.g. Tuesday becomes Monday) and the hour is incremented by 24 (e.g. 01:30 becomes 25:30).
func Get30HourSchedule(airingAt int64) (string, string) {
	jst := time.FixedZone("JST", 9*60*60)
	t := time.Unix(airingAt, 0).In(jst)

	hour := t.Hour()
	minute := t.Minute()

	var scheduleDay time.Weekday
	var virtualHour int

	if hour < 6 {
		// Belongs to the previous calendar day
		scheduleDay = t.AddDate(0, 0, -1).Weekday()
		virtualHour = hour + 24
	} else {
		scheduleDay = t.Weekday()
		virtualHour = hour
	}

	timing := fmt.Sprintf("%02d:%02d", virtualHour, minute)
	return scheduleDay.String(), timing
}

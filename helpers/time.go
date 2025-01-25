package helpers

import "time"

// ConvertToAsiaMumbaiTime converts a time.Time from UTC to Asia/Mumbai time zone.
// It returns the converted time.Time and an error if any.
//
// Parameters:
//   - utcTime: time.Time
//
// Returns:
//   - time.Time: Converted time.Time
func ConvertToAsiaMumbaiTime(utcTime time.Time) time.Time {
	location, _ := time.LoadLocation("Asia/Kolkata")
	return utcTime.In(location)
}

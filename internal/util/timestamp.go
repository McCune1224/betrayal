package util

import "time"

// Helper to get current time in EST
func GetEstTimeStamp() string {
	// get current time in est
	est := time.Now().UTC().Add(-4 * time.Hour)

	// format similar to Oct 24 11:00:00
	// without the est at the end
	return est.Format("Jan 2 15:04:05")
}

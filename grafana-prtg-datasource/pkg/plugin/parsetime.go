package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func parsePRTGDateTime(datetime string) (time.Time, string, error) {
	// Remove any whitespace
	datetime = strings.TrimSpace(datetime)

	// If datetime contains a range (e.g., "06.03.2025 15:11:00 - 15:12:00")
	if strings.Contains(datetime, " - ") {
		parts := strings.Split(datetime, " - ")
		datePart := strings.TrimSpace(strings.Split(parts[0], " ")[0])
		startTime := strings.TrimSpace(strings.Split(parts[0], " ")[1])
		endTime := strings.TrimSpace(parts[1])

		// Construct the full datetime string with end time
		datetime = datePart + " " + endTime

		// Parse start time to compare
		startTimeStr := datePart + " " + startTime
		loc, _ := time.LoadLocation("Europe/Berlin")
		startDateTime, err := time.ParseInLocation("02.01.2006 15:04:05", startTimeStr, loc)
		if err == nil {
			endDateTime, err := time.ParseInLocation("02.01.2006 15:04:05", datetime, loc)
			if err == nil && endDateTime.Before(startDateTime) {
				// If end time is before start time, add one day
				datetime = endDateTime.AddDate(0, 0, 1).Format("02.01.2006 15:04:05")
			}
		}
	}

	// PRTG sends times in local timezone
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		loc = time.Local
	}

	// Try multiple formats
	layouts := []string{
		"02.01.2006 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	var lastErr error
	for _, layout := range layouts {
		parsedTime, err := time.ParseInLocation(layout, datetime, loc)
		if err == nil {
			// Convert to UTC for consistency
			utcTime := parsedTime.UTC()
			return utcTime, strconv.FormatInt(utcTime.Unix(), 10), nil
		}
		lastErr = err
	}

	// Log the parsing failure
	backend.Logger.Error("Failed to parse datetime",
		"input", datetime,
		"error", lastErr,
	)

	return time.Time{}, "", fmt.Errorf("failed to parse datetime '%s': %v", datetime, lastErr)
}

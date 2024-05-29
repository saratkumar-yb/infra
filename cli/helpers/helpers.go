package helpers

import (
	"fmt"
	"strings"
)

type Schedule struct {
	InstanceID   string `json:"instance_id"`
	StartTime    string `json:"start_time"`
	StopTime     string `json:"stop_time"`
	Timezone     string `json:"timezone"`
	AWSRegion    string `json:"aws_region"`
	FriendlyName string `json:"friendly_name"`
}

type CloudProvider interface {
	CreateSchedule(instanceID, region, startTime, stopTime, timezone, friendlyName string)
	ListSchedules()
	DeleteSchedule(instanceID string)
}

var timezoneMapping = map[string]string{
	"EST":  "America/New_York",
	"PST":  "America/Los_Angeles",
	"CST":  "America/Chicago",
	"MST":  "America/Denver",
	"IST":  "Asia/Kolkata",
	"BST":  "Europe/London",
	"CET":  "Europe/Berlin",
	"EET":  "Europe/Helsinki",
	"GMT":  "Europe/London",
	"HST":  "Pacific/Honolulu",
	"AKST": "America/Anchorage",
	"AST":  "America/Halifax",
	"NST":  "America/St_Johns",
	"PDT":  "America/Los_Angeles",
	"EDT":  "America/New_York",
	"CDT":  "America/Chicago",
	"MDT":  "America/Denver",
	"CEST": "Europe/Berlin",
	"EEST": "Europe/Helsinki",
	"WET":  "Europe/Lisbon",
	"WEST": "Europe/Lisbon",
}

func ResolveTimezone(abbreviation string) (string, error) {
	if tz, exists := timezoneMapping[strings.ToUpper(abbreviation)]; exists {
		return tz, nil
	}
	return "", fmt.Errorf("unsupported timezone abbreviation: %s", abbreviation)
}

package sak

import (
	"regexp"
	"time"
)

type L struct {
	F string // "facility" equivalent/tag
	S string // "severity" equivalent
	C string // error code/key string
}

type Options struct {
	DebugLevel int
	MaxLogHist int64
	Behavior   struct {
		PrintTime      bool
		TimeMilli      bool
		Filter         []string
		LogShiftBuffer int
		filterRegexp   *regexp.Regexp
	}
}

type LogEntry struct {
	t         time.Time `json:"time"`
	Level     int       `json:"logLevel"`
	Facility  string    `json:"facility"`
	Severity  string    `json:"severity"`
	Code      string    `json:"code"`
	Msg       string    `json:"message"`
	OutputStr string    `json:"outputstr"`
	Printed   bool      `json:"printed"`
}

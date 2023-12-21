package sak

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	spew "github.com/davecgh/go-spew/spew"
)

const version = "1.0.14"

var (
	Opts    = Options{}
	LogHist []LogEntry
)

//  can be used for program output; specify n = 0 and no facility

func LOG(n int, msgs ...interface{}) {
	var (
		timeStr      string = ""
		lStr         string = ""
		fStr         string = ""
		now                 = time.Now()
		nowNano             = now.UnixNano()
		nowMilli            = nowNano / 1000000
		nowMilli_str        = strconv.FormatInt(nowMilli, 10)
		now_str             = strconv.FormatInt(nowMilli/1000, 10)
		logOpts             = L{}
	)

	if Opts.DebugLevel < 0 {
		Opts.DebugLevel = 0
	} else {
		// if we're not keeping the log history, and the message is higher than the debug level, nobody is ever gonna see it, so skip it
		if Opts.MaxLogHist == 0 && (n > Opts.DebugLevel) {
			return
		}
	}

	if Opts.MaxLogHist == 0 { // truncate it
		LogHist = []LogEntry{}
	}

	// this is how far we shift the log buffer back when we get to the end
	if Opts.Behavior.LogShiftBuffer <= 0 {
		Opts.Behavior.LogShiftBuffer = 10
	}

	override := os.Getenv("SAK_LOG_DLOVERRIDE")
	if override != "" {
		// we have an override to the debugging levels; force it to this value regardless of any other settings
		i, err := strconv.ParseInt(override, 10, 0)
		if err == nil {
			Opts.DebugLevel = int(i)
		}
	}

	filter := os.Getenv("SAK_LOG_FFILTER")
	if filter != "" && Opts.Behavior.filterRegexp == nil {
		// we have a filter request; parse it and then filter output against it
		// filter format is regexp in format specified in https://github.com/google/re2/wiki/Syntax
		f, err := regexp.Compile(filter)
		if err == nil {
			Opts.Behavior.filterRegexp = f
		}
	}

	ltmp := LogEntry{
		t:        now,
		Level:    n,
		Facility: "",
		Severity: "",
		Code:     "",
	}

	if Opts.Behavior.PrintTime {
		if Opts.Behavior.TimeMilli {
			timeStr = nowMilli_str
		} else {
			timeStr = now_str
		}
		timeStr += ": "
	}

	for m := range msgs {
		switch msgs[m].(type) {
		case L:
			logOpts = msgs[m].(L)
		case string:
			ltmp.Msg += msgs[m].(string)
		case int:
			ltmp.Msg += strconv.FormatInt(int64(msgs[m].(int)), 10)
		case int64:
			ltmp.Msg += strconv.FormatInt(msgs[m].(int64), 10)
		case float64:
			ltmp.Msg += strconv.FormatFloat(msgs[m].(float64), 'E', -1, 64)
		case error:
			ltmp.Msg += msgs[m].(error).Error()
		default:
			ltmp.Msg += spew.Sdump(msgs[m])
		}
	}

	if logOpts.F != "" {
		ltmp.Facility = logOpts.F
	}
	if logOpts.S != "" {
		ltmp.Severity = logOpts.S
	}
	if logOpts.C != "" {
		ltmp.Code = logOpts.C
	}

	if n > 0 {
		lStr = strconv.Itoa(n) + ": "
	}

	if ltmp.Facility != "" && n > 0 {
		fStr = "[" + ltmp.Facility + "]: "
	}

	ltmp.OutputStr = fmt.Sprintf("%s%s%s%s", lStr, timeStr, fStr, ltmp.Msg)
	if Opts.DebugLevel >= n {
		// logic:  logging at level 0 is equivalent to "output", so print in that case;
		//         logging at appropriate level with no filter in operation, print in that case;
		//         logging at appropriate level with a filter present and a matching facility, print in THAT case.
		if n == 0 || Opts.Behavior.filterRegexp == nil || Opts.Behavior.filterRegexp.MatchString(ltmp.Facility) {
			fmt.Printf("%s\n", ltmp.OutputStr)
			ltmp.Printed = true
		}
	}

	if Opts.MaxLogHist > 0 && int64(len(LogHist)) > Opts.MaxLogHist {
		LogHist = LogHist[int64(len(LogHist)+Opts.Behavior.LogShiftBuffer)-Opts.MaxLogHist:]
	}
	if Opts.MaxLogHist != 0 {
		LogHist = append(LogHist, ltmp)
	}

	// clean up
	logOpts = L{}
	ltmp = LogEntry{}
}

package sak

import (
	"fmt"
	"testing"
)

func TestLogZero(t *testing.T) {
	testMatch := "testing the log at level zero and no facility"
	Opts.DebugLevel = 0
	Opts.MaxLogHist = 10
	LOG(0, L{}, testMatch)
	if LogHist[len(LogHist)-1].OutputStr != testMatch {
		t.Errorf("error: log at level zero expected '" + testMatch + "', got '" + LogHist[len(LogHist)-1].OutputStr)
	}
}

func TestLogOne(t *testing.T) {
	testMatch := "testing a log entry at level 1 and no facility"
	Opts.DebugLevel = 1
	Opts.MaxLogHist = 10
	LOG(1, L{}, testMatch)
	if LogHist[len(LogHist)-1].OutputStr != "1: "+testMatch {
		t.Errorf("error: log at level zero expected '" + testMatch + "', got '" + LogHist[len(LogHist)-1].OutputStr)
	}
}

func TestLogOneF(t *testing.T) {
	testMatch := "testing a log entry at level 1 and sample facility"
	Opts.DebugLevel = 1
	Opts.MaxLogHist = 10
	LOG(1, L{F: "testfac"}, testMatch)
	if LogHist[len(LogHist)-1].OutputStr != "1: [testfac]: "+testMatch {
		t.Errorf("error: log at level zero expected '" + testMatch + "', got '" + LogHist[len(LogHist)-1].OutputStr)
	}
}

func TestLogOneDLvl(t *testing.T) {
	testMatch := "testing a log entry at level 1 and sample facility when debug level is lower"
	Opts.DebugLevel = 0
	Opts.MaxLogHist = 10
	LOG(1, L{F: "testfac"}, testMatch)
	if LogHist[len(LogHist)-1].Printed {
		t.Errorf("error: expected Printed=false")
	}
}

func TestLogCode(t *testing.T) {
	testMatch := "testing a log entry with a code attached"
	matchStr := "1: [testfac]: <C@3919> " + testMatch
	Opts.DebugLevel = 1
	Opts.MaxLogHist = 10
	LOG(1, L{F: "testfac", C: "3919"}, testMatch)
	if LogHist[len(LogHist)-1].OutputStr != matchStr {
		t.Errorf("error: log expected printed line '" + matchStr + "', got '" + LogHist[len(LogHist)-1].OutputStr + "'")
	}
}

type TestStruct struct {
	Str1 string
	Str2 string
	Int  int
}

func (ts TestStruct) Logify() string {
	return fmt.Sprintf("str1: %s, str2: %s, int: %d", ts.Str1, ts.Str2, ts.Int)
}

func TestLogify(t *testing.T) {
	testMatch := "testing a log entry with Logify() defined"
	var ts TestStruct
	ts.Str1 = "str1"
	ts.Str2 = "str2"
	ts.Int = 10
	Opts.DebugLevel = 1
	Opts.MaxLogHist = 10
	matchStr := "1: [testfac]: testing a log entry with Logify() defined - str1: str1, str2: str2, int: 10"
	LOG(1, L{F: "testfac"}, testMatch, " - ", ts)
	if LogHist[len(LogHist)-1].OutputStr != matchStr {
		t.Errorf("error: log expected printed line '" + matchStr + "', got '" + LogHist[len(LogHist)-1].OutputStr + "'")
	}
}

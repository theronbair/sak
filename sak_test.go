package sak

import (
   "testing"
)

func TestLogZero(t *testing.T) {
   testMatch := "testing the log at level zero and no facility"
   Opts.DebugLevel = 0
   LOG(0, L{}, testMatch)
   if ( LogHist[len(LogHist)-1].OutputStr != testMatch ) {
      t.Errorf("error: log at level zero expected '"+testMatch+"', got '"+LogHist[len(LogHist)-1].OutputStr)
   }
}

func TestLogOne(t *testing.T) {
   testMatch := "testing a log entry at level 1 and no facility"
   Opts.DebugLevel = 1
   LOG(1, L{}, testMatch)
   if ( LogHist[len(LogHist)-1].OutputStr != "1: "+testMatch ) {
      t.Errorf("error: log at level zero expected '"+testMatch+"', got '"+LogHist[len(LogHist)-1].OutputStr)
   }
}

func TestLogOneF(t *testing.T) {
   testMatch := "testing a log entry at level 1 and sample facility"
   Opts.DebugLevel = 1
   LOG(1, L{F: "testfac"}, testMatch)
   if ( LogHist[len(LogHist)-1].OutputStr != "1: [testfac]: "+testMatch ) {
      t.Errorf("error: log at level zero expected '"+testMatch+"', got '"+LogHist[len(LogHist)-1].OutputStr)
   }
}

func TestLogOneDLvl(t *testing.T) {
   testMatch := "testing a log entry at level 1 and sample facility when debug level is lower"
   Opts.DebugLevel = 0
   LOG(1, L{F: "testfac"}, testMatch)
   if ( LogHist[len(LogHist)-1].Printed ) {
      t.Errorf("error: expected Printed=false")
   }
}

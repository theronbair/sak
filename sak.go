package sak

import (
   "fmt"
   "time"
   "strconv"
   spew "github.com/davecgh/go-spew/spew"
)

const version = "1.0.10"

type L struct {
   F string       // "facility" equivalent
   S string       // "severity" equivalent
   C string       // error code/key string
}

type Options struct {
   DebugLevel int
   Behavior struct {
      PrintTime bool
      TimeMilli bool
   }
}

type LogEntry struct {
   t time.Time
   Level int
   Facility string
   Severity string
   Code string
   Msg string
   OutputStr string
   Printed bool
}

var (
   Opts = Options{}
   LogHist []LogEntry
)

//  can be used for program output; specify n = 0 and no facility

func LOG(n int, msgs ...interface{}) {
   var (
      ltmp LogEntry
      timeStr string = ""
      lStr string = ""
      fStr string = ""
      now = time.Now()
      nowNano = now.UnixNano()
      nowMilli = nowNano / 1000000
      nowMilli_str = strconv.FormatInt(nowMilli, 10)
      now_str = strconv.FormatInt(nowMilli / 1000, 10)
      logOpts = L{} 
   )

   ltmp.t = now
   ltmp.Level = n
   ltmp.Facility = ""
   ltmp.Severity = ""
   ltmp.Code = ""

   if ( Opts.Behavior.PrintTime ) {
      if ( Opts.Behavior.TimeMilli ) {
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

   ltmp.Facility = logOpts.F
   ltmp.Severity = logOpts.S
   ltmp.Code = logOpts.C

   if ( n > 0 ) {
      lStr = strconv.Itoa(n) + ": "
   }

   if ( logOpts.F != "" ) {
      fStr = "[" + logOpts.F + "]: "
   }

   ltmp.OutputStr = fmt.Sprintf("%s%s%s%s", lStr, timeStr, fStr, ltmp.Msg)
   if ( Opts.DebugLevel >= n ) {
      fmt.Printf("%s\n", ltmp.OutputStr)
      ltmp.Printed = true
   }
   LogHist = append(LogHist, ltmp)
}

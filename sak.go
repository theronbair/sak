package sak

import (
   "fmt"
   "time"
   "strconv"
   spew "github.com/davecgh/go-spew/spew"
)

const version = "1.0.2"

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

var (
   opts = Options{}
   now = time.Now()
   nowNano = now.UnixNano()
   nowMilli = nowNano / 1000000
   nowMilli_str = strconv.FormatInt(nowMilli, 10)
   now_str = strconv.FormatInt(nowMilli / 1000, 10)
)

// can be used for program output; specify n = 0 and no facility

func LOG(n int, logopts L, msgs ...interface{}) {
   if ( opts.DebugLevel >= n ) {
      if ( n > 0 ) {
         fmt.Printf("%d:", n)
      }
      if ( opts.Behavior.PrintTime ) {
         if ( opts.Behavior.TimeMilli ) {
            fmt.Printf("%s:", nowMilli_str)
         } else {
            fmt.Printf("%s", now_str)
         }
      }
      if ( logopts.F != "" ) {
         fmt.Printf("%s:", logopts.F)
      }
      if ( n > 0 || logopts.F != "" ) {
         fmt.Printf(" ")
      }
      for m := range msgs {
         switch msgs[m].(type) {
            case string:
               fmt.Printf("%s", msgs[m].(string))
            case error:
               fmt.Printf("%s", msgs[m].(error).Error())
            default:
               fmt.Printf("%s", spew.Sdump(msgs[m]))
         }
      }
   }
}

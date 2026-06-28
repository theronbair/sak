/*

sak, Theron Bair, (c) 2026
Licensed under the BSD 3-clause license (see LICENSE)

This package is designed to be (almost) the simplest conceivable logging mechanism that offers reasonable functionality.

WYSIWYG.

Import it, then just call the logging function:

		sak.LOG(1, sak.L{F:"whatever"}, "this is my log!")

If you don't have anything extra to put in it (facility, severity, code, whatever), you can skip that middle piece:

		sak.LOG(1, "this is my log too!")

... and it will print nothing at all before that log line (other than the log level).  Simple, no?

This is supposed to be simple and "cheap", in terms of memory and compute, in default operation.  It's not quite
as cheap as just fmt.Println, but I strive to be close.

It is probably not as thread-safe as it could be.  Some efforts have been made.  (h/t ck)

As I find ways to make it simpler and cheaper, I'll implement them.  Ideally it would cost "almost nothing" to log
a wide variety of stuff at a number of (dynamically-settable) levels, enabling you to figure out what the hell is
going on more easily.

And isn't that what logging is all about?

Things to like about it:
 • It requires only "import" and then call the function.  No setup necessary.  There are a couple of behavior knobs in options, though.
 • It assumes that any logs at a level >1 are going to stderr.  If you want it going somewhere else, too bad.  Redirect stderr.
 • It can, in fact, be used for output at a log level of "0" (which is semantically equivalent to 'always print').  This goes to stdout.
 • It does not offer a multiplicity of formats.  The format is as follows:
   X: (optional time:) [facility/severity] (optional code) <log>
   where X is the numeric logging level.  (Higher = more detailed)
 • It's meant to be as greppable as possible; if you want everything from log level 4, "^4:" is your friend.
   CAVEAT: spew will dump stuff in multiline.
 • If you stuff something into the logging function, it will attempt to print that thing out in a reasonable fashion.
   As a last resort, it will feed it to "spew".
   CAVEAT:  No limits are applied to what it will spew out.  If it's a 1GB struct, it'll dump the whole thing out.
   Be careful what you feed it.

Things to know about it:
 • If the type exposes a Logify() function, this will call that function to prettyprint a struct, redact things, whatever.
 • It can be used for program output; specify n = 0 and no facility and it will just do straight output.
 • Facility/Severity have the same approximate meaning as in syslog, though honestly, these are more "tags" than anything else.
 • "Code" is meant to be an unambiguous error code which can be searched on in the output.  Probably you won't use it.
 • It does have a rolling log buffer.  This is OFF by default.  However, you can turn it on (if you would like) and use that to
   store older logs, maybe ones that weren't printed (so you can dump them out later via some mechanism if needed).
 • That log buffer is a quasi-ring buffer that "shifts" forward by LogHistBuffer lines at a time, so as to mitigate some of the
   penalty of constant slice mutation.  If you are expecting one-line-at-a-time rolling... you have to set it.
 • The log buffer is not *guaranteed* to contain a log line.  That depends on whether you consider it acceptable to block on the
   buffer mutation mutex or not.  If blocking is NOT acceptable, and you call this in multiple threads, then you may lose lines.
   If you decide you do not want to lose lines, then you are going to block on calls to LOG().  This is the BlockForBufferMutex
   flag.

Things to hate about it:
 • You're not going to instantiate a "new logger object" or whatever.  You just call the function.
   Therefore, you will need to preface it with the package name.  Every time.
 • This goes for the L{} type as well, containing things like "facility", "severity", "code", etc.  Tedious, but not THAT bad.
   (I kept the package name short for that reason.)  Therefore, if you want to get rid of all of the logging stuff, just strip
   out "sak" (or, if you are paranoid, "sak.LOG") from your code and voila!  It's all gone.
 • It is not meant to be, and is not, compatible in any way (backward-, forward-, or -otherwise) with any other logging package.
   It is not a drop-in replacement for anything.  It is its own thing.  If you want the other thing, use the other thing.
 • It has made most of your decisions for you.  I think they're good ones.  You may think differently.
   The fork button is right over there :D --->

*/

package sak

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	spew "github.com/davecgh/go-spew/spew"
)

const version = "1.0.17"

var (
	Opts      = Options{}
	histMutex sync.Mutex
	LogHist   []LogEntry

	// atomic copies of hot-path config to avoid locking on reads
	debugLevel   atomic.Int32
	maxLogHist   atomic.Int64
	filterRegexp atomic.Pointer[regexp.Regexp]
)

// SetDebugLevel updates the debug level atomically.
func SetDebugLevel(n int) {
	debugLevel.Store(int32(n))
	Opts.DebugLevel = n
}

// SetMaxLogHist updates the max log history atomically.
// Setting to 0 disables the buffer; setting to -1 means unlimited.
func SetMaxLogHist(n int64) {
	maxLogHist.Store(n)
	Opts.MaxLogHist = n
	if n == 0 {
		histMutex.Lock()
		LogHist = []LogEntry{}
		histMutex.Unlock()
	}
}

func LOG(n int, msgs ...interface{}) {
	// first:  do a couple basic checks to make this cheap if we can
	// pull from atomics if set, otherwise fall back to Opts (backward compat with direct assignment)
	dl := int(debugLevel.Load())
	if dl == 0 && Opts.DebugLevel != 0 {
		dl = Opts.DebugLevel
	}
	mh := maxLogHist.Load()
	if mh == 0 && Opts.MaxLogHist != 0 {
		mh = Opts.MaxLogHist
	}

	// there are some intriguing semantics of making the debugging level negative, but that comes later; if negative, fix it
	if dl < 0 {
		SetDebugLevel(0)
		dl = 0
	}

	// if we're not keeping the log history, and the message is higher than the debug level, nobody is ever gonna see it, so save ourselves some cycles and return right now
	if mh == 0 && n > dl {
		return
	}

	var (
		timeStr      string = ""
		lStr         string = ""
		fStr         string = ""
		cStr         string = ""
		now                 = time.Now()
		nowNano             = now.UnixNano()
		nowMilli            = nowNano / 1000000
		nowMilli_str        = strconv.FormatInt(nowMilli, 10)
		now_str             = strconv.FormatInt(nowMilli/1000, 10)
		logOpts             = L{}
	)

	// this is how far we shift the log buffer back when we get to the end
	shiftBuf := Opts.Behavior.LogShiftBuffer
	if shiftBuf <= 0 {
		shiftBuf = 10
	}

	override := os.Getenv("SAK_LOG_DLOVERRIDE")
	if override != "" {
		// we have an override to the debugging levels; force it to this value regardless of any other settings
		i, err := strconv.ParseInt(override, 10, 0)
		if err == nil {
			SetDebugLevel(int(i))
			dl = int(i)
		} else {
			SetDebugLevel(0)
			dl = 0
		}
	}

	filter := os.Getenv("SAK_LOG_FFILTER")
	if filter != "" && filterRegexp.Load() == nil {
		// we have a filter request; parse it and then filter output against it
		// filter format is regexp in format specified in https://github.com/google/re2/wiki/Syntax
		f, err := regexp.Compile(filter)
		if err == nil {
			filterRegexp.Store(f)
			Opts.Behavior.filterRegexp = f
		}
	}

	ltmp := LogEntry{
		Time:     now,
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

	for _, m := range msgs {
		// if this thing exposes Logify(), use that instead
		var i interface{} = m
		if v, ok := i.(LogInterface); ok {
			ltmp.Msg += v.Logify()
		} else {
			switch m.(type) {
			// if this is an options struct, glom onto it and use it; last one wins
			case L:
				logOpts = m.(L)
			// otherwise, print it out as the type that it is
			case string:
				ltmp.Msg += m.(string)
			case int, uint, int64, uint64:
				ltmp.Msg += fmt.Sprintf("%d", m)
			case float32, float64:
				ltmp.Msg += fmt.Sprintf("%f", m)
			case error:
				ltmp.Msg += m.(error).Error()
			default:
				ltmp.Msg += spew.Sdump(m)
			}
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
		fStr = "[" + ltmp.Facility
		if ltmp.Severity != "" {
			fStr = fStr + "/" + ltmp.Severity
		}
		fStr = fStr + "]: "
	}

	if ltmp.Code != "" {
		cStr = "<C@" + ltmp.Code + "> "
	}

	ltmp.OutputStr = fmt.Sprintf("%s%s%s%s%s", lStr, timeStr, fStr, cStr, ltmp.Msg)

	if dl >= n {
		// logic:  logging at level 0 is equivalent to "output", so print in that case;
		//         logging at appropriate level with no filter in operation, print in that case;
		//         logging at appropriate level with a filter present and a matching facility, print in THAT case.
		filt := filterRegexp.Load()
		if n == 0 {
			fmt.Fprintf(os.Stdout, "%s\n", ltmp.OutputStr)
			ltmp.Printed = true
		} else {
			if filt == nil || filt.MatchString(ltmp.Facility) {
				fmt.Fprintf(os.Stderr, "%s\n", ltmp.OutputStr)
				ltmp.Printed = true
			}
		}
	}

	// juggle the log buffer if needed
	if mh != 0 {
		if Opts.BlockForBufferMutex {
			// caller has decided they want guaranteed buffer entries; block if needed
			histMutex.Lock()
			if mh > 0 && int64(len(LogHist)) > mh {
				LogHist = LogHist[int64(len(LogHist)+shiftBuf)-mh:]
			}
			LogHist = append(LogHist, ltmp)
			histMutex.Unlock()
		} else {
			// non-blocking: try to grab the lock, but if someone else has it, drop the buffer entry
			// the log was still printed; we just lose it from the in-memory history
			if histMutex.TryLock() {
				if mh > 0 && int64(len(LogHist)) > mh {
					LogHist = LogHist[int64(len(LogHist)+shiftBuf)-mh:]
				}
				LogHist = append(LogHist, ltmp)
				histMutex.Unlock()
			}
		}
	}

	// clean up
	logOpts = L{}
	ltmp = LogEntry{}
}

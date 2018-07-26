package govec

import (
	"fmt"
	"time"

	"github.com/daviddengcn/go-colortext"
)

type LogPriority int

//LogPriority enum provides all the valid Priority Levels that can be
//used to log events with.
const (
	DEBUG LogPriority = iota
	NORMAL
	NOTICE
	WARNING
	ERROR
	CRITICAL
)

//The GoPriorityLog struct provides an interface to creating and
//maintaining vector timestamp entries in the generated log file as well
//as priority to local events which are printed out to console. Local
//events with a priority equal to or higher are logged in the log file.
type GoPriorityLog struct {
	GoLog
	Priority LogPriority
}

//Returns a GoPriorityLog Struct taking in three arguments and
//truncates previous logs:
//* MyProcessName (string): local process name; must be unique in your
//distributed system.
//* LogFileName (string) : name of the log file that will store info.
//Any old log with the same name will be truncated
//* Priority (LogPriority) : priority which decides what future local
//events should be logged in the log file. Any local event with a
//priority level equal to or higher than this will be logged in the
//log file. This priority can be changed using SetPriority.
func InitGoVectorPriority(processid string, logfilename string, Priority LogPriority) *GoPriorityLog {
	gv := InitGoVector(processid, logfilename)
	return &GoPriorityLog{*gv, Priority}
}

func (l LogPriority) getColor() ct.Color {
	var color ct.Color
	switch l {
	case DEBUG:
		color = ct.Green
	case NORMAL:
		color = ct.White
	case NOTICE:
		color = ct.Cyan
	case WARNING:
		color = ct.Yellow
	case ERROR:
		color = ct.Red
	case CRITICAL:
		color = ct.Magenta
	default:
		color = ct.None
	}
	return color
}

func (l LogPriority) getPrefixString() string {
	var prefix string
	switch l {
	case DEBUG:
		prefix = "DEBUG"
	case NORMAL:
		prefix = "NORMAL"
	case NOTICE:
		prefix = "NOTICE"
	case WARNING:
		prefix = "WARNING"
	case ERROR:
		prefix = "ERROR"
	case CRITICAL:
		prefix = "CRITICAL"
	default:
		prefix = ""
	}
	return prefix
}

func (gv *GoPriorityLog) printColoredMessage(LogMessage string, Priority LogPriority) {
	color := Priority.getColor()
	prefix := Priority.getPrefixString()
	ct.Foreground(color, true)
	fmt.Print(time.Now().String() + ":" + prefix + "-")
	ct.ResetColor()
	fmt.Println(LogMessage)
}

//Sets the priority which is used to decide which future local events
//should be logged in the log file. Any future local event with a
//priority level equal to or higher than this will be logged in the log
//file.
func (gv *GoPriorityLog) SetPriority(Priority LogPriority) {
	gv.Priority = Priority
}

//If the priority of the logger is lower than or equal to the priority
//of this event then the current vector timestamp is incremented and the
//message is logged it into the Log File. A color coded string is also
//printed on the console.
//* LogMessage (string) : Message to be logged
//* Priority (LogPriority) : Priority at which the message is to be logged
func (gv *GoPriorityLog) LogLocalEventWithPriority(LogMessage string, Priority LogPriority) {
	if Priority >= gv.Priority {
		prefix := Priority.getPrefixString() + " - "
		gv.LogLocalEvent(prefix + LogMessage)
	}
	gv.printColoredMessage(LogMessage, Priority)
}

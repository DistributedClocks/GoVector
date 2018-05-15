package govec

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"time"
)

type LogPriority int

const (
	DEBUG LogPriority = iota
	NORMAL
	NOTICE
	WARNING
	ERROR
	CRITICAL
)

type GoPriorityLog struct {
	GoLog
	Priority LogPriority
}

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

func (gv *GoPriorityLog) SetPriority(Priority LogPriority) {
	gv.Priority = Priority
}

func (gv *GoPriorityLog) LogLocalEventWithPriority(LogMessage string, Priority LogPriority) {
	if Priority >= gv.Priority {
		gv.LogLocalEvent(LogMessage)
	}
	gv.printColoredMessage(LogMessage, Priority)
}

package main

import (
	"github.com/DistributedClocks/GoVector/govec"
)

func main() {
	Logger := govec.InitGoVectorPriority("MyProcess", "PrioritisedLogFile", govec.NORMAL)

	Logger.LogLocalEventWithPriority("Debug Priority Event", govec.DEBUG)
	Logger.LogLocalEventWithPriority("Normal Priority Event", govec.NORMAL)
	Logger.LogLocalEventWithPriority("Notice Priority Event", govec.NOTICE)
	Logger.LogLocalEventWithPriority("Warning Priority Event", govec.WARNING)
	Logger.LogLocalEventWithPriority("Error Priority Event", govec.ERROR)
	Logger.LogLocalEventWithPriority("Critical Priority Event", govec.CRITICAL)

	Logger.DisableLogging()
}
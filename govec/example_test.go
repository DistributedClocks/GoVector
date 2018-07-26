package govec_test

import (
	"fmt"

	"github.com/DistributedClocks/GoVector/govec"
)

func ExampleGoLog() {
	Logger := govec.InitGoVector("MyProcess", "LogFile")

	//In Sending Process

	//Prepare a Message
	messagepayload := []byte("samplepayload")
	finalsend := Logger.PrepareSend("Sending Message", messagepayload)

	//In Receiving Process

	//receive message
	var incommingMessage []byte
	Logger.UnpackReceive("Received Message from server", finalsend, &incommingMessage)
	fmt.Printf("Received Message: %s\n", incommingMessage)
	//Can be called at any point
	Logger.LogLocalEvent("Example Complete")

	Logger.DisableLogging()
	//No further events will be written to log file

	// Output: Received Message: samplepayload
}

func ExampleGoTSLog() {
	Logger := govec.InitGoVectorTimeStamp("MyProcess", "LogFile")

	//In Sending Process

	//Prepare a Message
	messagepayload := []byte("samplepayload")
	finalsend := Logger.PrepareSend("Sending Message", messagepayload)
	//In Receiving Process

	//receive message
	var incommingMessage []byte
	Logger.UnpackReceive("Received Message from server", finalsend, &incommingMessage)
	fmt.Printf("Received Message: %s\n", incommingMessage)
	//Can be called at any point
	Logger.LogLocalEvent("Example Complete")

	Logger.DisableLogging()
	// Output: Received Message: samplepayload
}

func ExampleGoPriorityLog() {
	Logger := govec.InitGoVectorPriority("MyProcess", "PrioritisedLogFile", govec.NORMAL)

	Logger.LogLocalEventWithPriority("Debug Priority Event", govec.DEBUG)
	Logger.LogLocalEventWithPriority("Normal Priority Event", govec.NORMAL)
	Logger.LogLocalEventWithPriority("Notice Priority Event", govec.NOTICE)
	Logger.LogLocalEventWithPriority("Warning Priority Event", govec.WARNING)
	Logger.LogLocalEventWithPriority("Error Priority Event", govec.ERROR)
	Logger.LogLocalEventWithPriority("Critical Priority Event", govec.CRITICAL)

	Logger.DisableLogging()
	//BUG Output contains timestamps so it cant be tested with *******
	//comments
	//Debug Priority Event
	//Normal Priority Event
	//Notice Priority Event
	//Warning Priority Event
	//Error Priority Event
	//Critical Priority Event
}

package govec_test

import (
	"fmt"

	"github.com/DistributedClocks/GoVector/govec"
)

func ExampleGoLog() {
	Logger := govec.InitGoVector("MyProcess", "LogFile", govec.GetDefaultConfig())

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

	// Output: Received Message: samplepayload
}

func ExampleGoTSLog() {
	config := govec.GetDefaultConfig()
	config.UseTimestamps = true
	Logger := govec.InitGoVector("MyProcess", "LogFile", config)

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

	// Output: Received Message: samplepayload
}

func ExampleGoPriorityLog() {
	config := govec.GetDefaultConfig()
	config.Priority = govec.DEBUG
	config.PrintOnScreen = true
	Logger := govec.InitGoVector("MyProcess", "PrioritisedLogFile", config)

	Logger.LogLocalEventWithPriority("Debug Priority Event", govec.DEBUG)
	Logger.LogLocalEventWithPriority("Info Priority Event", govec.INFO)
	Logger.LogLocalEventWithPriority("Warning Priority Event", govec.WARNING)
	Logger.LogLocalEventWithPriority("Error Priority Event", govec.ERROR)
	Logger.LogLocalEventWithPriority("Fatal Priority Event", govec.FATAL)

	//BUG Output contains timestamps so it cant be tested with *******
	//comments
	//Debug Priority Event
	//Info Priority Event
	//Warning Priority Event
	//Error Priority Event
	//Fatal Priority Event
}

package govec_test

import (
	"fmt"

	"github.com/DistributedClocks/GoVector/govec"
)

//Basic example of GoVectors key functions
func ExampleGoLog_basic() {
	//Initialize logger with default configuration. This can be done in
	//a var block to ensure that GoVector is initialized at boot time.
	Logger := govec.InitGoVector("MyProcess", "LogFile", govec.GetDefaultConfig())

	//An example message
	messagepayload := []byte("samplepayload")

	//Prior to sending a message, call PrepareSend on the payload to
	//encode the payload and append this processes vector clock to the
	//payload
	encodedVCpayload := Logger.PrepareSend("Sending Message", messagepayload)

	//encodedVCpayload is ready to be written to the network
	//ex) conn.Write(encodedVCpayload)

	//Receiving Example
	//First allocate a buffer to receive a message into. This must be
	//the same type as the encoded message. Here incommingMessage and
	//messagepayload are the same type []byte.
	var incommingMessage []byte

	//Prior to unpacking call a message must be received
	//ex) conn.Read(encodedVCPayload)

	//Call UnpackReceive on received messages to update local vector
	//clock values, and decode the original message.
	Logger.UnpackReceive("Received Message from server", encodedVCpayload, &incommingMessage)
	fmt.Printf("Received Message: %s\n", incommingMessage)

	//Important local events can be timestamped with vector clocks
	//using LogLocalEvent, which also increments the local clock.
	Logger.LogLocalEvent("Example Complete")

	// Output: Received Message: samplepayload
}

//Logging with priority trims all events which are lower from the
//specified priority from the log. This functionality is useful for
//isolating behaviour such as recovery protocols, from common
//behaviour like heartbeats.
func ExampleGoLog_priority() {
	//Access GoVectors default configureation, and set priority
	config := govec.GetDefaultConfig()
	config.Priority = govec.DEBUG
	config.PrintOnScreen = true
	//Initialize GoVector
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

//GoVector logs can be used to associate real time events for
//visualization with TSViz
func ExampleGoLog_tSVizCompatable() {
	//Access config and set timestamps (realtime) to true
	config := govec.GetDefaultConfig()
	config.UseTimestamps = true
	//Initalize GoVector
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

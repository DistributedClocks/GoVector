package main

import "github.com/DistributedClocks/GoVector/govec"
import "fmt"

func main() {
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
}

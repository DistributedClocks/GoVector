package main

import "./govec"

func main() {
	Logger := govec.InitializeMutipleExecutions("example", "example")
	Logger2 := govec.Initialize("example2", "example2")

	sendbuf := []byte("messagepayload")
	finalsend := Logger.PrepareSend("Sending Message", sendbuf)
	finalsend = Logger2.PrepareSend("Sending Message2", sendbuf)
	//send message
	Logger2.LogLocalEvent("Message Processed2")
	//s:= string(finalsend[:])
	//fmt.Println(s)

	//receive message
	recbuf := Logger.UnpackReceive("receivingmsg", finalsend)

	Logger.LogLocalEvent("Message Processed")
	//Logger.UnpackReceive(finalsend)
	//s:= string(recbuf[:])
	//fmt.Println(s)
	finalsend = Logger.PrepareSend("Sending Message Again", recbuf)

	// This should not appear in log
	Logger.DisableLogging()
	Logger.LogLocalEvent("-Logging was disabled here")

	//This should appear in log
	Logger.EnableLogging()
	Logger.LogLocalEvent("-Logging is renabled here-")
}

package govec

import (
	"testing"

	"github.com/DistributedClocks/GoVector/govec"
)

func TestSendReceive(t *testing.T) {
	Logger := govec.InitializeMutipleExecutions("example", "example")
	Logger2 := govec.Initialize("example2", "example2")

	sendbuf := "messagepayload"
	finalsend := Logger.PrepareSend("Sending Message", sendbuf)
	finalsend = Logger2.PrepareSend("Sending Message2", sendbuf)
	//send message
	Logger2.LogLocalEvent("Message Processed2")
	//s:= string(finalsend[:])
	//fmt.Println(s)

	var recbuf string
	//receive message
	Logger.UnpackReceive("receivingmsg", finalsend, &recbuf)
	if recbuf != sendbuf {
		t.Error("send and receive inconsistant")
	}

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

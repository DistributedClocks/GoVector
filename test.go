package main

import "./govec"


func main() {
	Logger:= govec.Initialize("example", true, true, true)
	
	sendbuf := []byte("messagepayload")
	finalsend := Logger.PrepareSend("Sending Message",sendbuf)
	//send message
	
	//s:= string(finalsend[:])
	//fmt.Println(s)
		
	//receive message
	recbuf:= Logger.UnpackReceive("receivingmsg", finalsend)
	Logger.MarkAndLogLocalEvent("Message Processed")
	//Logger.UnpackReceive(finalsend)
	//s:= string(recbuf[:])
	//fmt.Println(s)
    finalsend = Logger.PrepareSend("Sending Message Again" , recbuf)
}
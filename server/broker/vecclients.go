package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"os"
)


type Subscriber interface {
	Send(message Message)
	GetName() string 
}

type WSSub struct {
	Name      string
	Conn      *websocket.Conn
}

func (ws *WSSub) GetName() (name string) {
	return ws.Name
}

//Sending message block to the client
func (ws *WSSub) Send(message Message) {
	ws.Conn.Write([]byte(message.GetMessage()))
}

type FileSub struct {
	Name 	  string
	Logname	  string
}

func (fs *FileSub) GetName() (name string) {
	return fs.Name
}

func (fs *FileSub) CreateLog() {
	//Starting File IO. If Log exists, Log will be deleted and a new one will be created
	fs.Logname = fs.Name + "-Log.txt"

	if _, err := os.Stat(fs.Logname); err == nil {
		//it exists... deleting old log
		fmt.Println(fs.Logname, "exists! ... Deleting ")
		os.Remove(fs.Logname)
	}
	//Creating new Log
	file, err := os.Create(fs.Logname)
	if err != nil {
		panic(err)
	}
	file.Close()

	//Log it
	fs.Send("Initialization Complete\r\n")	
}

func (fs *FileSub) Send(message Message) {
	logMessage = message.name + ": " + message.message + " " + message.vclock + "\r\n"
	complete := true
	file, err := os.OpenFile(fs.Logname, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		complete = false
	}
	defer file.Close()

	if _, err = file.WriteString(logMessage); err != nil {
		complete = false
	}

	if (!complete) {
		fmt.Println("Could not write to log file.")
	}
}

type Publisher interface {
	GetName() string
}

type TCPPub struct {
	Name      string
	Conn      *websocket.Conn // net.Conn for tcp connection likely
}

func (tp *TCPPub) GetName() (name string) {
	return tp.Name
}

//Client has a new message to broadcast
//func (tp *TCPSub) Publish(msg string) {
//	tp.belongsTo.AddMsg(msg)
//}


package servervec

import (
	"./../websocket" //gorilla websocket implementation
	"fmt"
	"net"
	"net/http"
	"time"
	"sync"
	"strings"
	"os"
)


type Subscriber interface {
	Send(message string)
	GetName() string
	Filter() 
}

type WSSub struct {
	name      string
	conn      *websocket.Conn
}

func (ws *WSSub) GetName() (name string) {
	ws.name
}

//Sending message block to the client
func (ws *WSSub) Send(message string) {
	ws.conn.WriteMessage(websocket.TextMessage, []byte(msgs))
}

type FileSub struct {
	name 	  string
	logname	  string
}

func (fs *FileSub) GetName() (name string) {
	fs.name
}

func (fs *FileSub) CreateLog() {
	//Starting File IO. If Log exists, Log will be deleted and a new one will be created
	fs.logname := fs.name + "-Log.txt"

	if _, err := os.Stat(fs.logname); err == nil {
		//it exists... deleting old log
		fmt.Println(fs.logname, "exists! ... Deleting ")
		os.Remove(fs.logname)
	}
	//Creating new Log
	file, err := os.Create(fs.logname)
	if err != nil {
		panic(err)
	}
	file.Close()

	//Log it
	fs.Send("Initialization Complete\r\n")	
}

func (fs *FileSub) Send(message string) {
	complete := true
	file, err := os.OpenFile(fs.logname, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		complete = false
	}
	defer file.Close()

	if _, err = file.WriteString(message); err != nil {
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
	name      string
	conn      *websocket.Conn // net.Conn for tcp connection likely
}

func (tp *TCPSub) GetName() (name string) {
	tp.name
}

//Client has a new message to broadcast
//func (tp *TCPSub) Publish(msg string) {
//	tp.belongsTo.AddMsg(msg)
//}


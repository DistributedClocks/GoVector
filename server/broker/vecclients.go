package brokervec

import (
	"golang.org/x/net/websocket"
	"log"
	"net"
	"time"
	"os"
)


type Subscriber interface {
	Send(message Message)
	GetName() string 
	HasNetworkFilter() bool
	SetNetworkFilter(toggle bool) 
	AddFilterKey(key int)
	GetFilters() []int
}

type WSSub struct {
	Name      		string
	Conn      		*websocket.Conn
	TimeRegistered	time.Time
	NetworkFilter 	bool
	FilterKeys		[]int
}

func (ws *WSSub) GetName() (name string) {
	return ws.Name
}

func (ws *WSSub) HasNetworkFilter() bool {
	return ws.NetworkFilter
}

func (ws *WSSub) SetNetworkFilter(toggle bool) {
	ws.NetworkFilter = toggle
}

func (ws *WSSub) AddFilterKey(key int) {
	ws.FilterKeys = append(ws.FilterKeys, key)
}

func (ws *WSSub) GetFilters() []int {
	return ws.FilterKeys
}

//Sending message block to the client
func (ws *WSSub) Send(message Message) {
	ws.Conn.Write([]byte(message.GetMessage()))
}

type FileSub struct {
	Name 	  		string
	Logname	  		string
	NetworkFilter 	bool
	FilterKeys		[]int
}

func (fs *FileSub) GetName() (name string) {
	return fs.Name
}

func (fs *FileSub) HasNetworkFilter() bool {
	return fs.NetworkFilter
}

func (fs *FileSub) SetNetworkFilter(toggle bool) {
	fs.NetworkFilter = toggle
}

func (fs *FileSub) AddFilterKey(key int) {
	fs.FilterKeys = append(fs.FilterKeys, key)
}

func (fs *FileSub) GetFilters() []int {
	return fs.FilterKeys
}

func (fs *FileSub) CreateLog() {
	//Starting File IO. If Log exists, Log will be deleted and a new one will be created
	fs.Logname = fs.Name + "-Log.txt"

	if _, err := os.Stat(fs.Logname); err == nil {
		//it exists... deleting old log
		log.Println("FileSub: " + fs.Logname, "exists! ... Deleting ")
		os.Remove(fs.Logname)
	}
	//Creating new Log
	file, err := os.Create(fs.Logname)
	if err != nil {
		panic(err)
	}
	file.Close()

	//Log it
	logmsg := LogMessage{
		Message: "Initialization Complete\r\n",
		ReceiptTime: time.Now(),
	}
	fs.Send(&logmsg)	
}

func (fs *FileSub) Send(message Message) {
	logMessage := message.GetMessage()
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
		log.Println("FileSub: Could not write to log file.")
	}
}

type Publisher interface {
	GetName() string
}

type TCPPub struct {
	Name      string
	Conn      net.Conn 
}

func (tp *TCPPub) GetName() (name string) {
	return tp.Name
}



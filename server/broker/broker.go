package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/rpc"
	"net/http"
	"time"
	"log"
)

// Vector Messaging Server

type VectorBroker struct {
	pubManager *PubManager	
	subManager *SubManager

}

//initializing the server
func (vb *VectorBroker) Init(logfilename string) {
	fmt.Println("Server init")
	vb.pubManager = NewPubManager()
	
	fmt.Println("Trying to register pubmanager as an rpc server")
	rpc.Register(vb.pubManager)
	rpc.HandleHTTP()
	
	//fmt.Println("Registered as server, starting to listen")
	//http.Handle("/ws", websocket.Handler(vb.pubManager.pubWSHandler))
	
	//log.Fatal(http.ListenAndServe(":8000", nil))
	
	
	vb.subManager = NewSubManager(vb.pubManager.Queue)
	vb.subManager.AddLogSubscriber(logfilename)
	rpc.Register(vb.subManager)

	
	http.Handle("/ws", websocket.Handler(vb.subManager.subWSHandler))
	fmt.Println("Registered handler, trying to listen and serve")
	
	//the "heartbeat" for broadcasting messages
	fmt.Println("Starting heartbeat")
	go func() {		
		for {
			vb.subManager.BroadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()
			
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}



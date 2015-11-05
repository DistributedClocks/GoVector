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
	
	rpc.Register(vb.pubManager)
	http.Handle("/ws", websocket.Handler(vb.pubManager.pubWSHandler))
	http.HandleFunc("/", staticFiles)
	log.Fatal(http.ListenAndServe(":8000", nil))
	
	rpc.Register(vb.subManager)
	vb.subManager = NewSubManager()
	vb.subManager.AddLogSubscriber(logfilename)
	
	http.Handle("/ws", websocket.Handler(vb.subManager.subWSHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
	
	//the "heartbeat" for broadcasting messages
	go func() {
		for {
			vb.subManager.BroadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./../static/"+r.URL.Path)
}


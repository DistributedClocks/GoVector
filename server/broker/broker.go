package brokervec

import (
	//"./websocket" //gorilla websocket implementation
	"golang.org/x/net/websocket"
	"fmt"
	"net"
	"net/rpc/jsonrpc"
	"net/http"
	"encoding/json"
	"time"
	"sync"
//	"strings"
//	"os"
	"./clients"
)

// Vector Messaging Server

type VectorBroker struct {
	pubManager PubManager	
	subManager SubManager

	queue   chan Message

}

//initializing the server
func (vb *VectorBroker) Init(logfilename string) {
	fmt.Println("Server init")
	vb.queue = make(chan Message, 20)
	vb.pubManager = NewPubManager()
	
	vb.subManager = NewSubManager()
	vb.subManager.AddLogSubscriber(logfilename)
	
	//the "heartbeat" for broadcasting messages
	go func() {
		for {
			vb.BroadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

//broadcasting all the messages in the queue in one block
func (vs *VectorBroker) BroadCast() {
	msgBlock := nil
infLoop:
	for {
		select {
		case m := <- vs.queue:
			msgBlock = m
		default:
			break infLoop
		}
	}
	if msgBlock != nil {
		for _, client := range vs.subscribers {
			client.Send(msgBlock)
		}
	}
}


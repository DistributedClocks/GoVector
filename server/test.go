package main

import (
	//"golang.org/x/net/websocket"
	"fmt"
	//"net"
	//"net/rpc/jsonrpc"
	//"net/http"
	//"encoding/json"
	//"time"
	//"sync"
	//"./broker"
	"./../govec"
)


// Test functions...

func main() {
	gp := govec.NewGoPublisher("127.0.0.1", "8000")
	fmt.Println("Registered publisher, sending test message")
	fmt.Println("%v", gp)
	for {
		
	}
	//gp.SendLocalMessage("test", "test", "test")
}

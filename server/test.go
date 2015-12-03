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
	fmt.Println("Registered publisher gp, sending test message")
	
	gp1 := govec.NewGoPublisher("127.0.0.1", "8000")
	fmt.Println("Registered publisher gp1, sending test message")
	
	gp.PublishLocalMessage("test", "test", "test")
	gp.PublishNetworkMessage("net", "net", "net")
	
	gp1.PublishLocalMessage("test", "test", "test")
	gp1.PublishNetworkMessage("net", "net", "net")
	
}

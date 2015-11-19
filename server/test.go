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
	
	gp.SendTestMessage()
	gp1.SendTestMessage()
	
	gp.SendLocalMessage("test", "test", "test")
	gp.SendNetworkMessage("net", "net", "net")
	
	gp1.SendLocalMessage("test", "test", "test")
	gp1.SendNetworkMessage("net", "net", "net")
	
}

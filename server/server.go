package main

import (
	//"golang.org/x/net/websocket"
	"fmt"
	"net"
	//"net/rpc/jsonrpc"
	"net/http"
	//"encoding/json"
	//"time"
	//"sync"
	"./broker"
)

//global variable for handling all server traffic
var broker brokervec.VectorBroker

//Printing out the various ways the server can be reached by the clients
func printClientConnInfo() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Oops: " + err.Error())
		return
	}

	fmt.Println("Chat clients can connect at the following addresses:\n")

	for _, a := range addrs {
		if a.String() != "0.0.0.0" {
			fmt.Println("http://" + a.String() + ":8000/\n")
		}
	}
}

//#############MAIN FUNCTION and INITIALIZATIONS

func main() {
	printClientConnInfo()
	http.HandleFunc("/", staticFiles)
	broker.Init("E:/Documents/UBCCS/448/GoVector/server/test")

}

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/"+r.URL.Path)
}


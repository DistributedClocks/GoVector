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


// MANUAL MESSAGE PROCESSSING
//	delimiter := ";"
//	components := strings.Split(message, delimiter)
//	if len(components) >= 2 {
//		clientType := strings.Trim(components[0], " ")
//		clientName := strings.Trim(components[1], " ")
		
//		if clientType == "Publisher" {
//			client := server.RegisterPublisher(clientName, conn)
//			if client == nil {
//				conn.Close() //closing connection to indicate failed registration
//				return
//			}

//			//then watch for incoming messages
//			for {
//				_, msg, err := conn.ReadMessage()
//				if err != nil { //if error then assuming that the connection is closed
//					client.Exit()
//					return
//				}
//				client.NewMsg(string(msg))
//			}
//		} else if clientType == "Subscriber" {
//			client := server.RegisterSubscriber(clientName, conn)
//			if client == nil {
//				conn.Close() //closing connection to indicate failed registration
//				return
//			}
//		} 
//	} else {
//		fmt.Println("Error processing message: ", message, "\nPlease connect with a message in the format \"Publisher/Subscriber; Name\"")
//		conn.Close() //closing connection to indicate failed registration
//		return
//	}

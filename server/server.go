package main

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

type VectorServer struct {
	publishers map[string]servervec.Publisher	
	subscribers map[string]servervec.Subscriber
	publishersMtx sync.Mutex
	subscribersMtx sync.Mutex

	queue   chan Message
}

//initializing the server
func (vs *VectorServer) Init(logfilename string) {
	fmt.Println("Server init")
	vs.queue = make(chan Message, 5)
	vs.publishers = make(map[string]servervec.Publisher)
	vs.subscribers = make(map[string]servervec.Subscriber)

	// Setup the logfile so the server keeps a copy
	log := servervec.FileSub{
		Name:      logfilename,
		Logname:   "",
	}
	log.CreateLog()
	vs.subscribers[log.Name] = &log

	//the "heartbeat" for broadcasting messages
	go func() {
		for {
			vs.BroadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (vs *VectorServer) RegisterPublisher(name string, conn *websocket.Conn) {
	defer vs.publishersMtx.Unlock()
	vs.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	if _, exists := vs.publishers[name]; exists {
		conn.Close()
	}
	publisher := servervec.TCPPub{
		Name:      name,
		Conn:      conn,
	}
	vs.publishers[name] = &publisher
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
}

//Registers a new publisher for providing information to the server
func (vs *VectorServer) RegisterSubscriber(name string, conn *websocket.Conn) {
	defer vs.subscribersMtx.Unlock()
	vs.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if _, exists := vs.subscribers[name]; exists {
		conn.Close()
	}
	subscriber := servervec.WSSub{
		Name:      name,
		Conn:      conn,
	}
	vs.subscribers[name] = &subscriber
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
}

//disconnecting a publisher from the server
func (vs *VectorServer) UnregisterPublisher(name string) {
	vs.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	delete(vs.publishers, name)
	fmt.Println("<B>" + name + "</B> has left the server.")
	vs.publishersMtx.Unlock() 
}

//disconnecting a subscriber from the server
func (vs *VectorServer) UnregisterSubscriber(name string) {
	vs.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	delete(vs.subscribers, name)
	fmt.Println("<B>" + name + "</B> has left the server.")
	vs.subscribersMtx.Unlock() 
}

type Message struct {
	name string // pid
	vclock string
	message string
}
//adding message to queue
func (vs *VectorServer) AddMsg(msg *Message, reply *string) error {
	vs.queue <- *msg
	*reply = "Added to queue"
	return nil
}

//broadcasting all the messages in the queue in one block
func (vs *VectorServer) BroadCast() {
	msgBlock := ""
infLoop:
	for {
		select {
		case m := <- vs.queue:
			msgBlock += m.name + ": " +m.message + " " + m.vclock + "\r\n"
		default:
			break infLoop
		}
	}
	if len(msgBlock) > 0 {
		for _, client := range vs.subscribers {
			client.Send(msgBlock)
		}
	}
}


//################CLIENT TYPE AND METHODS

//type Publisher struct {
//	name      string
//	conn      *websocket.Conn
//	belongsTo *VectorServer
//}

////Client has a new message to broadcast
//func (cl *Publisher) NewMsg(msg string) {
//	cl.belongsTo.AddMsg(msg)
//}

////Exiting out
//func (cl *Publisher) Exit() {
//	cl.belongsTo.UnregisterPublisher(cl.name)
//}

//type Subscriber struct {
//	name      string
//	conn      *websocket.Conn
//	belongsTo *VectorServer
//}

////Exiting out
//func (cl *Subscriber) Exit() {
//	cl.belongsTo.UnregisterSubscriber(cl.name)
//}

////Sending message block to the client
//func (cl *Subscriber) Send(msgs string) {
//	cl.conn.WriteMessage(websocket.TextMessage, []byte(msgs))
//}

//global variable for handling all server traffic
var server VectorServer

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/"+r.URL.Path)
}

//##############HANDLING THE WEBSOCKET
//var upgrader = websocket.Upgrader{
//	ReadBufferSize:  1024,
//	WriteBufferSize: 1024,
//	CheckOrigin:     func(r *http.Request) bool { return true }, //not checking origin
//}

//this is also the handler for joining the server
func wsHandler(conn *websocket.Conn) {

//	conn, err := upgrader.Upgrade(w, r, nil)

//	if err != nil {
//		fmt.Println("Error upgrading to websocket:", err)
//		return
//	}
	//go func() {
		var msg []byte
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			fmt.Println("Error reading connection:", err)
			conn.Close()
			return
		}
		processMessage(msg, conn)
		jsonrpc.ServeConn(conn)
		
	//}()
}

func processMessage(message []byte, conn *websocket.Conn) {
	
	var client Client
	err := json.Unmarshal(message, &client)
	fmt.Println(string(message))
	if err == nil {
		if client.Type == "Publisher" {
			server.RegisterPublisher(client.Name, conn)
			return
		} else if client.Type == "Subscriber" {
			server.RegisterSubscriber(client.Name, conn)
			return
		}
	}
			
	fmt.Println("Error processing message: ", message, "\nPlease connect with a json object in the format \"Name\": \"name\", \"Type\": \"Publisher/Subscriber\"}")
	conn.Close() //closing connection to indicate failed registration

		
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
}

type Client struct {
	Name string
	Type string
}

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
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.HandleFunc("/", staticFiles)
	server.Init("E:/Documents/UBCCS/448/GoVector/server/test")
	http.ListenAndServe(":8000", nil)
}

package main

import (
	"./websocket" //gorilla websocket implementation
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/http"
	"time"
	"sync"
	"strings"
	"os"
	"./vecclients"
)

//############ CHATROOM TYPE AND METHODS

type VectorServer struct {
	publishers map[string]Publisher	
	subscribers map[string]Subscriber
	publishersMtx sync.Mutex
	subscribersMtx sync.Mutex

	queue   chan string
}

//initializing the chatroom
func (vs *VectorServer) Init(logfilename string) {
	fmt.Println("Server init")
	vs.queue = make(chan string, 5)
	vs.publishers = make(map[string]Publisher)
	vs.subscribers = make(map[string]Subscriber)

	// Setup the logfile so the server keeps a copy
	log := FileSub{
		name:      logfilename,
		logname:   "",
		belongsTo: vs,
	}
	log.CreateLog()
	vs.subscribers[log.name] = log

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
func (vs *VectorServer) RegisterPublisher(name string, conn *websocket.Conn) *Publisher {
	defer vs.publishersMtx.Unlock()
	vs.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	if _, exists := vs.publishers[name]; exists {
		return nil
	}
	publisher := TCPPub{
		name:      name,
		conn:      conn,
		belongsTo: vs,
	}
	vs.publishers[name] = publisher
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
	return &publisher
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (vs *VectorServer) RegisterSubscriber(name string, conn *websocket.Conn) *Subscriber {
	defer vs.subscribersMtx.Unlock()
	vs.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if _, exists := vs.subscribers[name]; exists {
		return nil
	}
	subscriber := WSSub{
		name:      name,
		conn:      conn,
		belongsTo: vs,
	}
	vs.subscribers[name] = subscriber
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
	return &subscriber
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

//adding message to queue
func (vs *VectorServer) AddMsg(msg string) {
	vs.queue <- msg
}

//broadcasting all the messages in the queue in one block
func (vs *VectorServer) BroadCast() {
	msgBlock := ""
infLoop:
	for {
		select {
		case m := <- vs.queue:
			msgBlock += m + "\r\n"
		default:
			break infLoop
		}
	}
	if len(msgBlock) > 0 {
		vs.Log(msgBlock)
		for _, client := range vs.subscribers {
			client.Send(msgBlock)
		}
	}
}


//################CLIENT TYPE AND METHODS

type Publisher struct {
	name      string
	conn      *websocket.Conn
	belongsTo *VectorServer
}

//Client has a new message to broadcast
func (cl *Publisher) NewMsg(msg string) {
	cl.belongsTo.AddMsg(msg)
}

//Exiting out
func (cl *Publisher) Exit() {
	cl.belongsTo.UnregisterPublisher(cl.name)
}

type Subscriber struct {
	name      string
	conn      *websocket.Conn
	belongsTo *VectorServer
}

//Exiting out
func (cl *Subscriber) Exit() {
	cl.belongsTo.UnregisterSubscriber(cl.name)
}

//Sending message block to the client
func (cl *Subscriber) Send(msgs string) {
	cl.conn.WriteMessage(websocket.TextMessage, []byte(msgs))
}

//global variable for handling all server traffic
var server VectorServer

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/"+r.URL.Path)
}

//##############HANDLING THE WEBSOCKET
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //not checking origin
}

//this is also the handler for joining the server
func wsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Error upgrading to websocket:", err)
		return
	}
	go func() {
//		_, msg, err := conn.ReadMessage()
//		if err != nil {
//			fmt.Println("Error upgrading to websocket:", err)
//			conn.Close()
//			return
//		}
//		processMessage(string(msg), conn)
		jsonrpc.ServeConn(conn)
		
	}()
}

func processMessage(message string, conn *websocket.Conn) {
	delimiter := ";"
	components := strings.Split(message, delimiter)
	if len(components) >= 2 {
		clientType := strings.Trim(components[0], " ")
		clientName := strings.Trim(components[1], " ")
		
		if clientType == "Publisher" {
			client := server.RegisterPublisher(clientName, conn)
			if client == nil {
				conn.Close() //closing connection to indicate failed registration
				return
			}

			//then watch for incoming messages
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil { //if error then assuming that the connection is closed
					client.Exit()
					return
				}
				client.NewMsg(string(msg))
			}
		} else if clientType == "Subscriber" {
			client := server.RegisterSubscriber(clientName, conn)
			if client == nil {
				conn.Close() //closing connection to indicate failed registration
				return
			}
		} 
	} else {
		fmt.Println("Error processing message: ", message, "\nPlease connect with a message in the format \"Publisher/Subscriber; Name\"")
		conn.Close() //closing connection to indicate failed registration
		return
	}	
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
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", staticFiles)
	server.Init("E:/Documents/UBCCS/448/GoVector/server/test")
	http.ListenAndServe(":8000", nil)
}

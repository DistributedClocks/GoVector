
package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/rpc/jsonrpc"
	"encoding/json"
	"sync"
	"./nonce"
)

type SubManager struct {
	Subscribers map[string]Subscriber	
	subscribersMtx sync.Mutex

	queue   chan Message
}

func NewSubManager() *SubManager {
    sm := &SubManager{
        Subscribers: make(map[string]Subscriber),
    }

    return sm
}

func (sm *SubManager) AddLogSubscriber(logFileName string) {
	// Setup the logfile so the server keeps a copy
	log := FileSub{
		Name:      logFileName,
		Logname:   "",
	}
	log.CreateLog()
	sm.Subscribers[log.Name] = &log
}

//Registers a new publisher for providing information to the server
func (sm *SubManager) RegisterSubscriber(name string, conn *websocket.Conn) {
	defer sm.subscribersMtx.Unlock()
	sm.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if _, exists := sm.Subscribers[name]; exists {
		conn.Close()
	}
	subscriber := WSSub{
		Name:      name,
		Conn:      conn,
	}
	sm.Subscribers[name] = &subscriber
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
}

//this is also the handler for joining the server
func (sm *SubManager) subWSHandler(conn *websocket.Conn) {

	var msg string
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		fmt.Println("Error reading connection:", err)
		conn.Close()
		return
	}
	sm.processMessage(msg, conn)
	go	jsonrpc.ServeConn(conn)
		
}

type Client struct {
	Name string
	Type string
}

func (sm *SubManager) processMessage(message string, conn *websocket.Conn) {
	
	//var client Client
	//err := json.Unmarshal(message, &client)
	//fmt.Println(string(message))
	
//	if err == nil {
//		if client.Type == "Subscriber" {
//			sm.RegisterSubscriber(client.Name, conn)
//			return
//		} else {
//			fmt.Println("Error processing message: ", message, "\nPlease connect with a message in the format \"Publisher/Subscriber; Name\"")
//			conn.Close() //closing connection to indicate failed registration
//			return
//		}
//	}
	//fmt.Println("Error processing message: ", message, "\nPlease connect with a json object in the format \"Name\": \"name\", \"Type\": \"Publisher/Subscriber\"}")
	//conn.Close() //closing connection to indicate failed registration
	var non *nonce.Nonce
	fmt.Println(message)
	non = nonce.NewNonce(message)		
	b, err := json.Marshal(non)
	if err != nil {
		conn.Write(b)
		fmt.Println("Sending nonce with nonce: ", non.Nonce)
		sm.RegisterSubscriber(non.Nonce, conn)
	} else {
		fmt.Println("Error creating nonce. Subscriber not registered.")
		conn.Close() //closing connection to indicate failed registration
	}
}

//broadcasting all the messages in the queue in one block
func (sm *SubManager) BroadCast() {
	var msgBlock Message
infLoop:
	for {
		select {
		case m := <- sm.queue:
			msgBlock = m
		default:
			break infLoop
		}
	}
	if msgBlock != nil {
		for _, client := range sm.Subscribers {
			client.Send(msgBlock)
		}
	}
}
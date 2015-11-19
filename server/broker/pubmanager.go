package brokervec

import (
	//"golang.org/x/net/websocket"
	"fmt"
	"net/rpc/jsonrpc"
	"net"
	"encoding/json"
	"sync"
	"./nonce"
	"log"
)

type PubManager struct {
	publishers map[string]Publisher	
	publishersMtx sync.Mutex

	Queue   chan Message
	messageStore []Message
}

func NewPubManager() *PubManager {
    pm := &PubManager{
        publishers: make(map[string]Publisher),
        messageStore: make([]Message, 1),
		Queue: make(chan Message),
    }

	go pm.setupPubManagerTCP()
    return pm
}

func (pm *PubManager) setupPubManagerTCP() {
	listener, e := net.Listen("tcp", ":8000")
    if e != nil {
        log.Fatal("listen error:", e)
    }
	fmt.Println("Listening")
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatal(err)
        }
		pm.registerPublisher(conn)
		fmt.Println("Serving connection")
        go jsonrpc.ServeConn(conn)
    }
}

//adding message to queue
func (pm *PubManager) AddLocalMsg(msg *LocalMessage, reply *string) error {
	fmt.Println("In add local message rpc call")
	pm.Queue <- msg
	*reply = "Added to queue"
	return nil
}

//adding message to queue
func (pm *PubManager) AddNetworkMsg(msg *NetworkMessage, reply *string) error {
	fmt.Println("In add net message rpc call", msg.GetNonce())
	pm.Queue <- msg
	*reply = "Added to queue"
	return nil
}

//test
func (pm *PubManager) Test(msg *string, reply *string) error {
	fmt.Println("In test rpc call, ", *msg)
	*reply = "Test successful"
	return nil
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (pm *PubManager) registerPublisher(conn net.Conn) {
	defer pm.publishersMtx.Unlock()

	var non *nonce.Nonce
	non = nonce.NewNonce("") 
	name := non.Nonce
	pm.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	if _, exists := pm.publishers[name]; exists {
		log.Panic("That publisher has already been registered, closing connection, please try again.")
		conn.Close()
		return
	}
	e := json.NewEncoder(conn)
    tcperr := e.Encode(&non)

	if tcperr != nil {
        log.Fatal(tcperr)
    }

	publisher := TCPPub{
		Name:      name,
		Conn:      conn,
	}
	pm.publishers[name] = &publisher
	 
	fmt.Println("<B>" + name + "</B> has joined the publisher list.")
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (pm *PubManager) UnregisterPublisher(msg Message, reply *string) error {
	defer pm.publishersMtx.Unlock()

	pm.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	name := msg.GetNonce()
	if _, exists := pm.publishers[name]; exists {
		delete(pm.publishers, name)
		*reply = "Successfully unregistered."
		log.Println(*reply, name)
	} else
	{
		panic("Could not find that publisher.")
	}	
	return nil
}

//this is also the handler for joining the server
//func (pm *PubManager) pubWSHandler(conn *websocket.Conn) {
//	var msg string
//	err := websocket.Message.Receive(conn, &msg)
//	if err != nil {
//		fmt.Println("Error reading connection:", err)
//		conn.Close()
//		return
//	}
//	pm.processFirstPublish(msg, conn)
//	fmt.Println("Serving the RPC connection")
//	go	func() {
//		jsonrpc.ServeConn(conn)
//		fmt.Println("RPC connection closed")
//	}()
//}

//func (pm *PubManager) processFirstPublish(message string, conn *websocket.Conn) {
	
//	var non *nonce.Nonce
//	non = nonce.NewNonce(message)	
//	err := websocket.JSON.Send(conn, non)
//	if err == nil {
//		fmt.Println("Sending nonce with nonce: ", non.Nonce)
//		pm.registerPublisher(non.Nonce, conn)
//	} else {
//		fmt.Println("Error creating nonce. Publisher not registered.")
//		conn.Close() //closing connection to indicate failed registration
//	}
//}




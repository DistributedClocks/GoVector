package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/rpc/jsonrpc"
	//"encoding/json"
	"sync"
	"./nonce"
)

type PubManager struct {
	publishers map[string]Publisher	
	publishersMtx sync.Mutex

	queue   chan Message
	messageStore []Message
}

func NewPubManager() *PubManager {
    pm := &PubManager{
        publishers: make(map[string]Publisher),
        messageStore: make([]Message, 1),
		queue: make(chan Message, 20),
    }
	
    return pm
}


//adding message to queue
func (pm *PubManager) AddMsg(msg *Message, reply *string) error {
	fmt.Println("In add message rpc call")
	pm.queue <- *msg
	*reply = "Added to queue"
	return nil
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (pm *PubManager) registerPublisher(name string, conn *websocket.Conn) {
	defer pm.publishersMtx.Unlock()
	pm.publishersMtx.Lock() //preventing simultaneous access to the `publishers` map
	if _, exists := pm.publishers[name]; exists {
		conn.Close()
	}
	publisher := TCPPub{
		Name:      name,
		Conn:      conn,
	}
	pm.publishers[name] = &publisher
	 
	fmt.Println("<B>" + name + "</B> has joined the server.")
}

//this is also the handler for joining the server
func (pm *PubManager) pubWSHandler(conn *websocket.Conn) {
	var msg string
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		fmt.Println("Error reading connection:", err)
		conn.Close()
		return
	}
	pm.processFirstPublish(msg, conn)
	fmt.Println("Serving the RPC connection")
	go	jsonrpc.ServeConn(conn)
	fmt.Println("RPC connection closed")
}

func (pm *PubManager) processFirstPublish(message string, conn *websocket.Conn) {
	
	var non *nonce.Nonce
	non = nonce.NewNonce(message)	
	err := websocket.JSON.Send(conn, non)
	if err == nil {
		fmt.Println("Sending nonce with nonce: ", non.Nonce)
		pm.registerPublisher(non.Nonce, conn)
	} else {
		fmt.Println("Error creating nonce. Publisher not registered.")
		conn.Close() //closing connection to indicate failed registration
	}
}




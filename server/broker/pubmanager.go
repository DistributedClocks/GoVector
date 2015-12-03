package brokervec

import (
	"net/rpc/jsonrpc"
	"net"
	"encoding/json"
	"sync"
	"time"
	"./nonce"
	"log"
	"errors"
)

/*
	This class manages the publishers connected to the broker and acts as an
	RPC server for them. It provides publishers with the ability to send 
	messages to the broker that can be read by subscribers and written to a log
	file.
	
	The RPC calls are intended to be used by the wrapper provided in the govec
	library called GoPublisher.
*/

type PubManager struct {
	publishers map[string]Publisher	
	publishersMtx sync.Mutex

	Queue   chan Message
	messageStore []Message
	
	listenPort	string
}

func NewPubManager(listenPort string) *PubManager {
    pm := &PubManager{
        publishers: make(map[string]Publisher),
        messageStore: make([]Message, 1),
		Queue: make(chan Message),
		listenPort: listenPort,
    }

	go pm.setupPubManagerTCP()
    return pm
}

func (pm *PubManager) setupPubManagerTCP() {
	port := ":" + pm.listenPort
	listener, e := net.Listen("tcp", port)
    if e != nil {
        log.Fatal("PubMgr: listen error:", e)
    }
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatal(err)
        }
		pm.registerPublisher(conn)
		log.Println("PubMgr: Serving connection")
        go jsonrpc.ServeConn(conn)
    }
}

//adding message to queue
func (pm *PubManager) AddLocalMsg(msg *LocalMessage, reply *string) error {
	log.Println("PubMgr: Adding local message from nonce: ", msg.GetNonce())
	if _, exists := pm.publishers[msg.GetNonce()]; exists {
		msg.Receipttime = time.Now()
		pm.Queue <- msg
		*reply = "Added to queue"
		return nil
	} else {
		return errors.New("We couldn't find that publisher.")
	}
}

//adding message to queue
func (pm *PubManager) AddNetworkMsg(msg *NetworkMessage, reply *string) error {
	log.Println("PubMgr: Adding net message from nonce: ", msg.GetNonce())
	if _, exists := pm.publishers[msg.GetNonce()]; exists {
		msg.Receipttime = time.Now()
		pm.Queue <- msg
		*reply = "Added to queue"
		return nil
	} else {
		return errors.New("We couldn't find that publisher.")
	}
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
		// Return error instead
		log.Panic("PubMgr: That publisher has already been registered, closing connection, please try again.")
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
	 
	log.Println("PubMgr: " + name + " has joined the publisher list.")
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
		log.Println("PubMgr: ", *reply, name)
	} else
	{
		panic("Could not find that publisher.")
	}	
	return nil
}

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




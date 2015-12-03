
package brokervec

import (
	"golang.org/x/net/websocket"
	"log"
	"fmt"
	"errors"
	"net/rpc/jsonrpc"
	"net/http"
	"time"
	"sync"
	"strconv"
	"regexp"
	"./nonce"
)

/*
	This class manages the subscribers connected to the broker and acts as an
	RPC server for them. It provides subscribers with the ability to receive 
	messages from the broker that were sent by publishers. It also provides the
	ability to filter messages.
	
	The RPC calls are intended to be used by web browsers connecting via 
	WebSocket, but may be used by other modes.
*/

type SubManager struct {
	Subscribers map[string]Subscriber	
	subscribersMtx sync.Mutex

	Filters map[int]string	
	
	vb *VectorBroker
	
	listenPort string
}

func NewSubManager(vb *VectorBroker, logfilename string, subport string) *SubManager {
    sm := &SubManager{
        Subscribers: make(map[string]Subscriber),
		Filters: make(map[int]string),
		vb: vb,
		listenPort: subport,
    }

	sm.setupSubManager(logfilename)
		
    return sm
}

func (sm *SubManager) setupSubManager(logfilename string) {
	log.Println("SubMgr: Registering handler")
	http.Handle("/ws", websocket.Handler(sm.subWSHandler))
	
	//the "heartbeat" for broadcasting messages
	log.Println("SubMgr: Starting heartbeat")
	go func() {		
		for {
			sm.broadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {	
		port := ":" + sm.listenPort	
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
	
	if len(logfilename) > 0 {
		sm.addLogSubscriber(logfilename)
	}
}

func (sm *SubManager) addLogSubscriber(logFileName string) {
	log.Println("SubMgr: Adding log subscriber.")
	// Setup the logfile so the server keeps a copy
	log := FileSub{
		Name:      logFileName,
		Logname:   "",
		NetworkFilter: false,
	}
	log.CreateLog()
	sm.Subscribers[log.Name] = &log
}

//Registers a new publisher for providing information to the server
func (sm *SubManager) registerSubscriber(name string, conn *websocket.Conn) {
	defer sm.subscribersMtx.Unlock()
	sm.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if _, exists := sm.Subscribers[name]; exists {
		log.Panic("SubMgr: That subscriber has already been registered, closing connection, please try again.")
		conn.Close()
		return
	}
	subscriber := WSSub{
		Name:      name,
		Conn:      conn,
		TimeRegistered: time.Now(),
		NetworkFilter: false,
	}
	sm.Subscribers[name] = &subscriber
	 
	log.Println("SubMgr: " + name + " has been registered.")
}

//this is also the handler for joining the server
func (sm *SubManager) subWSHandler(conn *websocket.Conn) {

	var msg string
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		log.Println("SubMgr: Error reading connection:", err)
		conn.Close()
		return
	}
	err = sm.processMessage(msg, conn)
	if err != nil {
		conn.Close()
		return
	}
	// Isn't strictly necessary, but try using a go statement here.
	jsonrpc.ServeConn(conn)
	
	// TODO: Remove subscriber from list here because the connection has closed.
	conn.Close()	
}

func (sm *SubManager) processMessage(message string, conn *websocket.Conn) error {
	
	var non *nonce.Nonce
	log.Println("SubMgr: Name is: ", message)
	non = nonce.NewNonce(message)		
	err := websocket.JSON.Send(conn, non.Nonce)
	if err == nil {
		log.Println("SubMgr: Sending nonce with nonce: ", non.Nonce)
		sm.registerSubscriber(non.Nonce, conn)
	} else {
		conn.Close() //closing connection to indicate failed registration
		log.Println("SubMgr: Error creating nonce. Subscriber not registered.", err)
		return err		
	}
	return nil
}

func (sm *SubManager) passesFilters(subscriber Subscriber, message Message) bool {
	messageType := fmt.Sprintf("%T", message)
	
	if subscriber.HasNetworkFilter() {
		// Message type is local then return false
		if messageType == "*brokervec.LocalMessage" {
			return false
		}
	}
	
	for filter := range subscriber.GetFilters() {
		value, ok := sm.Filters[filter]
		if ok {
			matched, err := regexp.MatchString(value, message.GetMessage())
			if err == nil && !matched {
				return false
			}
		}		
	}
	
	return true
}

//broadcasting all the messages in the queue in one block
func (sm *SubManager) broadCast() {
	var msgBlock Message
infLoop:
	for {
		select {
		case m := <- sm.vb.GetReadQueue():
			msgBlock = m
		default:
			break infLoop
		}
	}
	if msgBlock != nil {
		for _, client := range sm.Subscribers {
			if sm.passesFilters(client, msgBlock) {
				client.Send(msgBlock)
			}			
		}
	}
}

// **************
// RPC Calls
// **************

func (sm *SubManager) AddFilter(msg FilterMessage, reply *string) error {
	log.Println("SubMgr: In AddFilter, Message: ", msg.GetFilter(), "Nonce: ", msg.GetNonce())
	
	if sub, exists := sm.Subscribers[msg.GetNonce()]; exists {
		max := 0
		for key := range sm.Filters {
			if key > max {
				max = key
			}
		}
		sm.Filters[max+1] =	msg.GetFilter()
		sub.AddFilterKey(max+1)
		*reply = "Added filter: " + msg.GetFilter()
	} else {
		return errors.New("We couldn't find that subscriber.")
	}

	return nil
}

func (sm *SubManager) AddNetworkFilter(nonce string, reply *string) error {
	log.Println("SubMgr: In AddNetworkFilter, Nonce: ", nonce)
	
	if sub, exists := sm.Subscribers[nonce]; exists {
		sub.SetNetworkFilter(true)
		*reply = "You will no longer receive local messages."
	} else {
		return errors.New("We couldn't find that subscriber.")
	}

	return nil
}

func (sm *SubManager) SendOldMessages(nonce string, reply *string) error {
	log.Println("SubMgr: In SendOldMessages, Nonce: ", nonce)
	
	if sub, exists := sm.Subscribers[nonce]; exists {
		currentTime := time.Now()
		broker := sm.vb
		messages, num := broker.GetMessagesBefore(currentTime)
		*reply = strconv.Itoa(num)
		go func() {	
			for _, msg := range messages {
				sub.Send(msg)
			}
		}()
	} else {
		return errors.New("We couldn't find that subscriber.")
	}

	return nil
}
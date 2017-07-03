package brokervec

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc/jsonrpc"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/DistributedClocks/GoVector/broker/nonce"
	"golang.org/x/net/websocket"
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
	Subscribers    map[string]Subscriber
	subscribersMtx sync.Mutex // Mutex lock for preventing simultaneous access
	// to subscribers map
	Filters map[int]string

	vb *VectorBroker

	listenPort string
}

// Construct a new SubManager, initialize the network handling and broadcast
// routines
func NewSubManager(vb *VectorBroker, logfilename string, subport string) *SubManager {
	sm := &SubManager{
		Subscribers: make(map[string]Subscriber),
		Filters:     make(map[int]string),
		vb:          vb,
		listenPort:  subport,
	}

	sm.initSubManager(logfilename)

	return sm
}

// Initialize the websocket handling to listen for new connections, start the
// heartbeat for broadcasting messages and add a log subscriber if provided.
func (sm *SubManager) initSubManager(logfilename string) {
	log.Println("SubMgr: Registering handler")
	http.Handle("/ws", websocket.Handler(sm.subWSHandler))

	// Listen on the subscriber port for new websocket connections handled on
	// the subWSHandler.
	go func() {
		port := ":" + sm.listenPort
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	//the "heartbeat" for broadcasting messages
	log.Println("SubMgr: Starting heartbeat")
	go func() {
		for {
			sm.broadCast()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Add a log subscriber if a name is provided.
	if len(logfilename) > 0 {
		sm.addLogSubscriber(logfilename)
	}
}

// Create a log subscriber to record messages to a file and register it in the
// subscriber list.
func (sm *SubManager) addLogSubscriber(logFileName string) {
	log.Println("SubMgr: Adding log subscriber.")
	log := FileSub{
		Name:          logFileName,
		NetworkFilter: false,
	}
	log.CreateLog()
	sm.Subscribers[log.Name] = &log
}

// Registers a new subscriber for receiving messages from the server
func (sm *SubManager) registerSubscriber(name string, conn *websocket.Conn) {
	defer sm.subscribersMtx.Unlock()
	sm.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if _, exists := sm.Subscribers[name]; exists {
		log.Panic("SubMgr: That subscriber has already been registered, closing connection, please try again.")
		conn.Close()
		return
	}
	subscriber := WSSub{
		Name:           name,
		Conn:           conn,
		TimeRegistered: time.Now(),
		NetworkFilter:  false,
	}
	sm.Subscribers[name] = &subscriber

	log.Println("SubMgr: " + name + " has been registered.")
}

// Registers a new subscriber for receiving messages from the server
func (sm *SubManager) unregisterSubscriber(name string) {
	defer sm.subscribersMtx.Unlock()
	sm.subscribersMtx.Lock() //preventing simultaneous access to the `subscribers` map
	if sub, exists := sm.Subscribers[name]; exists {
		log.Println("SubMgr: Unregistering subscriber with nonce: " + name)
		sub.Close()
		delete(sm.Subscribers, name)
		return
	} else {
		log.Panic("Could not find that publisher.")
	}
}

// Handler for joining the server as a subscriber
func (sm *SubManager) subWSHandler(conn *websocket.Conn) {
	// Defer closing the connection because we block when serving the rpc
	// server. This will close the connection if something disrupts the
	// RPC server.
	defer conn.Close()

	// Receive the first message on the websocket connection containing the
	// name of the subscriber
	var msg string
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		log.Println("SubMgr: Error reading connection:", err)
		conn.Close()
		return
	}

	// Turn name into a nonce
	name := nonce.NewNonce(msg)
	log.Println("SubMgr: Name is: ", msg)

	// Use the string version of the nonce to identify the subscriber and
	// send the identifying nonce back to subscriber.
	err = websocket.JSON.Send(conn, name.Nonce)

	// If successful, register the subscriber.
	if err == nil {
		log.Println("SubMgr: Sending name as nonce: ", name.Nonce)
		sm.registerSubscriber(name.Nonce, conn)
	} else {
		conn.Close() //closing connection to indicate failed registration
		log.Println("SubMgr: Error creating nonce. Subscriber not registered.", err)
		return
	}

	// Serve the RPC server on the connection
	jsonrpc.ServeConn(conn)

	// Unregister the subscriber because the RPC connection was lost/closed.
	sm.unregisterSubscriber(name.Nonce)
}

// TODO: UPDATE WHEN FILTERS CHANGED TO USE FIELDS
// Check that a message to be sent passes the filters for the subscriber.
func (sm *SubManager) passesFilters(subscriber Subscriber, message Message) bool {
	messageType := fmt.Sprintf("%T", message)

	// Check for the Network Message filter
	if subscriber.HasNetworkFilter() {
		// If message type is local then return false
		if messageType == "*brokervec.LocalMessage" {
			return false
		}
	}

	// Check the list of filters and return false if the message doesn't
	// pass any of the filters
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

// Broadcasts all the messages in the queue to each subscriber
func (sm *SubManager) broadCast() {
infLoop:
	for {
		select {
		case msg := <-sm.vb.GetReadQueue():
			for _, client := range sm.Subscribers {
				if sm.passesFilters(client, msg) {
					client.Send(msg)
				}
			}
		default:
			break infLoop
		}
	}
}

// **************
// RPC Calls
// **************

// NEEDS TO BE SPLIT INTO CALLS FOR ADDING PID FILTERS, MESSAGE FILTERS AND
// VCLOCK FILTERS
// Adds a regex filter for messages
//func (sm *SubManager) AddFilter(msg FilterMessage, reply *string) error {
//    log.Println("SubMgr: In AddFilter, Regex: ", msg.GetFilter(), "Nonce: ", msg.GetNonce())

//    if sub, exists := sm.Subscribers[msg.GetNonce()]; exists {
//        max := 0
//        for key := range sm.Filters {
//            if key > max {
//                max = key
//            }
//        }
//        sm.Filters[max+1] =    msg.GetFilter()
//        sub.AddFilterKey(max+1)
//        *reply = "Added filter: " + msg.GetFilter()
//    } else {
//        return errors.New("We couldn't find that subscriber.")
//    }

//    return nil
//}

// RPC call to add a network filter. Prevents a subscriber from receiving
// local messages.
func (sm *SubManager) AddNetworkFilter(nonce string, reply *string) error {
	log.Println("SubMgr: In AddNetworkFilter, Nonce: ", nonce)

	if sub, exists := sm.Subscribers[nonce]; exists {
		sub.EnableNetworkFilter()
		*reply = "You will no longer receive local messages."
	} else {
		return errors.New("We couldn't find that subscriber.")
	}

	return nil
}

// RPC call to remove a network filter. Allows a subscriber to receive
// local and network messages.
func (sm *SubManager) RemoveNetworkFilter(nonce string, reply *string) error {
	log.Println("SubMgr: In AddNetworkFilter, Nonce: ", nonce)

	if sub, exists := sm.Subscribers[nonce]; exists {
		sub.DisableNetworkFilter()
		*reply = "You will now receive local and network messages."
	} else {
		return errors.New("We couldn't find that subscriber.")
	}

	return nil
}

// RPC call to send all messages that were received before the subscriber
// joined to that subscriber.
func (sm *SubManager) SendOldMessages(nonce string, reply *string) error {
	log.Println("SubMgr: In SendOldMessages, Nonce: ", nonce)

	if sub, exists := sm.Subscribers[nonce]; exists {
		currentTime := time.Now()
		broker := sm.vb
		messages, num := broker.GetMessagesBefore(currentTime)
		// Plausible timing issue here becauase reply isn't sent back until
		// this function returns and the old messages are sent in a go routine
		// that may or may not happen before it returns.
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

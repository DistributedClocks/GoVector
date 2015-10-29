
package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/rpc/jsonrpc"
	"net/http"
	"encoding/json"
	"sync"
)

type SubManager struct {
	Subscribers map[string]Subscriber	
	subscirbersMtx sync.Mutex

	queue   chan Message
}

func NewSubManager() SubManager {
    sm := SubManager{
        Subscribers: make(map[string]Subscriber),
    }

	http.Handle("/ws", websocket.Handler(subWSHandler))
	http.ListenAndServe(":8080", nil)
    return sm
}

func (sm *SubManager) AddLogSubscriber(logFileName string) {
	// Setup the logfile so the server keeps a copy
	log := FileSub{
		Name:      logfilename,
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
func subWSHandler(conn *websocket.Conn) {

	var msg []byte
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		fmt.Println("Error reading connection:", err)
		conn.Close()
		return
	}
	processMessage(msg, conn)
	go	jsonrpc.ServeConn(conn)
		
}

type Client struct {
	Name string
	Type string
}

func processMessage(message []byte, conn *websocket.Conn) {
	
	var client Client
	err := json.Unmarshal(message, &client)
	fmt.Println(string(message))
	if err == nil {
		if client.Type == "Subscriber" {
			RegisterSubcriber(client.Name, conn)
			return
		} else {
			fmt.Println("Error processing message: ", message, "\nPlease connect with a message in the format \"Publisher/Subscriber; Name\"")
			conn.Close() //closing connection to indicate failed registration
			return
		}
	}
			
	fmt.Println("Error processing message: ", message, "\nPlease connect with a json object in the format \"Name\": \"name\", \"Type\": \"Publisher/Subscriber\"}")
	conn.Close() //closing connection to indicate failed registration
}
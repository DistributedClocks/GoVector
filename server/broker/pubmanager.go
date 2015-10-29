package brokervec

import (
	"golang.org/x/net/websocket"
	"fmt"
	"net/rpc/jsonrpc"
	"net/http"
	"encoding/json"
	"sync"
)

type PubManager struct {
	publishers map[string]Publisher	
	publishersMtx sync.Mutex

	queue   chan Message
	messageStore []Message
}

func NewPubManager() PubManager {
    pm := PubManager{
        publishers: make(map[string]Publisher),
        messageStore: make([]Message, 1),
		queue: make(chan Message, 20),
    }
	
	http.Handle("/ws", websocket.Handler(pubWSHandler))
	http.HandleFunc("/", staticFiles)
	http.ListenAndServe(":8000", nil)
    return pm
}

//Registers a new publisher for providing information to the server
//returns pointer to a Client, or Nil, if the name is already taken
func (pm *PubManager) RegisterPublisher(name string, conn *websocket.Conn) {
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

//adding message to queue
func (pm *PubManager) AddMsg(msg *Message, reply *string) error {
	pm.queue <- *msg
	*reply = "Added to queue"
	return nil
}

//this is also the handler for joining the server
func pubWSHandler(conn *websocket.Conn) {

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
		processFirstPublish(msg, conn)
	go	jsonrpc.ServeConn(conn)
		
	//}()
}

func processFirstPublish(message []byte, conn *websocket.Conn) {
	
	var client Client
	err := json.Unmarshal(message, &client)
	fmt.Println(string(message))
	if err == nil {
		if client.Type == "Publisher" {
			RegisterPublisher(client.Name, conn)
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

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./../static/"+r.URL.Path)
}


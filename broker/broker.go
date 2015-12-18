package brokervec

import (
    "net/rpc"
    "bufio"
    "os"
    "time"
    "log"
    "fmt"
)

// Vector Messaging Server
/*
    - All licences like other licenses ...

    How to Use This Library

    Step 1:
    Create a Global Variable of type brokervec.VectorBroker and Initialize 
    it like this =

    broker.Init(logpath, pubport, subport)
    
    Where:
    - the logpath is the path and name of the log file you want created, or 
    "" if no log file is wanted. E.g. "C:/temp/test" will result in the file 
    "C:/temp/test-log.txt" being created.
    - the pubport is the port you want to be open for publishers to send
    messages to the broker.
    - the subport is the port you want to be open for subscribers to receive 
    messages from the broker.

    Step 2:
    Setup your GoVec so that the realtime boolean is set to true and the correct
    brokeraddr and brokerpubport values are set in the Initialize method you
    intend to use.

    Step 3 (optional):
    Setup a Subscriber to connect to the broker via a WebSocket over the correct
    subport. For example, setup a web browser running JavaScript to connect and
    display messages as they are received.

    A simple standalone program can be found in runbroker.go which will setup 
    a broker with the desired parameters. Please read the documentation in 
    runbroker.go for more details.
    
    Tests can be run via GoVector/test/broker_test.go

*/
type VectorBroker struct {
    pubManager *PubManager    
    subManager *SubManager

    queue   chan Message
    messageStore []Message

}

//initializing the server
func (vb *VectorBroker) Init(logfilename string, pubport string, subport string) {
    log.Println("Broker: Broker init")
	
	// Initialize message storing
    vb.messageStore = make([]Message, 0)
    vb.queue = make(chan Message)
    
	// Initialize the PubManager on the pubport
    vb.pubManager = NewPubManager(vb, pubport)    
    log.Println("Broker: Registering pubmanager as an rpc server")
    rpc.Register(vb.pubManager)
    
	// Initialize the SubManager on the subport
    vb.subManager = NewSubManager(vb, logfilename, subport)
    log.Println("Broker: Registering submanager as an rpc server")
    rpc.Register(vb.subManager)
    
	// Connect the default rpc server to look for http connections
    rpc.HandleHTTP()    

	// Allow a user to close the program by pressing enter.
    fmt.Println("Press enter to shut down broker.")
    reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')
    log.Println("Closing.")
}

// Adds a message to the broadcast queue and to the message archive for this
// session.
func (vb *VectorBroker) AddMessage(message Message) {
    vb.queue <- message 
    vb.messageStore = append(vb.messageStore, message)
}

// Returns a queue that messages can be read from but not written to.
func (vb *VectorBroker) GetReadQueue() <-chan Message {
    return vb.queue
}

// Return all messages that were received before registerTime and the number
// of messages returned.
func (vb *VectorBroker) GetMessagesBefore(registerTime time.Time) ([]Message, int) {
    var oldMessages []Message
    numMessages := 0

    for _, msg := range vb.messageStore {
        if msg.GetTime().Before(registerTime) {
            oldMessages = append(oldMessages, msg)
            numMessages++
        }
    }
    return oldMessages, numMessages
}





package brokervec

import (
	"net/rpc"
	"bufio"
	"os"
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
	display all messages on receipt.

	A simple standalone program can be found in runbroker.go which will setup 
	a broker with the desired parameters. Please read the documentation in 
	runbroker.go for more details.
	
	Tests can be run via GoVector/test/broker_test.go

*/
type VectorBroker struct {
	pubManager *PubManager	
	subManager *SubManager

}

//initializing the server
func (vb *VectorBroker) Init(logfilename string, pubport string, subport string) {
	log.Println("Broker: Broker init")

	vb.pubManager = NewPubManager(pubport)	
	log.Println("Broker: Registering pubmanager as an rpc server")
	rpc.Register(vb.pubManager)
	
	vb.subManager = NewSubManager(vb.pubManager.Queue, logfilename, subport)
	log.Println("Broker: Registering submanager as an rpc server")
	rpc.Register(vb.subManager)
	
	rpc.HandleHTTP()	

	fmt.Println("Press enter to shut down broker.")
	reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')
	log.Println("Closing.")
}



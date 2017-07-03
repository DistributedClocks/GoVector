GoVector
========

This library can be added to a Go project to
generate a [ShiViz](http://bestchai.bitbucket.org/shiviz/)-compatible
vector-clock timestamped log of events in a concurrent or distributed system.
GoVec is compatible with Go 1.4+ 

* govec/    : Contains the Library and all its dependencies
* example/  : Contains a client server example instrumented with GoVec
* test/     : A small set of tests for the library
* broker/   : Automatic live integration with ShiViz (Under
  Development)

### Usage

To use GoVec you must have a correctly configured go development
environment, see [How to write Go
Code](https://golang.org/doc/code.html)

Once you set up your environment, GoVec can be installed with the go
tool command:

> go get github.com/DistributedClocks/GoVector

*gofmt* will automatically add imports for GoVec. If you do not have a
working version of *gofmt* GoVec can be imported by adding:

```go
    "import github.com/DistributedClocks/GoVector/govec"
```

### Index

type GoLog
```go
func Initialize(ProcessName, LogName string) *GoLog
```
```go
func InitializeMultipleExecutions(ProcessName, LogName string) *GoLog
```
```go
func PrepareSend(LogMessage string, buf interface{}) []byte
```
```go
func UnpackReceive(LogMessage string, buf []byte, unpack interface{})
```
```go
func SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error)
```
```go
func LogLocalEvent(string LogMessage)
```
```go
func LogThis(Message string, ProcessID string, VCString string) bool
```
```go
func DisableLogging()
```

####   type GoLog

```go
	type GoLog struct{
		// contains filtered or unexported fields
	}
```
 The GoLog struct provides an interface to creating and maintaining vector timestamp entries in the generated log file
 
#####   func Initialize
```go
	func Initialize(ProcessName, LogName string) *GoLog
```
Returns a Go Log Struct taking in two arguments and truncates previous logs:
* MyProcessName (string): local process name; must be unique in your distributed system.
* LogFileName (string) : name of the log file that will store info. Any old log with the same name will be truncated


#####   func InitializeMultipleExecutions
```go
	func InitializeMultipleExecutions(ProcessName, LogName string) *GoLog
```
Returns a Go Log Struct taking in two arguments without truncating previous log entry:
* MyProcessName (string): local process name; must be unique in your distributed system.
* LogFileName (string) : name of the log file that will store info. Each run will append to log file separated 
by "=== Execution # ==="

#####   func PrepareSend
```go
	func PrepareSend(LogMessage string, buf interface{}) byte[]
```
This function is meant to be used immediately before sending.
LogMessage will be logged along with the time of the send
buf is encode-able data (structure or basic type)
Returned is an encoded byte array with logging information

This function is meant to be called before sending a packet. Usually,
it should Update the Vector Clock for its own process, package with the clock
using gob support and return the new byte array that should be sent onwards
using the Send Command

#####   func UnpackReceive
```go
	func UnpackReceive(LogMessage string, buf byte[], unpack interface{})
```

UnpackReceive is used to unmarshall network data into local structures.
LogMessage will be logged along with the vector time the receive happened
buf is the network data, previously packaged by PrepareSend unpack is
a pointer to a structure, the same as was packed by PrepareSend


This function is meant to be called immediately after receiving
a packet. It unpacks the data by the program, the vector clock. It
updates vector clock and logs it. and returns the user data

##### func SetEncoderDecoder
```go
func SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error)
```
SetEncoderDecoder allows users to specify the encoder, and decoder
used by GoVec in the case the default is unable to perform a
required task
encoder a function which takes an interface, and returns a byte
array, and an error if encoding fails
decoder a function which takes an encoded byte array and a pointer
to a structure. The byte array should be decoded into the structure.

For more information on encoders and decoders see [gob
encoder](https://golang.org/pkg/encoding/gob/) and
[goMsgPack](https://github.com/hashicorp/go-msgpack)

#####   func LogLocalEvent
```go
	func LogLocalEvent(LogMessage string)
```
Increments current vector timestamp and logs it into Log File. 

##### func LogThis

```go
func LogThis(Message string, ProcessID string, VCString string) bool
```
Logs a message along with a processID and a vector clock, the VCString
must be a valid vector clock, true is returned on success

#####   func DisableLogging
```go
	func DisableLogging()
```

Disables Logging. Log messages will not appear in Log file any longer.
Note: For the moment, the vector clocks are going to continue being updated.

###   Examples

The following is a basic example of how this library can be used 
```go
	package main

	import "./govec"

	func main() {
		Logger := govec.Initialize("MyProcess", "LogFile")
		
		//In Sending Process
		
		//Prepare a Message
		messagepayload := []byte("samplepayload")
		finalsend := Logger.PrepareSend("Sending Message", messagepayload)
		
		//send message
		connection.Write(finalsend)

		//In Receiving Process
		
		//receive message
		recbuf := Logger.UnpackReceive("Receiving Message", finalsend)

		//Can be called at any point 
		Logger.LogLocalEvent("Example Complete")
		
		Logger.DisableLogging()
		//No further events will be written to log file
	}
```

This produces the log "LogFile.txt" :

	MyProcess {"MyProcess":1}
	Initialization Complete
	MyProcess {"MyProcess":2}
	Sending Message
	MyProcess {"MyProcess":3}
	Receiving Message
	MyProcess {"MyProcess":4}
	Example Complete

An executable example of a similar program can be found in
[Examples/ClientServer.go](https://github.com/DistributedClocks/GoVector/blob/master/example/ClientServer.go)

### VectorBroker

type VectorBroker
   * func Init(logfilename string, pubport string, subport string)

### Usage

    A simple stand-alone program can be found in server/broker/runbroker.go 
    which will setup a broker with command line parameters.
   	Usage is: 
    "go run ./runbroker (-logpath logpath) -pubport pubport -subport subport"

    Tests can be run via GoVector/test/broker_test.go and "go test" with the 
    Go-Check package (https://labix.org/gocheck). To get this package use 
    "go get gopkg.in/check.v1".
    
Detailed Setup:

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

    Setup your GoVec so that the real-time boolean is set to true and the correct
    brokeraddr and brokerpubport values are set in the Initialize method you
    intend to use.

Step 3 (optional):

    Setup a Subscriber to connect to the broker via a WebSocket over the correct
    subport. For example, setup a web browser running JavaScript to connect and
    display messages as they are received. Make RPC calls by sending a JSON 
    object of the form:
            var msg = {
            method: "SubManager.AddFilter", 
            params: [{"Nonce":nonce, "Regex":regex}], 
            id: 0
            }
            var text = JSON.stringify(msg)

####   RPC Calls

    Publisher RPC calls are made automatically from the GoVec library if the 
    broker is enabled.
    
    Subscriber RPC calls:
    * AddNetworkFilter(nonce string, reply *string)
        Filters messages so that only network messages are sent to the 
        subscriber.      
    * RemoveNetworkFilter(nonce string, reply *string)
        Filters messages so that both network and local messages are sent to the 
        subscriber.
    * SendOldMessages(nonce string, reply *string)
        Sends any messages received before the requesting subscriber subscribed.
 

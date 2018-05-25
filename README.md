GoVector
========

This library can be added to a Go project to
generate a [ShiViz](http://bestchai.bitbucket.io/shiviz/)-compatible
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
func InitGoVector(ProcessName, LogName string) *GoLog
```
```go
func InitGoVectorMultipleExecutions(ProcessName, LogName string) *GoLog
```
```go
func PrepareSend(LogMessage string, buf interface{}) []byte
```
```go
func UnpackReceive(LogMessage string, buf []byte, unpack interface{})
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
 
#####   func InitGoVector
```go
	func InitGoVector(ProcessName, LogName string) *GoLog
```
Returns a Go Log Struct taking in two arguments and truncates previous logs:
* MyProcessName (string): local process name; must be unique in your distributed system.
* LogFileName (string) : name of the log file that will store info. Any old log with the same name will be truncated


#####   func InitGoVectorMultipleExecutions
```go
	func InitGoVectorMultipleExecutions(ProcessName, LogName string) *GoLog
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

#### func Flush
```go
	func Flush() bool
```

Writes the log messages stored in the buffer to the Log File. This function should be used by the application to also force writes in the case of interrupts and crashes.   
Note: Calling Flush when BufferedWrites is disabled is essentially a no-op.

#### func EnableBufferedWrites
```go
	func EnableBufferedWrites()
```

Enables buffered writes to the log file. All the log messages are only written
to the LogFile via an explicit call to the function Flush.
Note: Buffered writes are automatically disabled.

#### func DisableBufferedWrites
```go
	func DisableBufferedWrites()
```

Disables buffered writes to the log file. All the log messages from now on
will be written to the Log file immediately. Writes all the existing
log messages that haven't been written to Log file yet.

### RPC Capture

GoVector provides support for automatically logging RPC Calls from a RPC Client to a RPC Server

```go
	type RPCClientCodec
```

An extension of the default rpc codec which uses a logger of type GoLog to capture all the calls to a RPC Server as well as responses from a RPC server.

```go
	type RPCServerCodec
```

An extension of the default rpc codec which uses a logger of type of GoLog to capture all the requests made from the client to a RPC server as well as the server's to the clients.

```go
	func RPCDial(network, address string, logger *GoLog) (*rpc.Client, error) 
```

RPCDial connects to a RPC server at the specified network address. The logger is provided to be used by the RPCClientCodec for message capture.

```go
	func ServeRPCConn(server *rpc.Server, l net.Listener, logger *GoLog)
```

Convenience function that accepts connections for a given listener and starts a new goroutine for the server to serve a new connection. The logger is provided to be used by the RPCServerCodec for message capture.

### type GoPriorityLog
```go
    type GoPriorityLog struct {
        // Contians filtered or unexported fields
    }    
```

The GoPriorityLog struct provides an interface to creating and maintaining vector timestamp entries in the generated log file as well as priority to local events which are printed out to console. Local events with a priority equal to or higher are logged in the log file.

#### type LogPriority
```
    const (
        DEBUG LogPriority = iota
        NORMAL
        NOTICE
        WARNING
        ERROR
        CRITICAL
    )
```

LogPriority enum provides all the valid Priority Levels that can be used to log events with.

#### func InitGoVectorPriority
```go
	func InitGoVectorPriority(ProcessName, LogName string, Priority LogPriority) *GoLog
```

Returns a Go Log Struct taking in two arguments and truncates previous logs:
* MyProcessName (string): local process name; must be unique in your distributed system.
* LogFileName (string) : name of the log file that will store info. Any old log with the same name will be truncated
* Priority (LogPriority) : priority which decides what future local events should be logged in the log file. Any local event with a priority level equal to or higher than this will be logged in the log file. This priority can be changed using SetPriority.

#### func LogLocalEventWithPriority
```go
	func LogLocalEventWithPriority(LogMessage string, Priority LogPriority)
```

If the priority of the logger is lower than or equal to the priority of this event then the current vector timestamp is incremented and the message is logged it into the Log File. A color coded string is also printed on the console.
* LogMessage (string) : Message to be logged
* Priority (LogPriority) : Priority at which the message is to be logged

#### func SetPriority
```go
	func SetPriority(Priority LogPriority)
```

Sets the priroity which is used to decide which future local events should be logged in the log file. Any future local event with a priority level equal to or higher than this will be logged in the log file.

###   Examples

The following is a basic example of how this library can be used 
```go
	package main

	import "./govec"

	func main() {
		Logger := govec.InitGoVector("MyProcess", "LogFile")
		
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

An executable example of a RPC Client-Server program can be found in 
[Examples/RpcClientServer.go](https://github.com/DistributedClocks/GoVector/blob/master/example/RpcClientServer.go)

An executable example of Priority Logger can be found in
[Examples/PriorityLoggerExample.go](example/PriorityLoggerExample.go)

Here is a sample output of the priority logger

![Examples/Output/PriorityLoggerOutput.png](example/output/PriorityLoggerOutput.PNG)
<!-- July 2017: Brokers are no longer supported, maybe they will come back.

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
  -->

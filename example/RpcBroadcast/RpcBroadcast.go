package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
)

var done chan int = make(chan int, 1)
var ports = [3]string{"8080", "8081", "8082"}

// Args are mathematical arguments for the rpc operations
type Args struct {
	A, B int
}

// Quotient is the result of a Divide RPC
type Quotient struct {
	Quo, Rem int
}

// Arith is an RPC math server type
type Arith int

// Multiply performs multiplication on two integers
func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

// Divide divides a by b and returns a quotient with a remainder
func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

func rpcServer(no string, port string) {
	fmt.Println("Starting server no. " + no + " on port " + port)
	logger := govec.InitGoVector("server"+no, "server"+no+"logfile", govec.GetDefaultConfig())
	arith := new(Arith)
	server := rpc.NewServer()
	server.Register(arith)
	l, e := net.Listen("tcp", ":"+port)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	options := govec.GetDefaultLogOptions()
	vrpc.ServeRPCConn(server, l, logger, options)
}

func rpcClient() {
	fmt.Println("Starting client")
	logger := govec.InitGoVector("client", "clientlogfile", govec.GetDefaultConfig())
	options := govec.GetDefaultLogOptions()

	var clients [3]*rpc.Client
	var err error
	for i, port := range ports {
		clients[i], err = vrpc.RPCDial("tcp", "127.0.0.1:"+port, logger, options)
		if err != nil {
			log.Fatal(err)
		}
	}

	var calls [3]*rpc.Call
	var results [3]int
	logger.StartBroadcast("Broadcasting via RPC", options)
	for i, client := range clients {
		calls[i] = client.Go("Arith.Multiply", Args{i, i + 1}, &results[i], nil)
	}
	logger.StopBroadcast()

	for i, call := range calls {
		<-call.Done
		fmt.Println("Received result", results[i], "from server no.", strconv.Itoa(i+1))
	}

	done <- 1
}

func main() {
	for i, port := range ports {
		go rpcServer(strconv.Itoa(i+1), port)
	}
	time.Sleep(time.Millisecond)
	go rpcClient()
	<-done
}

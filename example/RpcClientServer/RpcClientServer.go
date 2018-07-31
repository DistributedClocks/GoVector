package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
)

var done chan int = make(chan int, 1)

//Args are matematical arguments for the rpc operations
type Args struct {
	A, B int
}

//Quotient is the result of a Divide RPC
type Quotient struct {
	Quo, Rem int
}

//Arith is an RPC math server type
type Arith int

//Multiply performs multiplication on two integers
func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

//Divide divides a by b and returns a quotient with a remainder
func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

func rpcserver() {
	fmt.Println("Starting server")
	logger := govec.InitGoVector("server", "serverlogfile", govec.GetDefaultConfig())
	arith := new(Arith)
	server := rpc.NewServer()
	server.Register(arith)
	l, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	vrpc.ServeRPCConn(server, l, logger)
}

func rpcclient() {
	fmt.Println("Starting client")
	logger := govec.InitGoVector("client", "clientlogfile", govec.GetDefaultConfig())
	client, err := vrpc.RPCDial("tcp", "127.0.0.1:8080", logger)
	if err != nil {
		log.Fatal(err)
	}
	var result int
	err = client.Call("Arith.Multiply", Args{5, 6}, &result)
	if err != nil {
		log.Fatal(err)
	}

	var qresult Quotient
	err = client.Call("Arith.Divide", Args{4, 2}, &qresult)
	if err != nil {
		log.Fatal(err)
	}
	done <- 1
}

func main() {
	go rpcserver()
	time.Sleep(time.Millisecond)
	go rpcclient()
	<-done
}

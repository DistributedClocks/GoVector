package main

import (
	"errors"
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
	"log"
	"net"
	"net/rpc"
	"time"
)

var done chan int = make(chan int, 1)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

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

package vrpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"testing"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
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

func rpcserver(logger *govec.GoLog) {
	fmt.Println("Starting server")
	arith := new(Arith)
	server := rpc.NewServer()
	server.Register(arith)
	l, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	options := govec.GetDefaultLogOptions()
	ServeRPCConn(server, l, logger, options)
}

func rpcclient(logger *govec.GoLog) {
	fmt.Println("Starting client")
	options := govec.GetDefaultLogOptions()
	client, err := RPCDial("tcp", "127.0.0.1:8080", logger, options)
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

func TestRPC(t *testing.T) {
	serverlogger := govec.InitGoVector("server", "serverlogfile", govec.GetDefaultConfig())
	clientlogger := govec.InitGoVector("client", "clientlogfile", govec.GetDefaultConfig())
	go rpcserver(serverlogger)
	time.Sleep(time.Millisecond)
	go rpcclient(clientlogger)
	<-done
	server_vc := serverlogger.GetCurrentVC()
	server_ticks, _ := server_vc.FindTicks("server")
	client_vc := clientlogger.GetCurrentVC()
	client_ticks, _ := client_vc.FindTicks("client")

	AssertEquals(t, uint64(5), server_ticks, "Server Clock value not incremented")
	AssertEquals(t, uint64(5), client_ticks, "Client Clock value not incremented")
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf(message+"Expected: %s, Actual: %s", expected, actual)
	}
}

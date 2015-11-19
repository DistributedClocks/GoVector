
package main

import (
	"golang.org/x/net/websocket"
	//"net/http"
	//"./rpc"
	"./rpc/jsonrpc"
	"fmt"
    //"errors"
    //"log"
	//"time"
)

type Args struct {
	A int
	B int
}

func main() { 
	conn, err := websocket.Dial("ws://localhost:7000/conn", "", "http://localhost/")
		
    if err != nil {
        panic(err)
    }
    defer conn.Close()
	
	addr := conn.RemoteAddr()
	fmt.Println(addr.String())
    c := jsonrpc.NewClient(conn)
	//time.Sleep(100 * time.Millisecond)
	// passing Args to RPC call
    var reply *int
    var args *Args
    args = &Args{7, 2}

    // calling "Arith.Mul" on RPC server
    err = c.Call("Arith.Multiply", args, &reply)
    if err != nil {
        fmt.Println(err)
		c.Close()
    }
    fmt.Printf("Arith: %d * %d = %d\n", args.A, args.B, reply)
}
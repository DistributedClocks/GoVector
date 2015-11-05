package main

import (
	"golang.org/x/net/websocket"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"fmt"
    "errors"
    "log"
)

type Args struct {
    A, B int
}

type Reply struct {
    C int
}

type Arith int

type ArithAddResp struct {
    Id     interface{} `json:"id"`
    Result Reply       `json:"result"`
    Error  interface{} `json:"error"`
}

func (t *Arith) Add(args *Args, reply *Reply) error {
	fmt.Println("In add")
    reply.C = args.A + args.B
    return nil
}

func (t *Arith) Mul(args *Args, reply *Reply) error {
	fmt.Println("In mul")
    reply.C = args.A * args.B
    return nil
}

func (t *Arith) Div(args *Args, reply *Reply) error {
	fmt.Println("In div")
    if args.B == 0 {
        return errors.New("divide by zero")
    }
    reply.C = args.A / args.B
    return nil
}

func (t *Arith) Error(args *Args, reply *Reply) error {
    panic("ERROR")
}

func startServer() {
    arith := new(Arith)

    server := rpc.NewServer()
    server.Register(arith)
	//rpc.Register(arith)
	http.Handle("/ws", websocket.Handler(serve))
	http.ListenAndServe("localhost:8000", nil)
}

func serve(ws *websocket.Conn) {
	go jsonrpc.ServeConn(ws)
}
//    l, e := net.Listen("tcp", ":8222")
//    if e != nil {
//        log.Fatal("listen error:", e)
//    }

//    for {
//        conn, err := l.Accept()
//        if err != nil {
//            log.Fatal(err)
//        }

//        go server.ServeCodec(jsonrpc.NewServerCodec(conn))
//    }
//}

func main() {

    // starting server in go routine (it ends on end
    // of main function
    go startServer()

    // now client part connecting to RPC service
    // and calling methods

    conn, err := websocket.Dial("ws://127.0.0.1:8000/ws", "", "http://127.0.0.1/")

    if err != nil {
        panic(err)
    }
    defer conn.Close()

    c := jsonrpc.NewClient(conn)

    var reply Reply
    var args *Args
    for i := 0; i < 11; i++ {
        // passing Args to RPC call
        args = &Args{7, i}

        // calling "Arith.Mul" on RPC server
        err = c.Call("Arith.Mul", args, &reply)
        if err != nil {
            log.Fatal("arith error:", err)
        }
        fmt.Printf("Arith: %d * %d = %v\n", args.A, args.B, reply.C)

        // calling "Arith.Add" on RPC server
        err = c.Call("Arith.Add", args, &reply)
        if err != nil {
            log.Fatal("arith error:", err)
        }
        fmt.Printf("Arith: %d + %d = %v\n", args.A, args.B, reply.C)

        // NL
        fmt.Printf("\033[33m%s\033[m\n", "---------------")

    }
}
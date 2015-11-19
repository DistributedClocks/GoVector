package main

import (
	"golang.org/x/net/websocket"
	"net/http"
	"./rpc"
	//"./rpc/jsonrpc"
	"encoding/json"
	"fmt"
    //"errors"
    //"log"
	//"time"
)

type Args struct {
	A int
	B int
}

type Arith int 

func (t *Arith) Multiply(args *Args, reply *int) error { 
        *reply = args.A * args.B 
        return nil 
} 

func main() { 
    rpc.Register(new(Arith)) 

	fmt.Println("Starting server")
	http.Handle("/ws", websocket.Handler(serve)) 
	http.HandleFunc("/", staticFiles)
    http.ListenAndServe("localhost:8080", nil) 

			
} 

func serve(ws *websocket.Conn) { 
//	var msg string
//	err := websocket.Message.Receive(ws, &msg)
//	if err != nil {
//		fmt.Println("Error reading connection:", err)
//		ws.Close()
//		return
//	}
//	fmt.Println(msg)
	var msg Blob
	d := json.NewDecoder(ws)
    wserr := d.Decode(&msg)
//	test := &Blob {
//		Method: "Arith.Multiply", 
//		Params: []int{5,8},
//		Id: 0}
	//tmp1, err := json.Marshal(test)
	//tmp := make([]byte, 256)
	//n, wserr := ws.Read(tmp)
	//json.Unmarshal(tmp[:n], &msg)
	//json.Unmarshal(tmp1, &msg)
	//s := string(tmp)
	fmt.Println("Message: ", msg, "error: ", wserr)
	//fmt.Println("Serving connection")
    //go jsonrpc.ServeConn(ws) 
} 

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/"+r.URL.Path)
}

type Blob struct {
	
	Method string 
	Params []int
	Id int64
}
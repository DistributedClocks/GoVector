package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/arcaneiceman/GoVector/govec"
)

const (
	SERVERPORT = "8080"
	CLIENTPORT = "8081"
	MESSAGES   = 5
)

var done chan int = make(chan int, 2)

func main() {
	go server(SERVERPORT)
	go client(CLIENTPORT, SERVERPORT)
	<-done
	<-done
}

//Msg is the message sent over the network
//Msg is capitolized so GoVecs encoder can acess it
//Furthermore its variables are capitolized to make them public
type Msg struct {
	Content, RealTimestamp string
}

func (m Msg) String() string {
	return "content: " + m.Content + "\ntime: " + m.RealTimestamp
}

func client(listen, send string) {
	Logger := govec.Initialize("client", "clientlogfile")
	// sending UDP packet to specified address and port
	conn := setupConnection(SERVERPORT, CLIENTPORT)

	for i := 0; i < MESSAGES; i++ {
		outgoingMessage := Msg{"Hello GoVec!", time.Now().String()}
		outBuf := Logger.PrepareSend("Sending message to server", outgoingMessage)
		_, errWrite := conn.Write(outBuf)
		printErr(errWrite)

		var inBuf [512]byte
		n, errRead := conn.Read(inBuf[0:])
		printErr(errRead)
		incommingMessage := new(Msg)
		Logger.UnpackReceive("Received Message from server", inBuf[0:n], &incommingMessage)
		fmt.Println(incommingMessage.String())
		time.Sleep(1)
	}
	done <- 1

}

func server(listen string) {
	Logger := govec.Initialize("server", "server")
	conn, err := net.ListenPacket("udp", ":"+listen)
	printErr(err)

	var buf [512]byte

	for i := 0; i < MESSAGES; i++ {
		_, addr, err := conn.ReadFrom(buf[0:])
		incommingMessage := new(Msg)
		Logger.UnpackReceive("Received Message From Client", buf[0:], &incommingMessage)
		fmt.Println(incommingMessage.String())
		printErr(err)

		outgoingMessage := Msg{"GoVecs Great?", time.Now().String()}
		conn.WriteTo(Logger.PrepareSend("Replying to client", outgoingMessage), addr)
		time.Sleep(1)
	}
	conn.Close()
	done <- 1

}

func setupConnection(sendingPort, listeningPort string) *net.UDPConn {
	rAddr, errR := net.ResolveUDPAddr("udp4", ":"+sendingPort)
	printErr(errR)
	lAddr, errL := net.ResolveUDPAddr("udp4", ":"+listeningPort)
	printErr(errL)

	conn, errDial := net.DialUDP("udp", lAddr, rAddr)
	printErr(errDial)
	if (errR == nil) && (errL == nil) && (errDial == nil) {
		return conn
	}
	return nil
}

func printErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"net"
	"os"
	"time"
)

const (
	SERVERPORT = "8080"
	CLIENTPORT = "8081"
	MESSAGES   = 10
)

var done chan int = make(chan int, 2)

func main() {
	go server(SERVERPORT)
	go client(CLIENTPORT, SERVERPORT)
	<-done
	<-done
}

func client(listen, send string) {
	Logger := govec.InitGoVector("client", "clientlogfile", govec.GetDefaultConfig())
	// sending UDP packet to specified address and port
	conn := setupConnection(SERVERPORT, CLIENTPORT)

	for i := 0; i < MESSAGES; i++ {
		outgoingMessage := i
		outBuf := Logger.PrepareSend("Sending message to server", outgoingMessage)
		_, errWrite := conn.Write(outBuf)
		printErr(errWrite)

		var inBuf [512]byte
		var incommingMessage int
		n, errRead := conn.Read(inBuf[0:])
		printErr(errRead)
		Logger.UnpackReceive("Received Message from server", inBuf[0:n], &incommingMessage)
		fmt.Printf("GOT BACK : %d\n", incommingMessage)
		time.Sleep(1)
	}
	done <- 1

}

func server(listen string) {

	Logger := govec.InitGoVector("server", "server", govec.GetDefaultConfig())

	fmt.Println("Listening on server....")
	conn, err := net.ListenPacket("udp", ":"+listen)
	printErr(err)

	var buf [512]byte

	var n, nMinOne, nMinTwo int
	n = 1
	nMinTwo = 1
	nMinTwo = 1

	for i := 0; i < MESSAGES; i++ {
		_, addr, err := conn.ReadFrom(buf[0:])
		var incommingMessage int
		Logger.UnpackReceive("Received Message From Client", buf[0:], &incommingMessage)
		fmt.Printf("Received %d\n", incommingMessage)
		printErr(err)

		switch incommingMessage {
		case 0:
			nMinTwo = 0
			n = 0
			break
		case 1:
			nMinOne = 0
			n = 1
			break
		default:
			nMinTwo = nMinOne
			nMinOne = n
			n = nMinOne + nMinTwo
			break
		}

		outBuf := Logger.PrepareSend("Replying to client", n)

		conn.WriteTo(outBuf, addr)
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

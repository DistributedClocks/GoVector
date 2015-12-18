package brokervec

import "net"

type Publisher interface {
    GetName() string
}

type TCPPub struct {
    Name      string
    Conn      net.Conn 
}

func (tp *TCPPub) GetName() (name string) {
    return tp.Name
}
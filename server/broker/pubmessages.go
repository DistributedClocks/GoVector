package brokervec

import (
    "time"
    "github.com/arcaneiceman/GoVector/govec/vclock"
)

// Message abstraction to hold different types of message for use in the broker
type Message interface {
    GetMessage()     string
    GetNonce()       string
    GetTime()        time.Time
}

// A LogMessage is an internal message for the broker's log file.
// It should not be sent to subscribers or received from publishers.
type LogMessage struct {
    Message      string
    ReceiptTime  time.Time
}
func (logm *LogMessage) GetMessage() string {
    return logm.ReceiptTime.String() + " " + logm.Message + "\n"
}

func (logm *LogMessage) GetNonce() string {
    return "Log does not have a nonce."
}

func (logm *LogMessage) GetTime() time.Time {
    return logm.ReceiptTime
}

// A LocalMessage is a message that was not sent over the network in the 
// distributive system using GoVec.
type LocalMessage struct {
    Pid         string 
    Vclock      []byte
    Message     string
    Nonce       string
    ReceiptTime time.Time
}

func (lm *LocalMessage) GetMessage() string {
    vc, err := vclock.FromBytes(lm.Vclock)

    if err != nil {
        panic(err)
    }
    return lm.Pid + " " + vc.ReturnVCString() + "\n" + lm.Message + "\n"
}
func (lm *LocalMessage) GetNonce() string {
    return lm.Nonce
}

func (lm *LocalMessage) GetTime() time.Time {
    return lm.ReceiptTime
}

// A NetworkMessage is a message that was sent over the network in the
// system using GoVec.
type NetworkMessage struct {
    Pid          string
    Vclock       []byte
    Message      string
    Nonce        string
    ReceiptTime  time.Time
}

func (nm *NetworkMessage) GetMessage() string {
    vc, err := vclock.FromBytes(nm.Vclock)
    
    if err != nil {
        panic(err)
    }
    return nm.Pid + " " + vc.ReturnVCString() + "\n" + nm.Message + "\n"
}
func (nm *NetworkMessage) GetNonce() string {
        
    return nm.Nonce
}

func (nm *NetworkMessage) GetTime() time.Time {
        
    return nm.ReceiptTime
}

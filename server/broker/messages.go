package brokervec

import (

)

// Message abstraction to hold different types of message
type Message interface {
	GetMessage()	string
}

type LogMessage struct {
	Message string
}
func (logm *LogMessage) GetMessage() string {
	return logm.Message
}

type LocalMessage struct {
	Pid string // pid
	Vclock string
	Message string
}

func (lm *LocalMessage) GetMessage() string {
		
	return lm.Pid + " " + lm.Vclock + "\n" + lm.Message + "\n"
}

type NetworkMessage struct {
	Pid string // pid
	Vclock string
	Message string
}

func (nm *NetworkMessage) GetMessage() string {
		
	return nm.Pid + " " + nm.Vclock + "\n" + nm.Message + "\n"
}
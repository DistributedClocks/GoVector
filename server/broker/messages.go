package brokervec

import (
	//"./nonce"
)

// Message abstraction to hold different types of message
type Message interface {
	GetMessage()	string
	GetNonce()		string
}

type LogMessage struct {
	Message string
	Nonce	string
}
func (logm *LogMessage) GetMessage() string {
	return logm.Message + "\n"
}

func (logm *LogMessage) GetNonce() string {
	return logm.Nonce
}

type LocalMessage struct {
	Pid string // pid
	Vclock string
	Message string
	Nonce string
}

func (lm *LocalMessage) GetMessage() string {
		
	return lm.Pid + " " + lm.Vclock + "\n" + lm.Message + "\n"
}
func (lm *LocalMessage) GetNonce() string {
	return lm.Nonce
}

type NetworkMessage struct {
	Pid string // pid
	Vclock string
	Message string
	Nonce string
}

func (nm *NetworkMessage) GetMessage() string {
		
	return nm.Pid + " " + nm.Vclock + "\n" + nm.Message + "\n"
}
func (nm *NetworkMessage) GetNonce() string {
		
	return nm.Nonce
}

type FilterMessage struct {
	Message string
	Nonce string
}

func (fm *FilterMessage) GetMessage() string {
		
	return fm.Message
}
func (fm *FilterMessage) GetNonce() string {
		
	return fm.Nonce
}
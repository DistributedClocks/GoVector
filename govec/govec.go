// Package govec is a vector clock logging library
package govec

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	logToTerminal                       = false
	_             msgpack.CustomEncoder = (*VClockPayload)(nil)
	_             msgpack.CustomDecoder = (*VClockPayload)(nil)
)

// LogPriority controls the minimum priority of logging events which
// will be logged.
type LogPriority int

// LogPriority enum provides all the valid Priority Levels that can be
// used to log events with.
const (
	DEBUG LogPriority = iota
	INFO
	WARNING
	ERROR
	FATAL
)

// colorLookup is the array of status to string from runtime/proc.go
var colorLookup = [...]ct.Color{
	DEBUG:   ct.Green,
	INFO:    ct.White,
	WARNING: ct.Yellow,
	ERROR:   ct.Red,
	FATAL:   ct.Magenta,
}

// prefixLookup translates priority enums into strings
var prefixLookup = [...]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARNING: "WARNING",
	ERROR:   "ERROR",
	FATAL:   "FATAL",
}

// GoLogConfig controls the logging parameters of GoLog and is taken as
// input to GoLog initialization. See defaults in GetDefaultConfig.
type GoLogConfig struct {
	// Buffered denotes if the logging events are buffered until flushed. This option
	// increase logging performance at the cost of safety.
	Buffered bool
	// PrintOnScreen denotes if logging events are printed to screen.
	PrintOnScreen bool
	// AppendLog determines to continue writing to a log from a prior execution.
	AppendLog bool
	// UseTimestamps determines to log real time timestamps for TSVis
	UseTimestamps bool
	// EncodingStrategy for customizable interoperability
	EncodingStrategy func(interface{}) ([]byte, error)
	// DecodingStrategy for customizable interoperability
	DecodingStrategy func([]byte, interface{}) error
	// LogToFile determines to write logging events to a file
	LogToFile bool
	// Priority determines the minimum priority event to log
	Priority LogPriority
	// InitialVC is the initial vector clock value, nil by default
	InitialVC vclock.VClock
}

// GetDefaultConfig returns the default GoLogConfig with default values
// for various fields.
func GetDefaultConfig() GoLogConfig {
	config := GoLogConfig{
		Buffered:      false,
		PrintOnScreen: false,
		AppendLog:     false,
		UseTimestamps: false,
		LogToFile:     true,
		Priority:      INFO,
		InitialVC:     nil,
	}
	return config
}

// GoLogOptions provides logging parameters for individual logging statements
type GoLogOptions struct {
	// The Log priority for this event
	Priority LogPriority
}

// GetDefaultLogOptions returns the default GoLogOptions with default values
// for the fields
func GetDefaultLogOptions() GoLogOptions {
	o := GoLogOptions{Priority: INFO}
	return o
}

// SetPriority returns a new GoLogOptions object with its priority field
// set to Priority. Follows the builder pattern.
// Priority : (GoLogPriority) The Priority that the new GoLogOptions object must have
func (o *GoLogOptions) SetPriority(Priority LogPriority) GoLogOptions {
	opts := *o
	opts.Priority = Priority
	return opts
}

// VClockPayload is the data structure that is actually end on the wire
type VClockPayload struct {
	Pid     string
	VcMap   map[string]uint64
	Payload interface{}
}

// PrintDataBytes prints the Data Stuct as Bytes
func (d *VClockPayload) PrintDataBytes() {
	fmt.Printf("%x \n", d.Pid)
	fmt.Printf("%X \n", d.VcMap)
	fmt.Printf("%X \n", d.Payload)
}

// String returns VClockPayload's pid as a string
func (d *VClockPayload) String() (s string) {
	s += "-----DATA START -----\n"
	s += string(d.Pid[:])
	s += "-----DATA END -----\n"
	return
}

// EncodeMsgpack is a custom encoder function, needed for msgpack interoperability
func (d *VClockPayload) EncodeMsgpack(enc *msgpack.Encoder) error {
	var err error

	err = enc.EncodeString(d.Pid)
	if err != nil {
		return err
	}

	err = enc.Encode(d.Payload)
	if err != nil {
		return err
	}

	err = enc.EncodeMapLen(len(d.VcMap))
	if err != nil {
		return err
	}

	for key, value := range d.VcMap {

		err = enc.EncodeString(key)
		if err != nil {
			return err
		}

		err = enc.EncodeUint(value)
		if err != nil {
			return err
		}
	}

	return nil

}

// DecodeMsgpack is a custom decoder function, needed for msgpack
// interoperability
func (d *VClockPayload) DecodeMsgpack(dec *msgpack.Decoder) error {
	var err error

	pid, err := dec.DecodeString()
	if err != nil {
		return err
	}
	d.Pid = pid

	err = dec.Decode(&d.Payload)
	if err != nil {
		return err
	}

	mapLen, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}
	var vcMap map[string]uint64
	vcMap = make(map[string]uint64)

	for i := 0; i < mapLen; i++ {

		key, err := dec.DecodeString()
		if err != nil {
			return err
		}

		value, err := dec.DecodeUint64()
		if err != nil {
			return err
		}
		vcMap[key] = value
	}
	d.VcMap = vcMap

	return nil
}

// GoLog struct provides an interface to creating and maintaining
// vector timestamp entries in the generated log file. GoLogs are
// initialized with Configs which control logger output, format, and
// frequency.
type GoLog struct {

	// Local Process ID
	pid string

	// Local vector clock in bytes
	currentVC vclock.VClock

	// Flag to Printf the logs made by Local Program
	printonscreen bool

	// If true GoLog will write to a file
	logging bool

	// If true logs are buffered in memory and flushed to disk upon
	// calling flush. Logs can be lost if a program stops prior to
	// flushing buffered logs.
	buffered bool

	// Flag to include timestamps when logging events
	usetimestamps bool

	// Flag to indicate if the log file will contain multiple executions
	appendLog bool

	// Flag to indicate if the message broadcast is on
	broadcast bool

	// Priority level at which all events are logged
	priority LogPriority

	// Logfile name
	logfile string

	// buffered string
	output string

	// encoding and decoding strategies for network messages
	encodingStrategy func(interface{}) ([]byte, error)
	decodingStrategy func([]byte, interface{}) error

	// Internal logger for printing errors
	logger *log.Logger

	mutex sync.RWMutex
}

// InitGoVector returns a GoLog which generates a logs prefixed with
// processid, to a file name logfilename.log. Any old log with the same
// name will be trucated. Config controls logging options. See GoLogConfig for more details.
func InitGoVector(processid string, logfilename string, config GoLogConfig) *GoLog {

	gv := &GoLog{}
	gv.pid = processid

	if logToTerminal {
		gv.logger = log.New(os.Stdout, "[GoVector]:", log.Lshortfile)
	} else {
		var buf bytes.Buffer
		gv.logger = log.New(&buf, "[GoVector]:", log.Lshortfile)
	}

	//Set parameters from config
	gv.printonscreen = config.PrintOnScreen
	gv.usetimestamps = config.UseTimestamps
	gv.priority = config.Priority
	gv.logging = config.LogToFile
	gv.buffered = config.Buffered
	gv.appendLog = config.AppendLog
	gv.output = ""

	// Use the default encoder/decoder. As of July 2017 this is msgPack.
	if config.EncodingStrategy == nil || config.DecodingStrategy == nil {
		gv.setEncoderDecoder(defaultEncoder, defaultDecoder)
	} else {
		gv.setEncoderDecoder(config.EncodingStrategy, config.DecodingStrategy)
	}

	// We create a new Vector Clock with processname and config.InitialClock
	// as the initial time
	var vc1 vclock.VClock
	if config.InitialVC == nil {
		vc1 = vclock.New()
		vc1.Set(processid, 0)
	} else {
		vc1 = config.InitialVC
	}
	gv.currentVC = vc1

	//Starting File IO . If Log exists, Log Will be deleted and A New one will be created
	logname := logfilename + "-Log.txt"
	gv.logfile = logname
	if gv.logging {
		gv.prepareLogFile()
	}

	return gv
}

func (gv *GoLog) prepareLogFile() {
	_, err := os.Stat(gv.logfile)
	if err == nil {
		if !gv.appendLog {
			gv.logger.Println(gv.logfile, "exists! ... Deleting ")
			os.Remove(gv.logfile)
		} else {
			executionnumber := time.Now().Format(time.UnixDate)
			gv.logger.Println("Execution Number is  ", executionnumber)
			executionstring := "=== Execution #" + executionnumber + "  ==="
			gv.logThis(executionstring, "", "", gv.priority)
			return
		}
	}
	// Create directory path to log if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(gv.logfile), 0750); err != nil {
		gv.logger.Println(err)
	}

	//Creating new Log
	file, err := os.Create(gv.logfile)
	if err != nil {
		gv.logger.Println(err)
	}

	file.Close()

	if gv.appendLog {
		executionnumber := time.Now().Format(time.UnixDate)
		gv.logger.Println("Execution Number is  ", executionnumber)
		executionstring := "=== Execution #" + executionnumber + "  ==="
		gv.logThis(executionstring, "", "", gv.priority)
	}

	gv.currentVC.Tick(gv.pid)
	ok := gv.logThis("Initialization Complete", gv.pid, gv.currentVC.ReturnVCString(), gv.priority)
	if ok == false {
		gv.logger.Println("Something went Wrong, Could not Log!")
	}
}

// GetCurrentVC returns the current vector clock
func (gv *GoLog) GetCurrentVC() vclock.VClock {
	return gv.currentVC
}

// setEncoderDecoder Sets the Encoding and Decoding functions which are to be used by the logger
// Encoder (func(interface{}) ([]byte, error)) : function to be used for encoding
// Decoder (func([]byte, interface{}) error) : function to be used for decoding
func (gv *GoLog) setEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error) {
	gv.encodingStrategy = encoder
	gv.decodingStrategy = decoder
}

// By default, encoding is performed by msgpack
func defaultEncoder(payload interface{}) ([]byte, error) {
	return msgpack.Marshal(payload)
}

// By default, decoding network payloads is perfomed by msgpack
func defaultDecoder(buf []byte, payload interface{}) error {
	return msgpack.Unmarshal(buf, payload)
}

// EnableBufferedWrites enables buffered writes to the log file. All
// the log messages are only written to the LogFile via an explicit
// call to the function Flush.  Note: Buffered writes are automatically
// disabled.
func (gv *GoLog) EnableBufferedWrites() {
	gv.buffered = true
}

// DisableBufferedWrites disables buffered writes to the log file. All
// the log messages from now on will be written to the Log file
// immediately. Writes all the existing log messages that haven't been
// written to Log file yet.
func (gv *GoLog) DisableBufferedWrites() {
	gv.buffered = false
	if gv.output != "" {
		gv.Flush()
	}
}

// Flush writes the log messages stored in the buffer to the Log File.
// This function should be used by the application to also force writes
// in the case of interrupts and crashes.   Note: Calling Flush when
// BufferedWrites is disabled is essentially a no-op.
func (gv *GoLog) Flush() bool {
	complete := true
	file, err := os.OpenFile(gv.logfile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		complete = false
	}
	defer file.Close()

	if _, err = file.WriteString(gv.output); err != nil {
		complete = false
	}

	gv.output = ""
	return complete
}

func (gv *GoLog) printColoredMessage(LogMessage string, Priority LogPriority) {
	color := colorLookup[Priority]
	prefix := prefixLookup[Priority]
	ct.Foreground(color, true)
	fmt.Print(time.Now().Format(time.UnixDate) + ":" + prefix + "-")
	ct.ResetColor()
	fmt.Println(LogMessage)
}

// Logs a message along with a processID and a vector clock, the VCString
// must be a valid vector clock, true is returned on success. logThis
// is the innermost logging function internally used by all other
// logging functions
func (gv *GoLog) logThis(Message string, ProcessID string, VCString string, Priority LogPriority) bool {
	var (
		complete = true
		buffer   bytes.Buffer
	)
	if gv.usetimestamps {
		buffer.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
		buffer.WriteString(" ")
	}
	buffer.WriteString(ProcessID)
	buffer.WriteString(" ")
	buffer.WriteString(VCString)
	buffer.WriteString("\n")
	buffer.WriteString(Message)
	buffer.WriteString("\n")
	output := buffer.String()

	gv.output += output
	if !gv.buffered {
		complete = gv.Flush()
	}

	if gv.printonscreen == true {
		gv.printColoredMessage(Message, Priority)
	}
	return complete
}

// logWriteWrapper is a helper function for wrapping common logging
// behaviour associated with logThis
func (gv *GoLog) logWriteWrapper(mesg, errMesg string, priority LogPriority) (success bool) {
	if gv.logging == true {
		prefix := prefixLookup[priority]
		wrappedMesg := prefix + " " + mesg
		success = gv.logThis(wrappedMesg, gv.pid, gv.currentVC.ReturnVCString(), priority)
		if !success {
			gv.logger.Println(errMesg)
		}
	}
	return
}

// Increment GoVectors local clock by 1
func (gv *GoLog) tickClock() {
	_, found := gv.currentVC.FindTicks(gv.pid)
	if !found {
		gv.logger.Println("Couldn't find this process's id in its own vector clock!")
	}
	gv.currentVC.Tick(gv.pid)
}

// LogLocalEvent implements LogLocalEvent with priority
// levels. If the priority of the logger is lower than or equal to the
// priority of this event then the current vector timestamp is
// incremented and the message is logged it into the Log File. A color
// coded string is also printed on the console.
// * LogMessage (string) : Message to be logged
// * Priority (LogPriority) : Priority at which the message is to be logged
func (gv *GoLog) LogLocalEvent(mesg string, opts GoLogOptions) (logSuccess bool) {
	logSuccess = true
	gv.mutex.Lock()
	if opts.Priority >= gv.priority {
		gv.tickClock()
		logSuccess = gv.logWriteWrapper(mesg, "Something went Wrong, Could not Log LocalEvent!", opts.Priority)
	}
	gv.mutex.Unlock()
	return
}

// PrepareSend is meant to be used immediately before sending.
// LogMessage will be logged along with the time of send
// buf is encode-able data (structure or basic type)
// Returned is an encoded byte array with logging information.
// This function is meant to be called before sending a packet. Usually,
// it should Update the Vector Clock for its own process, package with
// the clock using gob support and return the new byte array that should
// be sent onwards using the Send Command
func (gv *GoLog) PrepareSend(mesg string, buf interface{}, opts GoLogOptions) (encodedBytes []byte) {
	//Converting Vector Clock from Bytes and Updating the gv clock
	if !gv.broadcast {
		gv.mutex.Lock()
		if opts.Priority >= gv.priority {
			gv.tickClock()

			gv.logWriteWrapper(mesg, "Something went wrong, could not log prepare send", opts.Priority)

			d := VClockPayload{Pid: gv.pid, VcMap: gv.currentVC.GetMap(), Payload: buf}

			// encode the Clock Payload
			var err error
			encodedBytes, err = gv.encodingStrategy(&d)
			if err != nil {
				gv.logger.Println(err.Error())
			}

			// return encodedBytes which can be sent off and received on the other end!
		}
		gv.mutex.Unlock()

	} else {
		// Broadcast: do not acquire the lock, tick the clock or log an event as
		// all been done by StartBroadcast
		d := VClockPayload{Pid: gv.pid, VcMap: gv.currentVC.GetMap(), Payload: buf}

		var err error
		encodedBytes, err = gv.encodingStrategy(&d)
		if err != nil {
			gv.logger.Println(err.Error())
		}
	}
	return
}

func (gv *GoLog) mergeIncomingClock(mesg string, e VClockPayload, Priority LogPriority) {
	// First, tick the local clock
	gv.tickClock()
	gv.currentVC.Merge(e.VcMap)

	gv.logWriteWrapper(mesg, "Something went Wrong, Could not Log!", Priority)
}

// UnpackReceive is used to unmarshall network data into local structures.
// LogMessage will be logged along with the vector time receive happened
// buf is the network data, previously packaged by PrepareSend unpack is
// a pointer to a structure, the same as was packed by PrepareSend.
// This function is meant to be called immediately after receiving
// a packet. It unpacks the data by the program, the vector clock. It
// updates vector clock and logs it. and returns the user data
func (gv *GoLog) UnpackReceive(mesg string, buf []byte, unpack interface{}, opts GoLogOptions) {
	gv.mutex.Lock()

	if opts.Priority >= gv.priority {
		e := VClockPayload{}
		e.Payload = unpack

		// Just use msgpack directly
		err := gv.decodingStrategy(buf, &e)
		if err != nil {
			gv.logger.Println(err.Error())
		}

		// Increment and merge the incoming clock
		gv.mergeIncomingClock(mesg, e, opts.Priority)
	}
	gv.mutex.Unlock()

}

// StartBroadcast allows to use vector clocks in the context of casual broadcasts
// sent via RPC. Any call to StartBroadcast must have a corresponding call to 
// StopBroadcast, otherwise a deadlock will occur. All RPC calls made in-between 
// the calls to StartBroadcast and StopBroadcast will be logged as a single event, 
// will be sent out with the same vector clock and will represent broadcast messages
// from the current process to the process pool.
func (gv *GoLog) StartBroadcast(mesg string, opts GoLogOptions) {
	gv.mutex.Lock()
	gv.tickClock()
	gv.logWriteWrapper(mesg, "Something went wrong, could not log prepare send", opts.Priority)
	gv.broadcast = true
}

// StopBroadcast is called once all RPC calls of a message broadcast have been sent.
func (gv *GoLog) StopBroadcast() {
	gv.broadcast = false
	gv.mutex.Unlock()
}

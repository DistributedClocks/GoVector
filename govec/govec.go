package govec

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/vmihailenco/msgpack"
)

/*
   - All licences like other licenses ...

   How to Use This Library

   Step 1:
   Create a Global Variable and Initialize it using like this =

   Logger:= Initialize("MyProcess",ShouldYouSeeLoggingOnScreen,ShouldISendVectorClockonWire,Debug)

   Step 2:
   When Ever You Decide to Send any []byte , before sending call PrepareSend like this:
   SENDSLICE := PrepareSend("Message Description", YourPayload)
   and send the SENDSLICE instead of your Designated Payload

   Step 3:
   When Receiveing, AFTER you receive your message, pass the []byte into UnpackRecieve
   like this:

   UnpackReceive("Message Description", []ReceivedPayload, *RETURNSLICE)
   and use RETURNSLICE for further processing.
*/

var (
	logToTerminal                       = false
	_             msgpack.CustomEncoder = (*ClockPayload)(nil)
	_             msgpack.CustomDecoder = (*ClockPayload)(nil)
)

//The GoLog struct provides an interface to creating and maintaining
//vector timestamp entries in the generated log file
type GoLog struct {

	//Local Process ID
	pid string

	//Local vector clock in bytes
	currentVC vclock.VClock

	//Flag to Printf the logs made by Local Program
	printonscreen bool

	// This bool checks whether we should send logging data to a vecbroker whenever
	// a message is logged.
	realtime bool

	//activates/deactivates printouts at each preparesend and unpackreceive
	debugmode bool

	//If true GoLog will write to a file
	logging bool

	//If true logs are buffered in memory and flushed to disk upon
	//calling flush. Logs can be lost if a program stops prior to
	//flushing buffered logs.
	buffered bool

	//Logfile name
	logfile string

	//buffered string
	output string

	encodingStrategy func(interface{}) ([]byte, error)
	decodingStrategy func([]byte, interface{}) error

	//Internal Logging function
	logFunc func(message string, pid string, VCString string) bool

	logger *log.Logger

	mutex sync.RWMutex
}

//This is the data structure that is actually end on the wire
type ClockPayload struct {
	Pid     string
	VcMap   map[string]uint64
	Payload interface{}
}

//Prints the Data Stuct as Bytes
func (d *ClockPayload) PrintDataBytes() {
	fmt.Printf("%x \n", d.Pid)
	fmt.Printf("%X \n", d.VcMap)
	fmt.Printf("%X \n", d.Payload)
}

//Prints the Data Struct as a String
func (d *ClockPayload) String() (s string) {
	s += "-----DATA START -----\n"
	s += string(d.Pid[:])
	s += "-----DATA END -----\n"
	return
}

//Returns the current vector clock
func (gv *GoLog) GetCurrentVC() vclock.VClock {
	return gv.currentVC
}

//Returns a Go Log Struct taking in two arguments and truncates previous logs:
//MyProcessName (string): local process name; must be unique in your distributed system.
//LogFileName (string) : name of the log file that will store info. Any old log with the same name will be truncated
func InitGoVector(processid string, logfilename string) *GoLog {

	gv := &GoLog{}
	gv.pid = processid

	if logToTerminal {
		gv.logger = log.New(os.Stdout, "[GoVector]:", log.Lshortfile)
	} else {
		var buf bytes.Buffer
		gv.logger = log.New(&buf, "[GoVector]:", log.Lshortfile)
	}

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = true //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = false    // (Debug)
	gv.EnableLogging()
	gv.output = ""
	gv.DisableBufferedWrites()

	// Use the default encoder/decoder. As of July 2017 this is msgPack.
	gv.SetEncoderDecoder(gv.DefaultEncoder, gv.DefaultDecoder)

	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Tick(processid)
	gv.currentVC = vc1

	debugPrint(" ##### Initialization #####", vc1, gv)

	//Starting File IO . If Log exists, Log Will be deleted and A New one will be created
	logname := logfilename + "-Log.txt"

	if _, err := os.Stat(logname); err == nil {
		//its exists... deleting old log
		gv.logger.Println(logname, "exists! ... Deleting ")
		os.Remove(logname)
	}

	// Create directory path to log if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(logname), 0750); err != nil {
		gv.logger.Println(err)
	}

	//Creating new Log
	file, err := os.Create(logname)
	if err != nil {
		gv.logger.Println(err)
	}
	file.Close()

	gv.logfile = logname
	gv.logFunc = gv.logThis
	//Log it
	ok := gv.logFunc("Initialization Complete", gv.pid, vc1.ReturnVCString())
	if ok == false {
		gv.logger.Println("Something went Wrong, Could not Log!")
	}

	return gv
}

//Returns a Go Log Struct taking in two arguments without truncating previous log entry:
//* MyProcessName (string): local process name; must be unique in your distributed system.
//* LogFileName (string) : name of the log file that will store info. Each run will append to log file separated
//by "=== Execution # ==="
func InitGoVectorMultipleExecutions(processid string, logfilename string) *GoLog {

	gv := &GoLog{}
	gv.pid = processid

	if logToTerminal {
		gv.logger = log.New(os.Stdout, "[GoVector]:", log.Lshortfile)
	} else {
		var buf bytes.Buffer
		gv.logger = log.New(&buf, "[GoVector]:", log.Lshortfile)
	}

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = true //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = true     // (Debug)
	gv.EnableLogging()
	gv.output = ""
	gv.DisableBufferedWrites()

	// Use the default encoder/decoder. As of July 2017 this is msgPack.
	gv.SetEncoderDecoder(gv.DefaultEncoder, gv.DefaultDecoder)

	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Tick(processid)
	gv.currentVC = vc1

	debugPrint(" ###### Initialization ######", vc1, gv)
	//Starting File IO . If Log exists, it will find Last execution number and ++ it
	logname := logfilename + "-Log.txt"
	_, err := os.Stat(logname)
	gv.logfile = logname
	gv.logFunc = gv.logThis
	if err == nil {
		//its exists... deleting old log
		gv.logger.Println(logname, " exists! ...  Looking for Last Exectution... ")
		executionnumber := FindExecutionNumber(logname)
		executionnumber = executionnumber + 1
		gv.logger.Println("Execution Number is  ", executionnumber)
		executionstring := "=== Execution #" + strconv.Itoa(executionnumber) + "  ==="
		gv.logFunc(executionstring, "", "")
	} else {
		//Creating new Log
		file, err := os.Create(logname)
		if err != nil {
			gv.logger.Println(err.Error())
		}
		file.Close()

		//Mark Execution Number
		ok := gv.logFunc("=== Execution #1 ===", " ", " ")
		//Log it
		ok = gv.logFunc("Initialization Complete", gv.pid, vc1.ReturnVCString())
		if ok == false {
			gv.logger.Println("Something went Wrong, Could not Log!")
		}
	}
	return gv
}

func FindExecutionNumber(logname string) int {
	executionnumber := 0
	file, err := os.Open(logname)
	if err != nil {
		return 0
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for line, _, err := reader.ReadLine(); err != io.EOF; {
		if strings.Contains(string(line), " Execution #") {
			stemp := strings.Replace(string(line), "=== Execution #", "", -1)
			stemp = strings.Replace(stemp, " ===", "", -1)
			stemp = strings.Replace(stemp, " ", "", -1)
			currexecnumber, _ := strconv.Atoi(stemp)
			if currexecnumber > executionnumber {
				executionnumber = currexecnumber
			}
		}
		line, _, err = reader.ReadLine()
	}
	return executionnumber
}

func (gv *GoLog) SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error) {
	gv.encodingStrategy = encoder
	gv.decodingStrategy = decoder
}

//Enables buffered writes to the log file. All the log messages are only written
//to the LogFile via an explicit call to the function Flush.
//Note: Buffered writes are automatically disabled.
func (gv *GoLog) EnableBufferedWrites() {
	gv.buffered = true
}

//Disables buffered writes to the log file. All the log messages from now on
//will be written to the Log file immediately. Writes all the existing
//log messages that haven't been written to Log file yet.
func (gv *GoLog) DisableBufferedWrites() {
	gv.buffered = false
	if gv.output != "" {
		gv.Flush()
	}
}

/* Custom encoder function, needed for msgpack interoperability */
func (d *ClockPayload) EncodeMsgpack(enc *msgpack.Encoder) error {

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

/* Custom decoder function, needed for msgpack interoperability */
func (d *ClockPayload) DecodeMsgpack(dec *msgpack.Decoder) error {
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
	err = dec.Decode(&d.Pid, &d.Payload, &d.VcMap)
	d.VcMap = vcMap
	if err != nil {
		return err
	}

	return nil
}

func (gv *GoLog) DefaultEncoder(payload interface{}) ([]byte, error) {
	return msgpack.Marshal(payload)
}

func (gv *GoLog) DefaultDecoder(buf []byte, payload interface{}) error {
	return msgpack.Unmarshal(buf, payload)
}

//Writes the log messages stored in the buffer to the Log File. This
//function should be used by the application to also force writes in
//the case of interrupts and crashes.   Note: Calling Flush when
//BufferedWrites is disabled is essentially a no-op.
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

//Logs a message along with a processID and a vector clock, the VCString
//must be a valid vector clock, true is returned on success
func (gv *GoLog) logThis(Message string, ProcessID string, VCString string) bool {
	var (
		complete = true
		buffer   bytes.Buffer
	)
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
		gv.logger.Println(output)
	}
	return complete

}

//Increments current vector timestamp and logs it into Log File.
func (gv *GoLog) LogLocalEvent(Message string) (logSuccess bool) {
	gv.mutex.Lock()
	//Update the gv clock
	gv.tickClock()

	logSuccess = logWriteWrapper(Message, "Something went Wrong, Could not Log LocalEvent!", gv.currentVC, gv)
	gv.mutex.Unlock()

	return
}

/*
This function is meant to be used immediately before sending.
LogMessage will be logged along with the time of the send
buf is encode-able data (structure or basic type)
Returned is an encoded byte array with logging information.

This function is meant to be called before sending a packet. Usually,
it should Update the Vector Clock for its own process, package with
the clock using gob support and return the new byte array that should
be sent onwards using the Send Command
*/
func (gv *GoLog) PrepareSend(mesg string, buf interface{}) []byte {

	//Converting Vector Clock from Bytes and Updating the gv clock
	gv.mutex.Lock()
	gv.tickClock()

	debugPrint("Sending Message", gv.currentVC, gv)
	logWriteWrapper(mesg, "Something went wrong, could not log prepare send", gv.currentVC, gv)

	d := ClockPayload{Pid: gv.pid, VcMap: gv.currentVC.GetMap(), Payload: buf}

	// encode the Clock Payload
	encodedBytes, err := gv.encodingStrategy(&d)
	if err != nil {
		gv.logger.Println(err.Error())
	}

	// return encodedBytes which can be sent off and received on the other end!
	gv.mutex.Unlock()
	return encodedBytes
}

func (gv *GoLog) tickClock() {
	_, found := gv.currentVC.FindTicks(gv.pid)
	if !found {
		gv.logger.Println("Couldn't find this process's id in its own vector clock!")
	}
	gv.currentVC.Tick(gv.pid)
}

func (gv *GoLog) mergeIncomingClock(mesg string, e ClockPayload) {

	// First, tick the local clock
	gv.tickClock()
	gv.currentVC.Merge(e.VcMap)
	debugPrint("Now, Vector Clock is : ", gv.currentVC, gv)

	logWriteWrapper(mesg, "Something went Wrong, Could not Log!", gv.currentVC, gv)
}

func logWriteWrapper(logMessage, errorMessage string, vc vclock.VClock, gv *GoLog) (success bool) {
	if gv.logging == true {
		success = gv.logFunc(logMessage, gv.pid, vc.ReturnVCString())
		if !success {
			gv.logger.Println(errorMessage)
		}
	}
	return
}

func debugPrint(message string, vc vclock.VClock, gv *GoLog) {
	if gv.debugmode == true {
		gv.logger.Println(message)
		gv.logger.Print("VCLOCK IS :")
		vc.PrintVC()
	}
}

/*
UnpackReceive is used to unmarshall network data into local structures.
LogMessage will be logged along with the vector time the receive happened
buf is the network data, previously packaged by PrepareSend unpack is
a pointer to a structure, the same as was packed by PrepareSend

This function is meant to be called immediately after receiving
a packet. It unpacks the data by the program, the vector clock. It
updates vector clock and logs it. and returns the user data
*/
func (gv *GoLog) UnpackReceive(mesg string, buf []byte, unpack interface{}) {

	gv.mutex.Lock()

	e := ClockPayload{}
	e.Payload = unpack

	// Just use msgpack directly
	err := gv.decodingStrategy(buf, &e)
	if err != nil {
		gv.logger.Println(err.Error())
	}

	// Increment and merge the incoming clock
	gv.mergeIncomingClock(mesg, e)
	gv.mutex.Unlock()

}

func (gv *GoLog) EnableLogging() {
	gv.logging = true
}

//Disables Logging. Log messages will not appear in Log file any longer.
//Note: For the moment, the vector clocks are going to continue being updated.
func (gv *GoLog) DisableLogging() {
	gv.logging = false
}

func (gv *GoLog) SetLogFunc(logFunc func(message, pid, VCString string) bool) {
	gv.logFunc = logFunc
}

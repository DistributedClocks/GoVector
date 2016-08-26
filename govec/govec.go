package govec

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/arcaneiceman/GoVector/govec/vclock"
	"github.com/hashicorp/go-msgpack/codec"
	"runtime/pprof"
	"regexp"
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

const logToTerminal = false

//This is the Global Variable Struct that holds all the info needed to be maintained
type GoLog struct {

	//Local Process ID
	pid string

	//Local vector clock in bytes
	currentVC []byte

	//Flag to Printf the logs made by Local Program
	printonscreen bool

	//This bools Checks to send the VC Bundled with User Data on the Wire
	// if False, PrepareSend and UnpackReceive will simply forward their input
	//buffer to output and locally log event. If True, VC will be encoded into packet on wire
	VConWire bool

	// This bool checks whether we should send logging data to a vecbroker whenever
	// a message is logged.
	realtime bool

	//activates/deactivates printouts at each preparesend and unpackreceive
	debugmode bool

	logging bool

	//Logfile name
	logfile string

	// Publisher to enable sending messages to a vecbroker.
	publisher *GoPublisher

	//publicly supplied encoding function
	encodingStrategy func(interface{}) ([]byte, error)

	//publicly supplied decoding function
	decodingStrategy func([]byte, interface{}) error

	logger *log.Logger

	mutex sync.RWMutex
}

//This is the data structure that is actually end on the wire
type Data struct {
	pid         []byte
	vcinbytes   []byte
	programdata []byte
}

//Prints the Data Stuct as Bytes
func (d *Data) PrintDataBytes() {
	fmt.Printf("%x \n", d.pid)
	fmt.Printf("%X \n", d.vcinbytes)
	fmt.Printf("%X \n", d.programdata)
}

//Prints the Data Struct as a String
func (d *Data) PrintDataString() {
	fmt.Println("-----DATA START -----")
	s := string(d.pid[:])
	fmt.Println(s)
	s = string(d.vcinbytes[:])
	fmt.Println(s)
	fmt.Println("-----DATA END -----")
}

// TODO Pull request for this function
func (gv *GoLog) GetCurrentVC() []byte {
	return gv.currentVC
}

func (gv *GoLog) LogThis(Message string, ProcessID string, VCString string) bool {
	complete := true
	var buffer bytes.Buffer
	buffer.WriteString(ProcessID)
	buffer.WriteString(" ")
	buffer.WriteString(VCString)
	buffer.WriteString("\n")
	buffer.WriteString(Message)
	buffer.WriteString("\n")
	output := buffer.String()

	file, err := os.OpenFile(gv.logfile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		complete = false
	}
	defer file.Close()

	if _, err = file.WriteString(output); err != nil {
		complete = false
	}
	if gv.printonscreen == true {
		gv.logger.Println(output)
	}
	return complete

}

func (gv *GoLog) LogLocalEvent(Message string) bool {
	
	gv.mutex.Lock()
	//Converting Vector Clock from Bytes and Updating the gv clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if err != nil {
		gv.logger.Println(err.Error())
	}
	_ , found := vc.FindTicks(gv.pid)
	if found == false {
		gv.logger.Println("Couldnt find this process's id in its own vector clock!")
	}
	vc.Tick(gv.pid)
	gv.currentVC = vc.Bytes()

	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(Message, gv.pid, vc.ReturnVCString())
		if gv.realtime == true {
			// Send local message to broker
			//BUG go publisher never worked // gv.publisher.PublishLocalMessage(Message, gv.pid, *vc)
		}
	}
	gv.mutex.Unlock()

	return ok
}

//This function is meant to be used immidatly before sending.
//mesg will be loged along with the time of the send
//buf is encodeable data (structure or basic type)
//Returned is an encoded byte array with logging information
func (gv *GoLog) PrepareSend(mesg string, buf interface{}) []byte {
	/*
		This function is meant to be called before sending a packet. Usually,
		it should Update the Vector Clock for its own process, package with the clock
		using gob support and return the new byte array that should be sent onwards
		using the Send Command
	*/
	//Converting Vector Clock from Bytes and Updating the gv clock
	gv.mutex.Lock()
	vc, err := vclock.FromBytes(gv.currentVC)
	if err != nil {
		gv.logger.Println(err.Error())
	}
	_ , found := vc.FindTicks(gv.pid)
	if found == false {
		gv.logger.Println("Couldnt find this process's id in its own vector clock!")
	}
	vc.Tick(gv.pid)
	gv.currentVC = vc.Bytes()
	//WILL HAVE TO CHECK THIS OUT!!!!!!!

	if gv.debugmode == true {
		gv.logger.Println("Sending Message")
		gv.logger.Print("Current Vector Clock : ")
		vc.PrintVC()
	}

	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(mesg, gv.pid, vc.ReturnVCString())
	}

	if ok == false {
		gv.logger.Println("Something went Wrong, Could not Log!")
	}

	// Create New Data Structure and add information: data to be transfered
	d := Data{}
	d.pid = []byte(gv.pid)
	d.vcinbytes = gv.currentVC

	//first layer of encoding (user data)
	d.programdata, err = gv.encodingStrategy(buf)
	if err != nil {
		gv.logger.Println(err.Error())
	}

	//second layer wrapperEncoderoding, wrapperEncoderode wrapping structure
	wrapperBuffer := new(bytes.Buffer)
	wrapperEncoder := gob.NewEncoder(wrapperBuffer)
	err = wrapperEncoder.Encode(&d)
	if err != nil {
		gv.logger.Println(err.Error())
	}
	//return wrapperBuffer bytes which are wrapperEncoderoded as gob. Now these bytes can be sent off and
	// received on the other end!
	gv.mutex.Unlock()
	return wrapperBuffer.Bytes()
}

//UnpackReceive is used to unmarshall network data into local
//structures
//mesg will be logged along with the vector time the receive happened
//buf is the network data, previously packaged by PrepareSend
//unpack is a pointer to a structure, the same as was packed by
//PrepareSend
func (gv *GoLog) UnpackReceive(mesg string, buf []byte, unpack interface{}) {
	/*
		This function is meant to be called immediatly after receiving a packet. It unpacks the data
		by the program, the vector clock. It updates vector clock and logs it. and returns the user data
	*/

	gv.mutex.Lock()
	//if we are receiving a packet with a time stamp then:
	//Create a new buffer holder and decode into E, a new Data Struct
	buffer := new(bytes.Buffer)
	buffer = bytes.NewBuffer(buf)
	e := new(Data)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(e)
	if err != nil  && gv.debugmode {
		callingFunc := getCallingFunctionID()
		gv.logger.Printf("Vector clock decode failure. Likely one was not present on the incomming network payload. Bad payload received from %s, consider instrumenting the sender\n",callingFunc)
		gv.logger.Println(err.Error())
	}

	//In this case you increment your old clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if gv.debugmode {
		gv.logger.Println("Received :")
		e.PrintDataString()
		gv.logger.Print("Received Vector Clock : ")
		vc.PrintVC()
	}

	_ , found := vc.FindTicks(gv.pid)
	if !found {
		gv.logger.Println(fmt.Errorf("Couldnt find Local Process's ID in the Vector Clock. Could it be a stray message?"))
	}
	if err != nil {
		gv.logger.Println(err.Error())
	}
	vc.Tick(gv.pid)
	//merge it with the new clock and update GV
	tmp := []byte(e.vcinbytes[:])
	tempvc, err := vclock.FromBytes(tmp)

	if err != nil && gv.debugmode {
		gv.logger.Println(err.Error())
	}
	vc.Merge(tempvc)
	if gv.debugmode == true {
		gv.logger.Print("Now, Vector Clock is : ")
		vc.PrintVC()
	}
	gv.currentVC = vc.Bytes()

	//Log it
	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(mesg, gv.pid, vc.ReturnVCString())
	}
	if ok == false {
		gv.logger.Println("Something went Wrong, Could not Log!")
	}
	err = gv.decodingStrategy(e.programdata, unpack)
	if err != nil {
		gv.logger.Println(err.Error())
	}
	gv.mutex.Unlock()
}

//This function packs the Vector Clock with user program's data to send on wire
func (d *Data) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.pid)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.vcinbytes)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.programdata)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

//This function Unpacks packet containing the vector clock received from wire
func (d *Data) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.pid)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.vcinbytes)
	if err != nil {
		return err
	}
	return decoder.Decode(&d.programdata)
}

/*
This implementation of UnpackReceive returns the value of the unpacked
message. Due conflicting requirements for interfaces, it has been
removed
func (gv *GoLog) UnpackReceive(mesg string, buf []byte) interface{} {
	var unpackInto interface{}
	gv.unpack(mesg, buf, &unpackInto)
	return unpackInto
}
*/

//SetEncoderDecoder allows users to specify the encoder, and decoder
//used by GoVec in the case the default is unable to preform a
//required task
//encoder a function which takes an interface, and returns a byte
//array, and an error if encoding fails
//decoder a function which takes an encoded byte array and a pointer
//to a structure. The byte array should be decoded into the structure.
func (gv *GoLog) SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error) {
	gv.encodingStrategy = encoder
	gv.decodingStrategy = decoder
}

//gobDecodingStrategy decodes user data by using the Go Object decoder
func gobDecodingStrategy(programData []byte, unpack interface{}) error {
	//Decode the user defined message
	programDataBuffer := new(bytes.Buffer)
	programDataBuffer = bytes.NewBuffer(programData)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	msgdec := gob.NewDecoder(programDataBuffer)
	err := msgdec.Decode(unpack)
	if err != nil {
		err = fmt.Errorf("Unable to encode with gob encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
		return err
	}
	return nil
}

//gobEncodingStrategy encodes user data by using the Go Object encoder
func gobEncodingStrategy(buf interface{}) ([]byte, error) {
	//first layer encoding, encoding buffer argument
	programDataBuffer := new(bytes.Buffer)
	programDataEncoder := gob.NewEncoder(programDataBuffer)
	err := programDataEncoder.Encode(buf)
	if err != nil {
		err = fmt.Errorf("Unable to encode with gob encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
		return nil, err
	}
	programdata := programDataBuffer.Bytes()
	return programdata, nil
}

func msgPackDecodingStrategy(programData []byte, unpack interface{}) error {
	var (
		dec *codec.Decoder
	)
	dec = codec.NewDecoderBytes(programData, &codec.MsgpackHandle{})
	err := dec.Decode(unpack)
	if err != nil {
		err = fmt.Errorf("Unable to decode with msg-pack encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
		return err
	}
	return nil
}

func msgPackEncodingStrategy(buf interface{}) ([]byte, error) {
	var (
		b   []byte
		enc *codec.Encoder
	)
	enc = codec.NewEncoderBytes(&b, &codec.MsgpackHandle{})
	err := enc.Encode(buf)
	//fmt.Println(b)
	if err != nil {
		err = fmt.Errorf("Unable to encode with msg-pack encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
		return nil, err
	}
	return b, err
}

func (gv *GoLog) EnableLogging() {
	gv.logging = true
}

func (gv *GoLog) DisableLogging() {
	gv.logging = false
}

func New() *GoLog {
	return &GoLog{}
}

func Initialize(processid string, logfilename string) *GoLog {
	/*This is the Start Up Function That should be called right at the start of
	  a program
	*/
	gv := New() //Simply returns a new struct
	gv.pid = processid

	if logToTerminal {
		gv.logger = log.New(os.Stdout,"[GoVector]:",log.Lshortfile)
	} else {
		var buf bytes.Buffer
		gv.logger = log.New(&buf,"[GoVector]:",log.Lshortfile)
	}

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = false //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = false     // (Debug)
	gv.EnableLogging()

	/*
	//set the default encoder / decoder to gob
	gv.encodingStrategy = gobEncodingStrategy
	gv.decodingStrategy = gobDecodingStrategy
	*/

	//set the default encoder / decoder to msgpack
	gv.encodingStrategy = msgPackEncodingStrategy
	gv.decodingStrategy = msgPackDecodingStrategy

	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Tick(processid)

	//Vector Clock Stored in bytes
	//copy(gv.currentVC[:],vc1.Bytes())
	gv.currentVC = vc1.Bytes()

	if gv.debugmode == true {
		gv.logger.Println(" ###### Initialization ######")
		gv.logger.Print("VCLOCK IS :")
		vc1.PrintVC()
		gv.logger.Println(" ##### End of Initialization ")
	}

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
	//Log it
	ok := gv.LogThis("Initialization Complete", gv.pid, vc1.ReturnVCString())
	if ok == false {
		gv.logger.Println("Something went Wrong, Could not Log!")
	}

	return gv
}

func InitializeMutipleExecutions(processid string, logfilename string) *GoLog {
	/*This is the Start Up Function That should be called right at the start of
	  a program
	*/
	gv := New() //Simply returns a new struct
	gv.pid = processid

	if logToTerminal {
		gv.logger = log.New(os.Stdout,"[GoVector]:",log.Lshortfile)
	} else {
		var buf bytes.Buffer
		gv.logger = log.New(&buf,"[GoVector]:",log.Lshortfile)
	}

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = false //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = false     // (Debug)
	gv.EnableLogging()
	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Tick(processid)

	//Vector Clock Stored in bytes
	//copy(gv.currentVC[:],vc1.Bytes())
	gv.currentVC = vc1.Bytes()

	//set the default encoder / decoder to gob
	gv.encodingStrategy = gobEncodingStrategy
	gv.decodingStrategy = gobDecodingStrategy

	if gv.debugmode == true {
		gv.logger.Println(" ###### Initialization ######")
		gv.logger.Print("VCLOCK IS :")
		vc1.PrintVC()
		gv.logger.Println(" ##### End of Initialization ")
	}

	//Starting File IO . If Log exists, it will find Last execution number and ++ it
	logname := logfilename + "-Log.txt"
	_, err := os.Stat(logname)
	gv.logfile = logname
	if err == nil {
		//its exists... deleting old log
		gv.logger.Println(logname, " exists! ...  Looking for Last Exectution... ")
		executionnumber := FindExecutionNumber(logname)
		executionnumber = executionnumber + 1
		gv.logger.Println("Execution Number is  ", executionnumber)
		executionstring := "=== Execution #" + strconv.Itoa(executionnumber) + "  ==="
		gv.LogThis(executionstring, "", "")
	} else {
		//Creating new Log
		file, err := os.Create(logname)
		if err != nil {
			gv.logger.Println(err.Error())
		}
		file.Close()

		//Mark Execution Number
		ok := gv.LogThis("=== Execution #1 ===", " ", " ")
		//Log it
		ok = gv.LogThis("Initialization Complete", gv.pid, vc1.ReturnVCString())
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


//getCallingFunctionID returns the file name and line number of the
//program which called capture.go. This function is used to determine
//the calling function which did not receive a vector clock
func getCallingFunctionID() string {
	profiles := pprof.Profiles()
	block := profiles[1]
	var buf bytes.Buffer
	block.WriteTo(&buf, 1)
	//gv.logger.Printf("%s",buf)
	passedFrontOnStack := false
	re := regexp.MustCompile("([a-zA-Z0-9]+.go:[0-9]+)")
	ownFilename := regexp.MustCompile("capture.go") // hardcoded own filename
	matches := re.FindAllString(fmt.Sprintf("%s", buf), -1)
	for _, match := range matches {
		if passedFrontOnStack && !ownFilename.MatchString(match) {
			return match
		} else if ownFilename.MatchString(match) {
			passedFrontOnStack = true
		}
		//fmt.Printf("found %s\n", match)
	}
	fmt.Printf("%s\n", buf)
	return ""
}

package govec

import "fmt"
import "encoding/gob"
import "bytes"
import (
	"github.com/arcaneiceman/GoVector/govec/vclock"
	"github.com/hashicorp/go-msgpack/codec"
)
import "os"
import "strings"
import "strconv"
import "io"
import "bufio"

/*
	- All licneces like other licenses ...

	How to Use This Library

	Step 1:
	Create a Global Variable and Initialize it using like this =

	Logger:= Initialize("MyProcess",ShouldYouSeeLoggingOnScreen,ShouldISendVectorClockonWire,Debug)

	Step 2:
	When Ever You Decide to Send any []byte , before sending call PrepareSend like this:
	SENDSLICE := PrepareSend("Message Description", []YourPayload)
	and send the SENDSLICE instead of your Designated Payload

	Step 3:
	When Receiveing, AFTER you receive your message, pass the []byte into UnpackRecieve
	like this:

	RETURNSLICE := UnpackReceive("Message Description" []ReceivedPayload)
	and use RETURNSLICE for further processing.


*/
/*
//publicly supplied encoding function
func encode(interface{}) ([]byte, error) {}

//publicly supplied decoding function
func decode([]byte) (interface{}, error) {}
*/
//This is the Global Variable Struct that holds all the info needed to be maintained
type GoLog struct {

	//Local Process ID
	pid string

	//Local vector clock in bytes
	currentVC []byte

	//Flag to Printf the logs made by Local Program
	printonscreen bool

	//activates/deactivates printouts at each preparesend and unpackreceive
	debugmode bool

	logging bool

	//Logfilename
	logfile string

	//publicly supplied encoding function
	encodingStrategy func(interface{}) ([]byte, error)

	//publicly supplied decoding function
	decodingStrategy func([]byte, interface{}) error
}

//This is the data structure that is actually end on the wire
type Data struct {
	pid         []byte
	vcinbytes   []byte
	programdata []byte
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
		fmt.Println(output)
	}
	return complete

}

func (gv *GoLog) LogLocalEvent(Message string) bool {

	//Converting Vector Clock from Bytes and Updating the gv clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if err != nil {
		panic(err)
	}
	currenttime, found := vc.FindTicks(gv.pid)
	if found == false {
		panic("Couldnt find this process's id in its own vector clock!")
	}
	currenttime++
	vc.Update(gv.pid, currenttime)
	gv.currentVC = vc.Bytes()

	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(Message, gv.pid, vc.ReturnVCString())
	}

	return ok
}

func (gv *GoLog) PrepareSend(mesg string, buf interface{}) []byte {
	/*
		This function is meant to be called before sending a packet. Usually,
		it should Update the Vector Clock for its own process, package with the clock
		using gob support and return the new byte array that should be sent onwards
		using the Send Command
	*/
	//Converting Vector Clock from Bytes and Updating the gv clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if err != nil {
		panic(err)
	}
	currenttime, found := vc.FindTicks(gv.pid)
	if found == false {
		panic("Couldnt find this process's id in its own vector clock!")
	}
	currenttime++
	vc.Update(gv.pid, currenttime)
	gv.currentVC = vc.Bytes()
	//WILL HAVE TO CHECK THIS OUT!!!!!!!

	if gv.debugmode == true {
		fmt.Println("Sending Message")
		fmt.Print("Current Vector Clock : ")
		vc.PrintVC()
	}

	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(mesg, gv.pid, vc.ReturnVCString())
	}

	if ok == false {
		fmt.Println("Something went Wrong, Could not Log!")
	}

	// Create New Data Structure and add information: data to be transfered
	d := Data{}
	d.pid = []byte(gv.pid)
	d.vcinbytes = gv.currentVC

	//first layer of encoding (user data)
	d.programdata, err = gv.encodingStrategy(buf)
	if err != nil {
		panic(err)
	}

	//second layer wrapperEncoderoding, wrapperEncoderode wrapping structure
	wrapperBuffer := new(bytes.Buffer)
	wrapperEncoder := gob.NewEncoder(wrapperBuffer)
	err = wrapperEncoder.Encode(&d)
	if err != nil {
		panic(err)
	}
	//return wrapperBuffer bytes which are wrapperEncoderoded as gob. Now these bytes can be sent off and
	// received on the other end!
	return wrapperBuffer.Bytes()
}

func (gv *GoLog) unpack(mesg string, buf []byte, unpack interface{}) {
	/*
		This function is meant to be called immediatly after receiving a packet. It unpacks the data
		by the program, the vector clock. It updates vector clock and logs it. and returns the user data
	*/

	//if we are receiving a packet with a time stamp then:
	//Create a new buffer holder and decode into E, a new Data Struct
	buffer := new(bytes.Buffer)
	buffer = bytes.NewBuffer(buf)
	e := new(Data)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(e)
	if err != nil {
		fmt.Println("You said that I would be receiving a vector clock but I didnt! or decoding failed :P")
		panic(err)
	}

	//In this case you increment your old clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if gv.debugmode == true {
		fmt.Println("Received :")
		e.PrintDataString()
		fmt.Print("Received Vector Clock : ")
		vc.PrintVC()
	}

	currenttime, found := vc.FindTicks(gv.pid)
	if found == false {
		panic(fmt.Errorf("Couldnt find Local Process's ID in the Vector Clock. Could it be a stray message?"))
	}
	if err != nil {
		panic(err)
	}
	currenttime++
	vc.Update(gv.pid, currenttime)
	//merge it with the new clock and update GV
	tmp := []byte(e.vcinbytes[:])
	tempvc, err := vclock.FromBytes(tmp)

	if err != nil {
		panic(err)
	}
	vc.Merge(tempvc)
	if gv.debugmode == true {
		fmt.Print("Now, Vector Clock is : ")
		vc.PrintVC()
	}
	gv.currentVC = vc.Bytes()

	//Log it
	var ok bool
	if gv.logging == true {
		ok = gv.LogThis(mesg, gv.pid, vc.ReturnVCString())
	}
	if ok == false {
		fmt.Println("Something went Wrong, Could not Log!")
	}
	err = gv.decodingStrategy(e.programdata, unpack)
	if err != nil {
		panic(err)
	}
}

func (gv *GoLog) UnpacReceiveInto(mesg string, buf []byte, unpack interface{}) {
	gv.unpack(mesg, buf, unpack)
}

func (gv *GoLog) UnpackReceive(mesg string, buf []byte) interface{} {
	var unpackInto interface{}
	gv.unpack(mesg, buf, unpackInto)
	return unpackInto
}

//gobDecodingStrategy decodes user data by using the Go Object decoder
func gobDecodingStrategy(programData []byte, unpack interface{}) error {
	//Decode the user defined message
	programDataBuffer := new(bytes.Buffer)
	programDataBuffer = bytes.NewBuffer(programData)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	msgdec := gob.NewDecoder(programDataBuffer)
	err := msgdec.Decode(&unpack)
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
	err := programDataEncoder.Encode(&buf)
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
	err := dec.Decode(&unpack)
	if err != nil {
		err = fmt.Errorf("Unable to encode with gob encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
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
	fmt.Println(b)
	if err != nil {
		err = fmt.Errorf("Unable to encode with gob encoder, consider using a different type or custom encoder/decoder : or : %s", err.Error())
		return nil, err
	}
	return b, err
}

func (gv *GoLog) SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error) {
	gv.encodingStrategy = encoder
	gv.decodingStrategy = decoder
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

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = false //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = false     // (Debug)
	gv.EnableLogging()

	//set the default encoder / decoder to gob
	/*
		gv.encodingStrategy = gobEncodingStrategy
		gv.decodingStrategy = gobDecodingStrategy
	*/

	//set the default encoder / decoder to msgpack
	gv.encodingStrategy = msgPackEncodingStrategy
	gv.decodingStrategy = msgPackDecodingStrategy

	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Update(processid, 1)

	//Vector Clock Stored in bytes
	//copy(gv.currentVC[:],vc1.Bytes())
	gv.currentVC = vc1.Bytes()

	if gv.debugmode == true {
		fmt.Println(" ###### Initialization ######")
		fmt.Print("VCLOCK IS :")
		vc1.PrintVC()
		fmt.Println(" ##### End of Initialization ")
	}

	//Starting File IO . If Log exists, Log Will be deleted and A New one will be created
	logname := logfilename + "-Log.txt"

	if _, err := os.Stat(logname); err == nil {
		//its exists... deleting old log
		fmt.Println(logname, "exists! ... Deleting ")
		os.Remove(logname)
	}
	//Creating new Log
	file, err := os.Create(logname)
	if err != nil {
		panic(err)
	}
	file.Close()

	gv.logfile = logname
	//Log it
	ok := gv.LogThis("Initialization Complete", gv.pid, vc1.ReturnVCString())
	if ok == false {
		fmt.Println("Something went Wrong, Could not Log!")
	}

	return gv
}

func InitializeMutipleExecutions(processid string, logfilename string) *GoLog {
	/*This is the Start Up Function That should be called right at the start of
	  a program
	*/
	gv := New() //Simply returns a new struct
	gv.pid = processid

	//# These are bools that can be changed to change debuging nature of library
	gv.printonscreen = false //(ShouldYouSeeLoggingOnScreen)
	gv.debugmode = false     // (Debug)
	gv.EnableLogging()
	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Update(processid, 1)

	//Vector Clock Stored in bytes
	//copy(gv.currentVC[:],vc1.Bytes())
	gv.currentVC = vc1.Bytes()

	if gv.debugmode == true {
		fmt.Println(" ###### Initialization ######")
		fmt.Print("VCLOCK IS :")
		vc1.PrintVC()
		fmt.Println(" ##### End of Initialization ")
	}

	//Starting File IO . If Log exists, it will find Last execution number and ++ it
	logname := logfilename + "-Log.txt"
	_, err := os.Stat(logname)
	gv.logfile = logname
	if err == nil {
		//its exists... deleting old log
		fmt.Println(logname, " exists! ...  Looking for Last Exectution... ")
		executionnumber := FindExecutionNumber(logname)
		executionnumber = executionnumber + 1
		fmt.Println("Execution Number is  ", executionnumber)
		executionstring := "=== Execution #" + strconv.Itoa(executionnumber) + "  ==="
		gv.LogThis(executionstring, "", "")
	} else {
		//Creating new Log
		file, err := os.Create(logname)
		if err != nil {
			panic(err)
		}
		file.Close()

		//Mark Execution Number
		ok := gv.LogThis("=== Execution #1 ===", " ", " ")
		//Log it
		ok = gv.LogThis("Initialization Complete", gv.pid, vc1.ReturnVCString())
		if ok == false {
			fmt.Println("Something went Wrong, Could not Log!")
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

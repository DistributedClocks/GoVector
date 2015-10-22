package govec

import "fmt"
import "encoding/gob"
import "bytes"
import "github.com/arcaneiceman/GoVector/govec/vclock"
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

	//activates/deactivates printouts at each preparesend and unpackreceive
	debugmode bool

	logging bool

	//Logfile name
	logfile string
	
	// Connect to logging server?
	realtime bool
	
	// IP address of server to connect to
	serveraddr string
	
	serverconn PublishConn

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
	s = string(d.programdata[:])
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

	if gv.realtime == true {
		serverconn.Send(output)
	}	

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

func (gv *GoLog) PrepareSend(mesg string, buf []byte) []byte {
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

	//if only local logging the next is unnecceary since we can simply return buf as is
	if gv.VConWire == true {
		//if true, then we add relevant info and encode it
		// Create New Data Structure and add information: data to be transfered
		d := Data{}
		d.pid = []byte(gv.pid)
		d.vcinbytes = gv.currentVC
		d.programdata = buf

		//create a buffer to hold data and Encode it
		buffer := new(bytes.Buffer)
		enc := gob.NewEncoder(buffer)
		err := enc.Encode(&d)
		if err != nil {
			panic(err)

		}
		//return buffer bytes which are encoded as gob. Now these bytes can be sent off and
		// received on the other end!
		return buffer.Bytes()
	}
	// if we are only performing local logging, we have updated vector clock and logged it buffer can
	// be returned as is
	return buf
}

func (gv *GoLog) UnpackReceive(mesg string, buf []byte) []byte {
	/*
		This function is meant to be called immediatly after receiving a packet. It unpacks the data
		by the program, the vector clock. It updates vector clock and logs it. and returns the user data
	*/
	//if only our program is logging there is nothing attached in the byte buffer
	if gv.VConWire == false {
		//simple adding to current time
		vc, err := vclock.FromBytes(gv.currentVC)
		if err != nil {
			panic(err)
		}
		currenttime, found := vc.FindTicks(gv.pid)
		if found == false {
			panic("Couldnt find own process in local clock!")
		}
		currenttime++
		vc.Update(gv.pid, currenttime)
		gv.currentVC = vc.Bytes()

		if gv.debugmode == true {
			fmt.Println("A Message was Recieved")
			fmt.Print("New Clock : ")
			vc.PrintVC()
		}

		//logging local
		var ok1 bool
		if gv.logging == true {
			ok1 = gv.LogThis(mesg, gv.pid, vc.ReturnVCString())
		}

		if ok1 == false {
			fmt.Println("Something went Wrong, Could not Log!")
		}

		return buf
	}

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
		fmt.Println("Couldnt find Local Process's ID in the Vector Clock. Could it be a stray message?")
		return nil
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

	//  Out put the recieved Data
	tmp2 := []byte(e.programdata[:])
	return tmp2
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
	gv.VConWire = true       // (ShouldISendVectorClockonWire)
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
	gv.VConWire = true       // (ShouldISendVectorClockonWire)
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

type PublishConn struct {
	conn      *websocket.Conn
	address   string
}

func (pc *PublishConn) Init(string address) {
	pc.address = address
	// net.conn via dial, needs more work
}

package main

import "fmt"
import "encoding/gob"
import "bytes"
import "./vclock"
import "strings"

/*
	- All licneces like other licenses ...
	
	How to Use This Library (After its Complete)
	
	Step 1:
	Create a Global Variable and Initilize it using like this = 
	
	Logger:= Initialize("MyProcess","ProcessId",ShouldYouSeeLoggingOnScreen,ShouldISendVectorClockonWire,Debug)
	
	Step 2:
	When Ever You Decide to Send any []byte , before sending call PrepareSend like this:
	SENDSLICE := PrepareSend([]YourPayload)
	and send the SENDSLICE instead of your Designated Payload
	
	Step 3:
	When Receiveing, AFTER you receive your message, pass the []byte into UnpackRecieve
	like this:
	
	RETURNSLICE := UnpackReceive([]ReceivedPayload)
	and use RETURNSLICE for further processing.
	
	Restrictions
	
*/


//This is the Global Variable Struct that holds all the info needed to be maintained 
type GoLog struct {
	// Local Process Name 
	processname	string
	
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
}

//This is the data structure that is actually end on the wire
type Data struct {
    name [200]byte
	pid [4]byte
	vcinbytes [200] byte
	programdata [1024]byte
}

//This function packs the Vector Clock with user program's data to send on wire
func (d *Data) GobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err:= encoder.Encode(d.name)
    if (err!=nil) {
        return nil, err
    }
	err = encoder.Encode(d.pid)
	if (err!=nil) {
        return nil, err
    }
	err = encoder.Encode(d.vcinbytes)
	if (err!=nil) {
		return nil,err
	}
	err = encoder.Encode(d.programdata)
	if (err!=nil) {
		return nil,err
	}
    return w.Bytes(), nil
}

//This function Unpacks packet containing the vector clock received from wire
func (d *Data) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
	err:= decoder.Decode(&d.name)
	if (err!=nil) {
        return err
	}
	err = decoder.Decode(&d.pid)
	if (err!=nil) {
        return err
    }
	err = decoder.Decode(&d.vcinbytes)
	if (err!=nil) {
        return err
    }
    return decoder.Decode(&d.programdata)
}

//Prints the Data Stuct as Bytes
func (d *Data) PrintDataBytes() {
	fmt.Printf("%x \n",d.name)
	fmt.Printf("%x \n",d.pid)
	fmt.Printf("%X \n",d.vcinbytes)
	fmt.Printf("%X \n",d.programdata)
}

//Prints the Data Struct as a String
func (d *Data) PrintDataString() {
	fmt.Println("-----DATA START -----")
	s:= string(d.name[:])
	fmt.Println(strings.TrimSpace(s))
	s= string(d.pid[:])
	fmt.Println(s)
	s= string(d.vcinbytes[:])
	fmt.Println(s)
	s= string(d.programdata[:])
	fmt.Println(strings.Trim(s," "))
	fmt.Println("-----DATA END -----")
}


func (gv *GoLog) PrepareSend(buf []byte) ([]byte){
/*
	This function is meant to be called before sending a packet. Usually,
	it should Update the Vector Clock for its own process, package with the clock
	using gob support and return the new byte array that should be sent onwards
	using the Send Command
*/
	//Converting Vector Clock from Bytes and Updating the gv clock
	vc, err := vclock.FromBytes(gv.currentVC)
	if (err!= nil) {
			panic(err)
		}
	currenttime , found := vc.FindTicks(gv.pid)
	if (found ==false){
        panic("Couldnt find this process's id in its own vector clock!")
	}
	currenttime++
	vc.Update(gv.pid,currenttime)
	gv.currentVC=vc.Bytes()
	//WILL HAVE TO CHECK THIS OUT!!!!!!! 
	
	if (gv.debugmode == true){
		fmt.Println("Sending Message")
		fmt.Print("Current Vector Clock : ")
		vc.PrintVC()
	}
	
	//lets log the event
	//print
	
	//if only local logging the next is unnecceary since we can simply return buf as is 
	if (gv.VConWire == true) {
	//if true, then we add relevant info and encode it
		// Create New Data Structure and add information: data to be transfered
		d := Data{} 
		copy(d.name[:], []byte(gv.processname))
		copy(d.pid[:], []byte(gv.pid))
		copy(d.vcinbytes[:],gv.currentVC)
		copy(d.programdata[:], buf)
		
		//create a buffer to hold data and Encode it
		buffer := new(bytes.Buffer)
		enc := gob.NewEncoder(buffer)
        err := enc.Encode(d)
		if (err!=nil) {
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

func (gv *GoLog) UnpackReceive(buf []byte) ([] byte){ 
/*
	This function is meant to be called immediatly after receiving a packet. It unpacks the data 
	by the program, the vector clock. It updates vector clock and logs it. and returns the user data
*/
	//if only our program is logging there is nothing attached in the byte buffer
	if (gv.VConWire == false) {
		//simple adding to current time
		vc , err :=vclock.FromBytes(gv.currentVC)
		if err!= nil {
			panic(err)
		}
		currenttime , found := vc.FindTicks(gv.pid)
		if (found ==false){
		panic("Couldnt find own process in local clock!")
	}
	    currenttime++
	    vc.Update(gv.pid,currenttime)
		gv.currentVC=vc.Bytes()
		
		if (gv.debugmode == true){
		fmt.Println("A Message was Recieved")
		fmt.Print("New Clock : ")
		vc.PrintVC()
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
	if err!= nil {
		fmt.Println("You said that I would be receiving a vector clock but I didnt! or decoding failed :P")
		panic(err)
	}
	
	//In this case you increment your old clock
	vc , err :=vclock.FromBytes(gv.currentVC)
	if (gv.debugmode == true){
		fmt.Println("Received :")
		e.PrintDataString()
		fmt.Print("Received Vector Clock : ")
		vc.PrintVC()
	}
	
	currenttime , found := vc.FindTicks(gv.pid)
	if (found ==false){
		fmt.Println("Couldnt find Local Process's ID in the Vector Clock. Could it be a stray message?")
        return nil
	}
	if err!= nil {
			panic(err)
		}
	currenttime++
	vc.Update(gv.pid,currenttime)
	//merge it with the new clock
	
	
	tmp := []byte(e.vcinbytes[:])
	tempvc , err := vclock.FromBytes(tmp)
	
	if err!= nil {
			panic(err)
		}
	vc.Merge(tempvc)
	fmt.Print("Now, Vector Clock is : ")
	vc.PrintVC()
	
	//Log it
	
    //  Out put the recieved Data
	tmp2 := []byte(e.programdata[:])
	return tmp2
}

func New() *GoLog {
	return &GoLog{}
}

func Initilize(nameofprocess string, processid string, printouts bool , vectorclockonwire bool , debugmode bool) (*GoLog){
/*This is the Start Up Function That should be called right at the start of 
a program

Assumption that Code Creates Logger Struct using :
Logger := GoVec.New()
Logger.Initialize(nameofprocess, printlogline, locallogging)
*/
	gv := New() //Simply returns a new struct
	gv.processname = nameofprocess
	gv.pid = processid
	gv.printonscreen = printouts
	gv.VConWire = vectorclockonwire
	gv.debugmode= debugmode
	
	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vclock.New()
	vc1.Update(processid, 1)
	
	//Vector Clock Stored in bytes
	//copy(gv.currentVC[:],vc1.Bytes())
	gv.currentVC=vc1.Bytes()
	
	if (gv.debugmode == true){
		fmt.Println(" ###### Initilization ######")
		fmt.Print("VCLOCK IS :")
		vc1.PrintVC()
	    fmt.Println(" ##### End of Initilization ")
	}
	return gv
}


func main() {

	
	Logger:= Initilize("waliprocess", "0001", true, true, true)
	
	sendbuf := []byte("messagepayload")
	finalsend := Logger.PrepareSend(sendbuf)
	//send message
	
	//s:= string(finalsend[:])
	//fmt.Println(s)
	fmt.Println("End of Message")
		
		
	//receive message
	recbuf:= Logger.UnpackReceive(finalsend)
	//Logger.UnpackReceive(finalsend)
	//s:= string(recbuf[:])
	//fmt.Println(s)
    finalsend = Logger.PrepareSend(recbuf)
	
	
	
}
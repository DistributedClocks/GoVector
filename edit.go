package main

import "fmt"
import "encoding/gob"
import "bytes"
import "labix.org/v1/vclock"

//This is the Global Variable Struct that holds precious info
type GoVec struct {
	//processname 
	processname	string
	//vector clock in bytes
	currrentvc [32]byte
	//should I print the log on screen every time I store it?
	printonscreen bool
	//should I assume that only I am logging (or are other remote hosts
	//doing so as well
	locallogging bool
	//current time for clock
	currenttime uint64
}

type Data struct {
    id int32
    name [16]byte
	vcinbytes [32]byte
	programdata [32]byte
}


func (d *Data) GobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(d.id)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(d.name)
    if err!=nil {
        return nil, err
    }
	err = encoder.Encode(d.vcinbytes)
	if err!=nil {
		return nil,err
	}
	err = encoder.Encode(d.programdata)
	if err!=nil {
		return nil,err
	}
    return w.Bytes(), nil
}

func (d *Data) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&d.id)
    if err!=nil {
        return err
    }
	decoder.Decode(&d.name)
	decoder.Decode(&d.vcinbytes)
    return decoder.Decode(&d.programdata)
}

func (d *Data) PrintDataBytes() {
	fmt.Println(d.id)
	fmt.Printf("%x \n",d.name)
	fmt.Printf("%X \n",d.vcinbytes)
}

func (gv *GoVec) PrepareSend(buf []byte) ([]byte){
/*
	This function is meant to be called before sending a packet. Usually,
	it should Update the Vector Clock for its own process, package with the clock
	using gob support and return the new byte array that should be sent onwards
	using the Send Command
*/
	//Converting Vector Clock from Bytes and Updating 
	vc:=FromBytes(gv.currentVC)
	gv.currenttime++
	vc.Update(gv.processname,gv.currenttime)
	//WILL HAVE TO CHECK THIS OUT!!!!!!! 
	
	//lets log the event
	//print
	
	//if only local logging the next is unnecceary since we can simply return buf as is 
	if gv.locallogging == true {
	//if true, then we add relevant info and encode it
		// Create New Data Structure and add information: data to be transfered
		d := Data{id:15} 
		copy(d.name[:], []byte(gv.processname))
		copy(d.vcinbytes[:], vc.Bytes())
		copy(d.programdata[:], buf)
		
		//create a buffer to hold data and Encode it
		buffer := new(bytes.Buffer)
		enc := gob.NewEncoder(buffer)
        err := enc.Encode(d)
		if err!=nil {
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

//func UnpackRecieve(buf []byte) ([] byte){
//  Create a new Data Struct 
//  Decode Relevant Data //if fail it means that this doesnt not hold vector clock
//  Merge new Vector Clock with Old one
//  Log Event
//  Out put the recieved Data
//}

func (gv *GoVec) StartUp(n string, p bool , l bool){
/*This is the Start Up Function That should be called right at the start of 
a program

Assumption that Code Creates Logger Struct using :
LogVarName := GoVec.StartUp(nameofprocess, printlogline)
*/
	gv.processname = n
	gv.printonscreen = p
	gv.locallogging = l
	gv.currenttime = 0
	
	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vlclock.New()
	vc1.Update(n, gv.currenttime)
	
	//Vector Clock Stored in bytes
	gv.currentvc=vc1.Bytes()

}


func main() {	
    vc1 := vclock.New()
	vc1.Update("A", 1)
    d := Data{id: 7}
    copy(d.name[:], []byte("tree"))
	copy(d.vcinbytes[:],vc1.Bytes())
	
	d.PrintDataBytes()
    buffer := new(bytes.Buffer)
    // writing
    enc := gob.NewEncoder(buffer)
    err := enc.Encode(d)
    if err != nil {
        fmt.Println("error")
    }
    // reading
    buffer = bytes.NewBuffer(buffer.Bytes())
    e := new(Data)
    dec := gob.NewDecoder(buffer)
    err = dec.Decode(e)
    //fmt.Println(e, err)
	e.PrintDataBytes()
}
package main

import "fmt"
import "encoding/gob"
import "bytes"
import "labix.org/v1/vclock"

//This is the Global Variable Struct that holds precious info
type GoVec struct {
	processname	string
	cvc [16]byte
	printonscreen bool
}

type Data struct {
    id int32
    name [16]byte
	vcinbytes [32]byte
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
    return decoder.Decode(&d.vcinbytes)
}

func (d *Data) PrintDataBytes() {
	fmt.Println(d.id)
	fmt.Printf("%x \n",d.name)
	fmt.Printf("%X \n",d.vcinbytes)
}

//func PrepareSend(buf []byte) ([]byte){
// Here a new "Data" Stut will be formed
// Copy in PID and the rest
// Encode Data to Gob
// Update VC and Log Event
// Output the Gob of data to be sent
//}

//func UnpackRecieve(buf []byte) ([] byte){
//  Create a new Data Struct 
//  Decode Relevant Data 
//  Merge new Vector Clock with Old one
//  Log Event
//  Out put the recieved Data
//}

func StartUp(n string, p bool ){
//This is the Start Up Function That should be called right at the start of 
// a program

//Assumption that Code Creates Logger Struct using  LogVarName := GoVec.New
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
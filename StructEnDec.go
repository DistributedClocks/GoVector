package main

import "fmt"
import "encoding/gob"
import "bytes"

type Data struct {
    id int32
    name [16]byte
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
    return w.Bytes(), nil
}

func (d *Data) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&d.id)
    if err!=nil {
        return err
    }
    return decoder.Decode(&d.name)
}

func (d *Data) PrintData() {
	fmt.Println(d.id)
	fmt.Printf("%x \n",d.name)
}


func main() {
    d := Data{id: 7}
    copy(d.name[:], []byte("tree"))
	d.PrintData()
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
	e.PrintData()
}
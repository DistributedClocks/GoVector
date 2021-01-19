package main

import (
	"fmt"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/vmihailenco/msgpack/v5"
)

//ClockPayload is the wire type for vector clocks, and their
//associated payloads
type ClockPayload struct {
	Pid     string
	VcMap   map[string]uint64
	Payload interface{}
}

var _ msgpack.CustomEncoder = (*ClockPayload)(nil)
var _ msgpack.CustomDecoder = (*ClockPayload)(nil)

//EncodeMsgpack is a custom encoder function, needed for msgpack interoperability
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

//DecodeMsgpack is a custom decoder function, needed for msgpack
//interoperability
func (d *ClockPayload) DecodeMsgpack(dec *msgpack.Decoder) error {

	var err error
	err = dec.DecodeMulti(&d.Pid, &d.Payload, &d.VcMap)

	if err != nil {
		return err
	}

	return nil
}

func main() {

	fmt.Println("Hello World")

	/*	gv := govec.InitGoVector("myPID", "log")
		buf := gv.PrepareSend("Message", 137)

		var reply int
		gv.UnpackReceive("message", buf, &reply)
		fmt.Println(reply)*/

	clock := vclock.New()
	clock.Tick("abc")

	cp := ClockPayload{}
	cp.Pid = "pid"
	cp.VcMap = clock
	cp.Payload = 14

	b, err := msgpack.Marshal(&cp)
	if err != nil {
		panic(err)
	}

	var v ClockPayload
	err = msgpack.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}

	fmt.Println(v.Pid)
	fmt.Println(v.VcMap)
	fmt.Println(v.Payload)

}

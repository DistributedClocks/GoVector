package main

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"os"

	logging "github.com/op/go-logging"
	"github.com/ugorji/go/codec"
)

const (
	BLOOMSIZE    = 1 << 13 // 13 bytes 2^16 bits
	INT16BYTES   = 2
	INTBYTES     = 4
	BITSINBITE   = 8
	BLOOMINDEXES = 3

	LOGFORMAT = `%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`
)

type ClockLogger interface {
	//This function is meant to be used immidatly before sending.
	//mesg will be loged along with the time of the send
	//buf is encodeable data (structure or basic type)
	//Returned is an encoded byte array with logging information
	PrepareSend(mesg string, buf interface{}) []byte
	//UnpackReceive is used to unmarshall network data into local
	//structures
	//mesg will be logged along with the vector time the receive happened
	//buf is the network data, previously packaged by PrepareSend
	//unpack is a pointer to a structure, the same as was packed by
	//PrepareSend
	UnpackReceive(mesg string, buf []byte, unpack interface{})
	//SetEncoderDecoder allows users to specify the encoder, and decoder
	//used by GoVec in the case the default is unable to preform a
	//required task
	//encoder a function which takes an interface, and returns a byte
	//array, and an error if encoding fails
	//decoder a function which takes an encoded byte array and a pointer
	//to a structure. The byte array should be decoded into the structure.
	SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error)
	//Initalize a new clock logger with the given id and log file name
	Initialize(processid string, logfilename string) *ClockLogger
	//GetClock returns the clock being used to track time in the
	//system
	GetClock() []byte
}

type Info struct {
	BClock  BloomClock
	Hosts   VectorClock
	Lamport int
	Id      string
	Payload []byte
}

type BloomLog struct {
	BClock  *BloomClock
	Hosts   VectorClock
	Id      string
	Lamport int

	//Logfile name
	logfile string

	//publicly supplied encoding function
	encodingStrategy func(interface{}) ([]byte, error)

	//publicly supplied decoding function
	decodingStrategy func([]byte, interface{}) error

	logger *logging.Logger

	//instrumentation information
	addinfo     bool
	sendbclocks bool
	sendvclocks bool
	sendlamport bool
	sendid      bool
}

func Initalize(id string) *BloomLog {
	b := &BloomLog{}
	b.Id = id
	b.Lamport = 1
	b.BClock = NewBloomClock()
	b.Hosts = NewVectorClock(id)

	b.logger = setupLogger()
	b.encodingStrategy = msgPackEncodingStrategy
	b.decodingStrategy = msgPackDecodingStrategy

	//instrumentation information
	b.addinfo = true
	b.sendbclocks = true
	b.sendvclocks = true
	b.sendlamport = true
	b.sendid = true
	return b
}

func (b *BloomLog) PrepareSend(msg string, buf interface{}) []byte {
	//update clocks
	b.Lamport++
	b.Hosts[b.Id] = b.Lamport
	b.BClock.Update(b.Id, b.Lamport)
	//TODO log clocks to local file
	info := Info{}
	if b.addinfo {
		switch {
		case b.sendbclocks:
			info.BClock = *b.BClock
			fallthrough
		case b.sendvclocks:
			info.Hosts = b.Hosts
			fallthrough
		case b.sendlamport:
			info.Lamport = b.Lamport
			fallthrough
		case b.sendid:
			info.Id = b.Id
		}
	}
	var err error
	//first layer of encoding (user data)
	info.Payload, err = b.encodingStrategy(buf)
	if err != nil {
		b.logger.Error(err.Error())
	}

	wrapperBuffer := new(bytes.Buffer)
	wrapperEncoder := gob.NewEncoder(wrapperBuffer)
	err = wrapperEncoder.Encode(&info)
	if err != nil {
		b.logger.Error(err.Error())
	}
	return wrapperBuffer.Bytes()
}

func (b *BloomLog) UnpackReceive(mesg string, buf []byte, unpack interface{}) {
	buffer := new(bytes.Buffer)
	buffer = bytes.NewBuffer(buf)
	i := new(Info)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(i)

	if err != nil {
		b.logger.Fatalf("Unable to Info Payload: %s", err.Error())
	}

	b.Lamport++
	//This is pointless but why not
	//TODO actually implemnt vector clocks if you want to do this
	if i.Lamport > 0 && i.Id != "" {
		b.Hosts[i.Id] = i.Lamport
	}
	//merge BloomClock
	b.BClock.Merge(&i.BClock)
	//unpack user payload
	err = b.decodingStrategy(i.Payload, unpack)
	if err != nil {
		b.logger.Critical("Unable to unpack User payload check encoder and decoder configurations: %s", err.Error())
	}
}

func (b *BloomLog) SetEncoderDecoder(encoder func(interface{}) ([]byte, error), decoder func([]byte, interface{}) error) {
	b.encodingStrategy = encoder
	b.decodingStrategy = decoder
}

//TODO test everythinh

type BloomClock [BLOOMSIZE]byte

func NewBloomClock() *BloomClock {
	return &BloomClock{}
}

type VectorClock map[string]int

func NewVectorClock(id string) map[string]int {
	m := make(map[string]int, 0)
	m[id] = 1
	return m
}

func (b *BloomClock) Merge(other *BloomClock) {
	for i := range b {
		b[i] |= other[i]
	}
}

//returns true on chache it, false on miss
func (b *BloomClock) Check(id string, lamport int) bool {
	checkHash := getHash(id, lamport)
	indicies := getIndicies(checkHash)
	return b.checkIndicies(indicies)
}

func (b *BloomClock) Update(id string, lamport int) {
	updateHash := getHash(id, lamport)
	indicies := getIndicies(updateHash)
	b.addIndicies(indicies)

	//get indicies
}

func (b *BloomClock) addIndicies(indicies [BLOOMINDEXES]int) {
	var index, offset int
	//fmt.Printf("A: %d, B:%d, C:%d\n", indicies[0], indicies[1], indicies[2])
	for i := 0; i < BLOOMINDEXES; i++ {
		index = indicies[i] / BITSINBITE
		offset = indicies[i] % BITSINBITE
		//fmt.Printf("index %d, offset %d, i %d\n", index, offset, i)
		b[index] |= (1 << uint(offset))
	}
}

func (b *BloomClock) checkIndicies(indicies [BLOOMINDEXES]int) bool {
	var index, offset int
	for i := 0; i < BLOOMINDEXES; i++ {
		index = indicies[i] / BITSINBITE
		offset = indicies[i] % BITSINBITE
		if (b[index] & (1 << uint(offset))) == 0 {
			return false
		}
	}
	return true
}

//gets a 16 bit hash of the id, lamport
func getHash(id string, lamport int) []byte {
	//concatenate id, and lamport
	clockbytes := int2Bytes(lamport)
	id_lamport := id + string(clockbytes[:])
	//fmt.Println(id_lamport)
	h := md5.Sum([]byte(id_lamport))
	//fmt.Println(h)
	return h[0 : BLOOMINDEXES*INT16BYTES]
}

//convert id to bytes
func int2Bytes(integer int) (clockbytes [INTBYTES]byte) {
	clockbytes[3] = byte((integer >> 24) & 0xFF)
	clockbytes[2] = byte((integer >> 16) & 0xFF)
	clockbytes[1] = byte((integer >> 8) & 0xFF)
	clockbytes[0] = byte(integer & 0xFF)
	return

}

func getIndicies(h []byte) [BLOOMINDEXES]int {
	var indicies [BLOOMINDEXES]int
	for i := 0; i < BLOOMINDEXES; i++ {
		for j := 0; j < INT16BYTES; j++ {
			indicies[i] += int(h[(INT16BYTES*i)+j]) << uint(BITSINBITE*j)
		}
	}
	return indicies
}

func (b *BloomClock) String() string {
	var s string
	for i := range b {
		s += fmt.Sprintf("%x,", (*b)[i])
	}
	return s
}

func main() {
	//fmt.Println(BLOOMSIZE)
	b := NewBloomClock()
	//fmt.Println(b.String())
	//b.addIndicies([3]int{0, 8, 32})
	b.Update("alex", 1)
	if !b.Check("alex", 1) {
		fmt.Println("FAIL SINGLE CHECK")
	} else {
		fmt.Println("PASS SINGLE CHECK")
	}

	//check failure rate up to 50
	b = NewBloomClock()
	var i = 1
	var fail = 0
	for {
		if b.Check("alex", i) {
			fail++
			//fmt.Printf(b.String())
			fmt.Printf("i: %d\t Fail%d\n", i, fail)
			if fail > 50 {
				break
			}
		}
		b.Update("alex", i)
		i++
	}

	//check for merge correctness
	a, b := NewBloomClock(), NewBloomClock()
	for i := 0; i < 100; i++ {
		a.Update("stew", i)
		b.Update("alex", i)
	}
	a.Merge(b)
	for i := 0; i < 100; i++ {
		if !a.Check("stew", i) {
			panic("merge failed own clock information lost")
		}
		if !b.Check("alex", i) {
			panic("merge failed other clock information lost")
		}
	}
	fmt.Println("Pass merge")

	fmt.Println("Making a new Bloom Logger")

	al := Initalize("stew")
	buf := al.PrepareSend("sending test 1", 8)
	var t int
	al.UnpackReceive("receiving test 1", buf, &t)
	if t == 8 {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
	}

}

func setupLogger() *logging.Logger {
	// For demo purposes, create two backend for os.Stderr.

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backendFormatter := logging.NewBackendFormatter(backend, logging.MustStringFormatter(LOGFORMAT))
	// Only errors and more severe messages should be sent to backend1
	backendlevel := logging.AddModuleLevel(backendFormatter)
	backendlevel.SetLevel(logging.Level(5), "")
	// Set the backends to be used.
	logging.SetBackend(backendlevel)
	return logging.MustGetLogger("Go Bloom")

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

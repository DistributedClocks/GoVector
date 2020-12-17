package vclock

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

// Condition constants define how to compare a vector clock against another,
// and may be ORed together when being provided to the Compare method.
type Condition int

//Constants define compairison conditions between pairs of vector
//clocks
const (
	Equal Condition = 1 << iota
	Ancestor
	Descendant
	Concurrent
)

//Vector clocks are maps of string to uint64 where the string is the
//id of the process, and the uint64 is the clock value
type VClock map[string]uint64

//FindTicks returns the clock value for a given id, if a value is not
//found false is returned
func (vc VClock) FindTicks(id string) (uint64, bool) {
	ticks, ok := vc[id]
	return ticks, ok
}

//New returns a new vector clock
func New() VClock {
	return VClock{}
}

//Copy returs a copy of the clock
func (vc VClock) Copy() VClock {
	cp := make(map[string]uint64, len(vc))
	for key, value := range vc {
		cp[key] = value
	}
	return cp
}

//CopyFromMap copys a map to a vector clock
func (vc VClock) CopyFromMap(otherMap map[string]uint64) VClock {
	return otherMap
}

//GetMap returns the map typed vector clock
func (vc VClock) GetMap() map[string]uint64 {
	return map[string]uint64(vc)
}

//Set assigns a clock value to a clock index
func (vc VClock) Set(id string, ticks uint64) {
	vc[id] = ticks
}

//Tick has replaced the old update
func (vc VClock) Tick(id string) {
	vc[id] = vc[id] + 1
}

//LastUpdate returns the clock value of the oldest clock
func (vc VClock) LastUpdate() (last uint64) {
	for key := range vc {
		if vc[key] > last {
			last = vc[key]
		}
	}
	return last
}

//Merge takes the max of all clock values in other and updates the
//values of the callee
func (vc VClock) Merge(other VClock) {
	for id := range other {
		if vc[id] < other[id] {
			vc[id] = other[id]
		}
	}
}

//Bytes returns an encoded vector clock
func (vc VClock) Bytes() []byte {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(vc)
	if err != nil {
		log.Fatal("Vector Clock Encode:", err)
	}
	return b.Bytes()
}

//FromBytes decodes a vector clock
func FromBytes(data []byte) (vc VClock, err error) {
	b := new(bytes.Buffer)
	b.Write(data)
	clock := New()
	dec := gob.NewDecoder(b)
	err = dec.Decode(&clock)
	return clock, err
}

//PrintVC prints the callees vector clock to stdout
func (vc VClock) PrintVC() {
	fmt.Println(vc.ReturnVCString())
}

//ReturnVCString returns a string encoding of a vector clock
func (vc VClock) ReturnVCString() string {
	//sort
	ids := make([]string, len(vc))
	i := 0
	for id := range vc {
		ids[i] = id
		i++
	}

	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i := range ids {
		buffer.WriteString(fmt.Sprintf("\"%s\":%d", ids[i], vc[ids[i]]))
		if i+1 < len(ids) {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

//Compare takes another clock and determines if it is Equal, an
//Ancestor, Descendant, or Concurrent with the callees clock.
func (vc VClock) Compare(other VClock, cond Condition) bool {
	var otherIs Condition
	// Preliminary qualification based on length
	if len(vc) > len(other) {
		if cond&(Ancestor|Concurrent) == 0 {
			return false
		}
		otherIs = Ancestor
	} else if len(vc) < len(other) {
		if cond&(Descendant|Concurrent) == 0 {
			return false
		}
		otherIs = Descendant
	} else {
		otherIs = Equal
	}

	//Compare matching items
	for id := range other {
		if _, found := vc[id]; found {
			if other[id] > vc[id] {
				switch otherIs {
				case Equal:
					otherIs = Descendant
					break
				case Ancestor:
					return cond&Concurrent != 0
				}
			} else if other[id] < vc[id] {
				switch otherIs {
				case Equal:
					otherIs = Ancestor
					break
				case Descendant:
					return cond&Concurrent != 0
				}
			}
		} else {
			if otherIs == Equal {
				return cond&Concurrent != 0
			} else if (len(other) - len(vc) - 1) < 0 {
				return cond&Concurrent != 0
			}
		}
	}

	//Equal clocks are concurrent
	if otherIs == Equal && cond == Concurrent {
		cond = Equal
	}
	return cond&otherIs != 0
}

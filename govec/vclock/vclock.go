package vclock

import (
	"encoding/gob"
	"bytes"
	"log"
	"fmt"
	"sort"
)

// Condition constants define how to compare a vector clock against another,
// and may be ORed together when being provided to the Compare method.
type Condition int

const (
	Equal Condition = 1 << iota
	Ancestor
	Descendant
	Concurrent
)

type VClock map[string]uint64

func (vc VClock) FindTicks(id string) (uint64, bool) {
	ticks, ok := vc[id]
	return ticks, ok
}

//returns a new vector clock
func New() VClock {
	return VClock{}
}

func (vc VClock) Copy() VClock {
	cp := make(map[string]uint64,len(vc))
	for key, value := range vc {
		cp[key] = value
	}
	return cp
}

func (vc VClock) Set(id string, ticks uint64) {
	vc[id] = ticks
}

//Tick has replaced the old update
func (vc VClock) Tick(id string) {
	vc[id] = vc[id] + 1
}
	
func (vc VClock) LastUpdate() (last uint64) {
	for key := range vc {
		if vc[key] > last {
			last = vc[key]
		}
	}
	return last
}

func (vc VClock) Merge(other VClock) {
	for id := range other {
		if vc[id] < other[id] {
			vc[id] = other[id]
		}
	}
}

func (vc VClock) Bytes() []byte {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(vc)
	if err != nil {
		log.Fatal("Vector Clock Encode:", err)
	}
	return b.Bytes()
}

func FromBytes(data []byte) (vc VClock, err error) {
	b := new(bytes.Buffer)
	b.Write(data)
	clock := New()
	dec := gob.NewDecoder(b)
	err = dec.Decode(&clock)
	return clock, err
}

func (vc VClock) PrintVC() {
	fmt.Println(vc.ReturnVCString())
}

func (vc VClock) ReturnVCString() string {
	//sort
	ids := make([]string,len(vc))
	i := 0
	for id := range vc {
		ids[i] = id
		i++
	}
	sort.Strings(ids)

	var buffer bytes.Buffer
	buffer.WriteString("{")
	for i := range ids {
		buffer.WriteString(fmt.Sprintf("\"%s\":%d",ids[i],vc[ids[i]]))
		if (i +1 < len(ids) ) {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

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
					if cond&Descendant == 0 {
						return false
					}
					otherIs = Descendant
					break
				case Ancestor:
					return cond&Concurrent != 0
				}
			} else if other[id] < vc[id] {
				switch otherIs{
				case Equal:
					if cond&Ancestor == 0 {
						return false
					}
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
	return cond&otherIs != 0
}

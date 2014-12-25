// vclock - vector clocks for Go.
//
// Copyright (c) 2010-2014 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package vclock offers a vector clock implementation for Go.
//
// For more information, see the site at:
//
//    http://labix.org/vclock
//
package vclock

import (
	"errors"
	"math"
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

type itemType struct {
	id         string
	ticks      uint64
	lastUpdate uint64
}

// VClock represents a vector clock.
type VClock struct {
	hasUpdateTime bool
	items         []itemType
}

// findItem finds the index for the item with the given id.
func (vc *VClock) findItem(id string) (index int, found bool) {
	for i := range vc.items {
		if vc.items[i].id == id {
			return i, true
		}
	}
	return 0, false
}

// updateItem changes or appends the given id with ticks and when.
func (vc *VClock) updateItem(id string, ticks, when uint64) {
	if when > 0 {
		vc.hasUpdateTime = true
	}
	if i, found := vc.findItem(id); found {
		vc.items[i].ticks += ticks
		if when > vc.items[i].lastUpdate {
			vc.items[i].lastUpdate = when
		}
	} else {
		// Append new item at the end of the array.
		if cap(vc.items) < len(vc.items)+1 {
			// Updates should rarely happen more than once per vc in practice,
			// so append a single item.
			items := make([]itemType, len(vc.items)+1)
			copy(items, vc.items)
			vc.items = items
		} else {
			// But truncation pre-allocates full array.
			vc.items = vc.items[:len(vc.items)+1]
		}
		vc.items[len(vc.items)-1] = itemType{id, ticks, when}
	}
}

// New returns a new vector clock.
func New() *VClock {
	return &VClock{}
}

// Copy returns a copy of vc.
func (vc *VClock) Copy() *VClock {
	other := New()
	other.items = make([]itemType, len(vc.items))
	copy(other.items, vc.items)
	return other
}

// Update increments id's clock ticks in vc. The when update time is associated
// with id and may be used for pruning the vector clock. It may have any unit,
// but smaller values are represented in shorter space.
func (vc *VClock) Update(id string, when uint64) {
	vc.updateItem(id, 1, when)
}

// LastUpdate returns the most recent (maximum) update time of
// all ids known to vc.
func (vc *VClock) LastUpdate() (last uint64) {
	for i := 0; i != len(vc.items); i++ {
		if vc.items[i].lastUpdate > last {
			last = vc.items[i].lastUpdate
		}
	}
	return last
}

// Compare returns whether other matches any one of the conditions ORed
// together within cond (Equal, Ancestor, Descendant, or Concurrent).
func (vc *VClock) Compare(other *VClock, cond Condition) bool {
	var otherIs Condition

	lenVC := len(vc.items)
	lenOther := len(other.items)

	// Preliminary qualification based on length of vclock.
	if lenVC > lenOther {
		if cond&(Ancestor|Concurrent) == 0 {
			return false
		}
		otherIs = Ancestor
	} else if lenVC < lenOther {
		if cond&(Descendant|Concurrent) == 0 {
			return false
		}
		otherIs = Descendant
	} else {
		otherIs = Equal
	}

	// Compare matching items.
	lenDiff := lenOther - lenVC
	for oi := 0; oi != len(other.items); oi++ {
		if vci, found := vc.findItem(other.items[oi].id); found {
			otherTicks := other.items[oi].ticks
			vcTicks := vc.items[vci].ticks
			if otherTicks > vcTicks {
				if otherIs == Equal {
					if cond&Descendant == 0 {
						return false
					}
					otherIs = Descendant
				} else if otherIs == Ancestor {
					return cond&Concurrent != 0
				}
			} else if otherTicks < vcTicks {
				if otherIs == Equal {
					if cond&Ancestor == 0 {
						return false
					}
					otherIs = Ancestor
				} else if otherIs == Descendant {
					return cond&Concurrent != 0
				}
			}
		} else {
			// Other has an item which vc does not. Other must be
			// either an ancestor, or concurrent.
			if otherIs == Equal {
				// With the same length concurrent is the only choice.
				return cond&Concurrent != 0
			} else if lenDiff--; lenDiff < 0 {
				// Missing items. Can't be a descendant anymore.
				return cond&Concurrent != 0
			}
		}
	}
	return cond&otherIs != 0
}

// Merge merges other into vc, so that vc becomes a descendant of other.
// This means that every clock tick in other which doesn't exist in vc
// or which is smaller in vc will be copied from other to vc.
func (vc *VClock) Merge(other *VClock) {
	appends := 0
	for oi := range other.items {
		// First pass, updating old ticks and counting missing items.
		if vci, found := vc.findItem(other.items[oi].id); found {
			if vc.items[vci].ticks < other.items[oi].ticks {
				vc.items[vci].ticks = other.items[oi].ticks
			}
		} else {
			appends++
		}
	}
	if appends > 0 {
		// Second pass, now appending the missing ones.
		pos := len(vc.items)
		items := make([]itemType, len(vc.items)+appends)
		copy(items, vc.items)
		vc.items = items
		for oi := range other.items {
			if _, found := vc.findItem(other.items[oi].id); !found {
				vc.items[pos].id = other.items[oi].id
				vc.items[pos].ticks = other.items[oi].ticks
				pos++
			}
		}
	}
}

// Bytes returns the serialized representation of vc.
// The returned data may be loaded by FromBytes.
func (vc *VClock) Bytes() []byte {
	if len(vc.items) == 0 {
		return []byte{}
	}
	resultSize := vc.computeBytesSize()
	result := make([]byte, resultSize)
	if vc.hasUpdateTime {
		result[0] |= 0x1 // We'll store times too.
	}
	pos := 1 // result[0] is header byte.
	for i := range vc.items {
		pos += packInt(vc.items[i].ticks, result[pos:])
		if vc.hasUpdateTime {
			pos += packInt(vc.items[i].lastUpdate, result[pos:])
		}
		pos += packInt(uint64(len(vc.items[i].id)), result[pos:])
		copy(result[pos:], vc.items[i].id)
		pos += len(vc.items[i].id)
	}
	return result
}

// FromBytes returns the vector clock represented by the provided data,
// which must have been generated by VClock.Bytes.
func FromBytes(data []byte) (vc *VClock, err error) {
	vc = New()
	if len(data) == 0 {
		return vc, nil
	}
	header := data[0]
	if (header &^ 0x01) != 0 {
		return nil, errors.New("bad vclock header")
	}
	vc.hasUpdateTime = (header & 0x01) != 0
	pos := 1
	lastUpdate := uint64(0)
	for pos != len(data) {
		ticks, size, ok := unpackInt(data[pos:])
		pos += size
		if !ok || pos >= len(data) {
			return nil, errors.New("bad vclock ticks")
		}
		if vc.hasUpdateTime {
			lastUpdate, size, ok = unpackInt(data[pos:])
			pos += size
			if !ok {
				return nil, errors.New("bad vclock time")
			}
		}
		idLen, size, ok := unpackInt(data[pos:])
		pos += size
		if !ok || (pos+int(idLen)) > len(data) {
			return nil, errors.New("bad vclock id")
		}
		id := data[pos : pos+int(idLen)]
		pos += int(idLen)
		vc.updateItem(string(id), ticks, lastUpdate)
	}
	return
}

func (vc *VClock) computeBytesSize() int {
	size := 0
	for i := range vc.items {
		size += packedIntSize(vc.items[i].ticks)
		size += packedIntSize(uint64(len(vc.items[i].id)))
		size += len(vc.items[i].id)
		if vc.hasUpdateTime {
			size += packedIntSize(vc.items[i].lastUpdate)
		}
	}
	if size > 0 {
		return size + 1 // Space for the header byte.
	}
	return 0
}

// packInt packs an int in big-endian format, using the 8th
// bit of each byte as a continuation flag, meaning that the
// next byte is still part of the integer.
func packInt(value uint64, out []byte) (size int) {
	size = packedIntSize(value)
	for i := size - 1; i != -1; i-- { // Big-endian.
		out[i] = uint8(value | 0x80)
		value >>= 7
	}
	out[size-1] &^= 0x80 // Turn off the continuation bit.
	return size
}

func unpackInt(in []byte) (value uint64, size int, ok bool) {
	size = 0
	for size < len(in) && (in[size]&0x80) != 0 {
		value |= uint64(in[size]) & 0x7f
		value <<= 7
		size += 1
	}
	if size < len(in) {
		value |= uint64(in[size])
		size += 1
		ok = true
	}
	return
}

// packedIntSize returned the number of bytes used when value is
// packed via packInt.
func packedIntSize(value uint64) int {
	if value < 128 {
		return 1
	}
	return int(math.Ceil(math.Log2(float64(value+1)) / 7))
}

// Truncation defines a truncation strategy for use with Truncate.
type Truncation struct {
	// If the number of entries in the vector clock is <= KeepMinN or all the
	// remaining entries were updated on or after KeepAfter, truncation stops.
	// Otherwise, the oldest entries last updated prior to CutBefore, or getting
	// the vc above CutAboveN entries are dropped.
	KeepMinN  int
	KeepAfter uint64
	CutAboveN int
	CutBefore uint64
}

// Truncate vc using the rules defined by t.
func (vc *VClock) Truncate(t *Truncation) *VClock {
	// As an optimization, check to see if there are items to be removed
	// before going through the trouble of rebuilding the truncated VClock.
	nitems := len(vc.items)
	if nitems > t.KeepMinN {
		for i := 0; i != nitems; i++ {
			item := &vc.items[i]
			if (t.KeepAfter == 0 || item.lastUpdate < t.KeepAfter) && (item.lastUpdate < t.CutBefore || nitems > t.CutAboveN) {
				// There are items to be removed.
				return vc.actuallyTruncate(t)
			}
		}
	}
	// Nothing to do with vc.
	return vc.Copy()
}

func (vc *VClock) actuallyTruncate(t *Truncation) *VClock {
	items := sortItems(vc)
	truncated := New()
	truncated.items = make([]itemType, 0, len(vc.items)) // Pre-allocate all.
	for _, item := range items {
		if len(truncated.items) < t.KeepMinN ||
			(t.KeepAfter > 0 && item.lastUpdate > t.KeepAfter) ||
			((t.CutAboveN == 0 || len(truncated.items) < t.CutAboveN) && (item.lastUpdate >= t.CutBefore)) {
			truncated.updateItem(item.id, item.ticks, item.lastUpdate)
		}
	}
	return truncated
}

func sortItems(vc *VClock) []*itemType {
	items := make([]*itemType, len(vc.items))
	for i := 0; i != len(vc.items); i++ {
		items[i] = &vc.items[i]
	}
	sorter := itemSorter{items}
	sort.Sort(&sorter)
	return items
}

type itemSorter struct {
	items []*itemType
}

func (sorter *itemSorter) Len() int {
	return len(sorter.items)
}

func (sorter *itemSorter) Less(i, j int) bool {
	// Inverted. We want greater items first.
	return sorter.items[i].lastUpdate > sorter.items[j].lastUpdate
}

func (sorter *itemSorter) Swap(i, j int) {
	sorter.items[i], sorter.items[j] = sorter.items[j], sorter.items[i]
}

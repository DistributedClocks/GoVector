package vclock

import (
	"testing"
	"fmt"
	"./oldVclock"
)

func TestBasicInit(t *testing.T) {
	n := New()
	n.Tick("a")
	n.Tick("b")
	o := oldVclock.New()
	o.Update("a",1)
	o.Update("b",1)
	nString := n.ReturnVCString()
	oString := o.ReturnVCString()
	if nString != oString {
		t.Fatalf("new string not equal to old under basic conditions new = %s | old = %s",nString,oString)
	}
}

func TestEncodeDecode(t *testing.T) {
	n := New()
	n.Set("a",4)
	n.Set("b",1)
	n.Set("c",8)
	n.Set("d",32)
	byteClock := n.Bytes()
	decoded, err := FromBytes(byteClock)
	if err != nil {
		t.Fatal(err)
	} else if ! n.Compare(decoded, Equal) {
		nString := n.ReturnVCString()
		dString := decoded.ReturnVCString()
		t.Fatalf("decoded not the same as encoded enc = %s | dec = %s",nString,dString)
	}
}

func TestCopy(t *testing.T){
	n := New()
	n.Set("a",4)
	n.Set("b",1)
	n.Set("c",3)
	n.Set("d",2)
	nc := n.Copy()
	o := oldVclock.New()
	o.Update("a",1)
	o.Update("a",1)
	o.Update("a",1)
	o.Update("a",1)
	o.Update("b",1)
	o.Update("c",1)
	o.Update("c",1)
	o.Update("c",1)
	o.Update("d",1)
	o.Update("d",1)
	oc := o.Copy()
	an, _ := nc.FindTicks("a")
	bn, _ := nc.FindTicks("b")
	cn, _ := nc.FindTicks("c")
	dn, _ := nc.FindTicks("d")

	ao, _ := oc.FindTicks("a")
	bo, _ := oc.FindTicks("b")
	co, _ := oc.FindTicks("c")
	do, _ := oc.FindTicks("d")
	if an != ao || bn != bo || cn != co || dn != do {
		nString := nc.ReturnVCString()
		oString := oc.ReturnVCString()
		t.Fatalf("Copy not the same as the original new = %s , old = %s ",nString,oString)
	}
}

func TestMerge(t *testing.T) {
	n1 := New()
	n2 := New()
	n1.Set("b",1)
	n1.Set("a",2)
	n2.Set("b",3)
	n2.Set("c",2)

	o1 := oldVclock.New()
	o2 := oldVclock.New()
	o1.Update("a",1)
	o1.Update("a",1)
	o1.Update("b",1)
	o2.Update("b",1)
	o2.Update("b",1)
	o2.Update("b",1)
	o2.Update("c",1)
	o2.Update("c",1)

	n1.Merge(n2)
	o1.Merge(o2)
	
	an, _ := n1.FindTicks("a")
	bn, _ := n1.FindTicks("b")
	cn, _ := n1.FindTicks("c")

	ao, _ := o1.FindTicks("a")
	bo, _ := o1.FindTicks("b")
	co, _ := o1.FindTicks("c")
	if an != ao || bn != bo || cn != co {
		nString := n1.ReturnVCString()
		oString := o1.ReturnVCString()
		t.Fatalf("Copy not the same as the original new = %s , old = %s ",nString,oString)
	}
		nString := n1.ReturnVCString()
		oString := o1.ReturnVCString()
	fmt.Printf("new = %s , old = %s ",nString,oString)

}






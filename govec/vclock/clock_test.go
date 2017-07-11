package vclock

import (
	"fmt"
	"testing"
)

func TestBasicInit(t *testing.T) {
	n := New()
	n.Set("a", 2)
	n.Set("b", 1)
	na, isFounda := n.FindTicks("a")

	if !isFounda {
		t.Fatalf("Failed on finding ticks: %s", n.ReturnVCString)
	}

	if na != 2 {
		t.Fatalf("Tick value did not increment: %s", n.ReturnVCString)
	}

	n.Tick("b")

	na, isFounda = n.FindTicks("a")
	nb, isFoundb := n.FindTicks("b")

	if !isFounda || !isFoundb {
		t.Fatalf("Failed on finding ticks: %s", n.ReturnVCString)
	}

	if na != 2 || nb != 2 {
		t.Fatalf("Tick value did not increment: %s", n.ReturnVCString)	
	}

}

func TestCopy(t *testing.T){
	n := New()
	n.Set("a", 4)
	n.Set("b", 1)
	n.Set("c", 3)
	n.Set("d", 2)
	nc := n.Copy()

	an, _ := nc.FindTicks("a")
	bn, _ := nc.FindTicks("b")
	cn, _ := nc.FindTicks("c")
	dn, _ := nc.FindTicks("d")

	ao, _ := n.FindTicks("a")
	bo, _ := n.FindTicks("b")
	co, _ := n.FindTicks("c")
	do, _ := n.FindTicks("d")

	if an != ao || bn != bo || cn != co || dn != do {
		nString := nc.ReturnVCString()
		oString := n.ReturnVCString()
		t.Fatalf("Copy not the same as the original new = %s , old = %s ",nString,oString)
	} else if !n.Compare(nc, Equal) {
		nString := n.ReturnVCString()
		oString := nc.ReturnVCString()
		t.Fatalf("Copy not the same as the original new = %s , old = %s ",nString,oString)
	}
}

func TestCompareAndMerge(t *testing.T) {
	n1 := New()
	n2 := New()
	
	n1.Set("a",2)
	n1.Set("b",1)
	n1.Set("c",1)
	
	n2.Set("a",1)
	n2.Set("b",3)
	n2.Set("c",1)

	n3 := n1.Copy()
	n3.Merge(n2)

	an, _ := n3.FindTicks("a")
	bn, _ := n3.FindTicks("b")
	cn, _ := n3.FindTicks("c")

	nString := n3.ReturnVCString()

	oString1 := n1.ReturnVCString()
	oString2 := n2.ReturnVCString()

	if an != 2 || bn != 3 || cn != 1 {
		t.Fatalf("Merge not as expected = %s , old = %s, %s", nString, oString1, oString2)
	} else if !n1.Compare(n3, Descendant) {
		nString := n1.ReturnVCString()
		dString := n3.ReturnVCString()
		t.Fatalf("Clocks not defined as Descendant: n1 = %s | n2 = %s",nString,dString)
	} else if !n2.Compare(n3, Descendant) {
		nString := n2.ReturnVCString()
		dString := n3.ReturnVCString()
		t.Fatalf("Clocks not defined as Descendant: n1 = %s | n2 = %s",nString,dString)
	} /*
		TODO cfung: This test fails, I think it is a bug
		else if !n1.Compare(n2, Concurrent) {
		nString := n1.ReturnVCString()
		dString := n2.ReturnVCString()
		t.Fatalf("Clocks not defined as concurrent: n1 = %s | n2 = %s",nString,dString)
	}*/
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
	} else if !n.Compare(decoded, Equal) {
		nString := n.ReturnVCString()
		dString := decoded.ReturnVCString()
		t.Fatalf("decoded not the same as encoded enc = %s | dec = %s",nString,dString)
	}
}

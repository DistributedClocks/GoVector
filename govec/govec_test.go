package govec

import (
	"testing"
	//"fmt"
)

var TestPID string = "TestPID"

func TestBasicInit(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile")
	
	if gv.pid != TestPID {
		t.Fatalf("Setting Process ID Failed.")
	}

	vc := gv.GetCurrentVCAsClock()
	n, found := vc.FindTicks(TestPID)

	AssertTrue(t, found, "Initializing clock: Init PID not found")
	AssertEquals(t, uint64(1), n, "PrepareSend: Clock value incremented")

}

func TestLogLocal(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile")
	gv.LogLocalEvent("TestMessage1")

	vc := gv.GetCurrentVCAsClock()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "LogLocalEvent: Clock value not incremented")

}

func TestSendAndUnpackInt(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile")
	packed := gv.PrepareSend("TestMessage1", 1337)

	vc := gv.GetCurrentVCAsClock()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "PrepareSend: Clock value incremented")

	var response int
	gv.UnpackReceive("TestMessage2", packed, &response)

	vc = gv.GetCurrentVCAsClock()
	n, _ = vc.FindTicks(TestPID)
	
	AssertEquals(t, 1337, response, "PrepareSend: Clock value incremented.")
	AssertEquals(t, uint64(3), n, "PrepareSend: Clock value incremented.")
	
}

func TestSendAndUnpackStrings(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile")
	packed := gv.PrepareSend("TestMessage1", "DistClocks!")

	vc := gv.GetCurrentVCAsClock()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "PrepareSend: Clock value incremented. ")

	var response string
	gv.UnpackReceive("TestMessage2", packed, &response)

	vc = gv.GetCurrentVCAsClock()
	n, _ = vc.FindTicks(TestPID)
	
	AssertEquals(t, "DistClocks!", response, "PrepareSend: Clock value incremented.")
	AssertEquals(t, uint64(3), n, "PrepareSend: Clock value incremented.")
	
}

func BenchmarkPrepare(b *testing.B) {

	gv := InitGoVector(TestPID, "TestLogFile")

	var packed []byte

	for i := 0; i < b.N; i++ {
		packed = gv.PrepareSend("TestMessage1", 1337)
	}

	var response int
	gv.UnpackReceive("TestMessage2", packed, &response)

}

func BenchmarkUnpack(b *testing.B) {

	gv := InitGoVector(TestPID, "TestLogFile")

	var packed []byte
	packed = gv.PrepareSend("TestMessage1", 1337)

	var response int

	for i := 0; i < b.N; i++ {
		gv.UnpackReceive("TestMessage2", packed, &response)
	}	

}

func AssertTrue(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatalf(message)
	}
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf(message + "Expected: %s, Actual: %s", expected, actual)	
	}
}
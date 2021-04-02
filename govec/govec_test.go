package govec

import (
	"testing"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	//"fmt"
)

var TestPID = "TestPID"

func TestBasicInit(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())

	if gv.pid != TestPID {
		t.Fatalf("Setting Process ID Failed.")
	}

	vc := gv.GetCurrentVC()
	n, found := vc.FindTicks(TestPID)

	AssertTrue(t, found, "Initializing clock: Init PID not found")
	AssertEquals(t, uint64(1), n, "Initializing clock: wrong initial clock value")

}

func TestInitialVC(t *testing.T) {
	initialVC := vclock.VClock(map[string]uint64{
		TestPID: 7,
	})

	config := GetDefaultConfig()
	config.InitialVC = initialVC.Copy()
	gv := InitGoVector(TestPID, "TestLogFile", config)

	vc := gv.GetCurrentVC()
	n, found := vc.FindTicks(TestPID)

	AssertTrue(t, found, "Initializing clock: Init PID not found")
	AssertEquals(t, initialVC[TestPID]+1, n, "Initializing clock: wrong initial clock value")
}

func TestLogLocal(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())
	opts := GetDefaultLogOptions()
	gv.LogLocalEvent("TestMessage1", opts)

	vc := gv.GetCurrentVC()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "LogLocalEvent: Clock value not incremented")

}

func TestSendAndUnpackInt(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())
	opts := GetDefaultLogOptions()
	packed := gv.PrepareSend("TestMessage1", 1337, opts)

	vc := gv.GetCurrentVC()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "PrepareSend: Clock value incremented")

	var response int
	gv.UnpackReceive("TestMessage2", packed, &response, opts)

	vc = gv.GetCurrentVC()
	n, _ = vc.FindTicks(TestPID)

	AssertEquals(t, 1337, response, "PrepareSend: Clock value incremented.")
	AssertEquals(t, uint64(3), n, "PrepareSend: Clock value incremented.")

}

func TestSendAndUnpackStrings(t *testing.T) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())
	opts := GetDefaultLogOptions()
	packed := gv.PrepareSend("TestMessage1", "DistClocks!", opts)

	vc := gv.GetCurrentVC()
	n, _ := vc.FindTicks(TestPID)

	AssertEquals(t, uint64(2), n, "PrepareSend: Clock value incremented. ")

	var response string
	gv.UnpackReceive("TestMessage2", packed, &response, opts)

	vc = gv.GetCurrentVC()
	n, _ = vc.FindTicks(TestPID)

	AssertEquals(t, "DistClocks!", response, "PrepareSend: Clock value incremented.")
	AssertEquals(t, uint64(3), n, "PrepareSend: Clock value incremented.")

}

func BenchmarkPrepare(b *testing.B) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())
	opts := GetDefaultLogOptions()

	var packed []byte

	for i := 0; i < b.N; i++ {
		packed = gv.PrepareSend("TestMessage1", 1337, opts)
	}

	var response int
	gv.UnpackReceive("TestMessage2", packed, &response, opts)

}

func BenchmarkUnpack(b *testing.B) {

	gv := InitGoVector(TestPID, "TestLogFile", GetDefaultConfig())
	opts := GetDefaultLogOptions()

	var packed []byte
	packed = gv.PrepareSend("TestMessage1", 1337, opts)

	var response int

	for i := 0; i < b.N; i++ {
		gv.UnpackReceive("TestMessage2", packed, &response, opts)
	}

}

func AssertTrue(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatalf(message)
	}
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf(message+"Expected: %s, Actual: %s", expected, actual)
	}
}

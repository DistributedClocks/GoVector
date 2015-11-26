package broker_test

import (
    "testing"
    //"io"
	"os"
  	"bufio"
    . "gopkg.in/check.v1"
	//"fmt"
	"net"
	//"net/rpc/jsonrpc"
	//"net/http"
	//"encoding/json"
	//"time"
	//"sync"
	"./../server/broker"
	"./../govec"
)

//global variable for handling all server traffic
const brokeraddr string = "127.0.0.1"
const brokerpubport string = "8000"
const brokersubport string = "8080"
const brokerlogfile string = "./test"

// Hook up gocheck into the "go test" runner.
// Maybe don't need this
func Test(t *testing.T) { TestingT(t) }

type BrokerSuite struct{
	broker brokervec.VectorBroker
	gopub *govec.GoPublisher
}

var s = Suite(&BrokerSuite{})

func (s *BrokerSuite) SetUpSuite(c *C) {

	// Start the broker
	go s.broker.Init(brokerlogfile)

	// Check if publisher port is open
	puburl := brokeraddr + ":" + brokerpubport
	pubconn, err := net.Dial("tcp", puburl)
	c.Check(err, IsNil)
	pubconn.Close()
	
	// Check if subscriber port is open
	suburl := brokeraddr + ":" + brokersubport
	subconn, err := net.Dial("tcp", suburl)
	c.Check(err, IsNil)
	subconn.Close()

	s.gopub = govec.NewGoPublisher(brokeraddr, brokerpubport)
}

func (s *BrokerSuite) TestPublishLocalMessage(c *C) {
	testpid := "42"
	testvcstring := "[0, 0]"
	testmessage := "This is a local test message"
	s.gopub.PublishLocalMessage(testmessage, testpid, testvcstring)

	// Assert that line is the last line of the log file.
	message := testpid + " " + testvcstring + testmessage
	logpath := brokerlogfile + "-Log.txt"
	result, err := readLines(logpath)
	c.Check(err, IsNil)
	c.Check(result, Equals, message)
}

func readLines(path string) (string, error) {
  file, err := os.Open(path)
  if err != nil {
    return "", err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines[len(lines)-2]+lines[len(lines)-1], scanner.Err()
}


// Test functions...
//func connectTestPublishers(num int) []govec.GoPublisher {
//	var gp []govec.GoPublisher
//	gp = make([]govec.GoPublisher, num)
	
//	for i := 0; i < num; i++ {
//		gp[i] = govec.NewGoPublisher(brokeraddr, brokerpubport)
//	}
//	return gp
//}

	
//	gp1 := govec.NewGoPublisher("127.0.0.1", "8000")
//	fmt.Println("Registered publisher gp1, sending test message")
	
//	gp.SendTestMessage()
//	gp1.SendTestMessage()
	
//	gp.SendLocalMessage("test", "test", "test")
//	gp.SendNetworkMessage("net", "net", "net")
	
//	gp1.SendLocalMessage("test", "test", "test")
//	gp1.SendNetworkMessage("net", "net", "net")
	
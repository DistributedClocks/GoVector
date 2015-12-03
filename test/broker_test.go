package broker_test

import (
    "testing"
    //"io"
	"golang.org/x/net/websocket"
	"os"
  	"bufio"
    . "gopkg.in/check.v1"
	//"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	//"net/http"
	//"encoding/json"
	"time"
	//"sync"
	"./../server/broker"
	//"./../server/broker/nonce"
	"./../govec"
	"strings"
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
	go s.broker.Init(brokerlogfile, brokerpubport, brokersubport)

	// Check if publisher port is open
	puburl := brokeraddr + ":" + brokerpubport
	pubconn, err := net.Dial("tcp", puburl)
	c.Assert(err, IsNil)
	pubconn.Close()
	
	// Check if subscriber port is open
	suburl := brokeraddr + ":" + brokersubport
	subconn, err := net.Dial("tcp", suburl)
	c.Assert(err, IsNil)
	subconn.Close()

	s.gopub = govec.NewGoPublisher(brokeraddr, brokerpubport)
}

func (s *BrokerSuite) TestPublishLocalMessage(c *C) {
	testpid := "13"
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

func (s *BrokerSuite) TestPublishNetworkMessage(c *C) {
	testpid := "42"
	testvcstring := "[1, 2]"
	testmessage := "This is a network test message"
	s.gopub.PublishNetworkMessage(testmessage, testpid, testvcstring)

	// Assert that line is the last line of the log file.
	message := testpid + " " + testvcstring + testmessage
	logpath := brokerlogfile + "-Log.txt"
	result, err := readLines(logpath)
	c.Check(err, IsNil)
	c.Check(result, Equals, message)
}

func (s *BrokerSuite) TestConnectSubscriber(c *C) {
	nonce, jrpc, err := connectSubscriber("TestConnectSubscriber", brokersubport)
	defer jrpc.Close()
	
	c.Assert(err, IsNil)
	c.Assert(jrpc, NotNil)
	c.Assert(nonce, Not(Equals), "")
}

func (s *BrokerSuite) TestConnectSubscriberToWrongPort(c *C) {
	// Try to connect to open but wrong port.
	_, _, err := connectSubscriber("TestConnectSubscriber", brokerpubport)	
	c.Assert(err, NotNil)
	
	// Try to connect to random port.
	_, _, err2 := connectSubscriber("TestConnectSubscriber", "8010")	
	c.Assert(err2, NotNil)
}

func (s *BrokerSuite) TestSubscriberRPC(c *C) {
	nonce, jrpc, err := connectSubscriber("TestSubscriberRPC", brokersubport)
	
	c.Assert(err, IsNil)
	c.Assert(jrpc, NotNil)
	c.Assert(nonce, Not(Equals), "")
	
	message := brokervec.FilterMessage{
		Regex: "TestSubscriberRPC",
		Nonce: nonce}
		
	var reply string
	jerr := jrpc.Call("SubManager.AddFilter", message, &reply)

	c.Assert(jerr, IsNil)
	expected_reply := "Your subscriber exists!"
	c.Check(reply, Equals, expected_reply)

	jerr = jrpc.Call("SubManager.AddFilter", message, &reply)

	c.Assert(jerr, IsNil)
	c.Check(reply, Equals, expected_reply)
}

func connectSubscriber(testname string, port string) (nonce string, jrpc *rpc.Client, err error) {
	origin := "http://" + brokeraddr + "/"
	url := "ws://" + brokeraddr + ":" + port + "/ws"
	
	ws, err := websocket.Dial(url, "", origin)
	
	if err != nil {
        return "", nil, err
    }
	
	ws.SetDeadline(time.Now().Add(time.Duration(3e8)))

	websocket.Message.Send(ws, testname)
	
	nonce = ""
	websocket.Message.Receive(ws, &nonce)
	
	nonce = strings.Replace(nonce, "\"", "", -1)
		
	jrpc = jsonrpc.NewClient(ws)
	
	return nonce, jrpc, err
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



	
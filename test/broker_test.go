package broker_test

import (
    "testing"
    //"io"
    "golang.org/x/net/websocket"
    "os"
      "bufio"
    . "gopkg.in/check.v1"
    "net"
    //"net/rpc"
    "net/rpc/jsonrpc"
    //"net/http"
    //"encoding/json"
    "time"
    "./../server/broker"
    "./../govec"
    "strings"
    "strconv"
    "github.com/arcaneiceman/GoVector/govec/vclock"
)

/*
    To use this test package, type "go test". 
    
    The test package uses Go-Check (https://labix.org/gocheck). To get this 
    package type "go get gopkg.in/check.v1" into your go enabled console.
*/


//global variable for handling all server traffic
const brokeraddr string = "127.0.0.1"
const brokerpubport string = "8000"
const brokersubport string = "8080"
const brokerlogfile string = "./test_broker"

// Hook up gocheck into the "go test" runner.
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
    testmessage := "This is a local test message"
    testvc := vclock.New()
    testvc.Update(testpid, 1)
    s.gopub.PublishLocalMessage(testmessage, testpid, *testvc)

    // Assert that line is the last line of the log file.
    message := testpid + " " + testvc.ReturnVCString() + testmessage
    logpath := brokerlogfile + "-Log.txt"
    result, err := readLines(logpath)
    c.Assert(err, IsNil)
    c.Check(result, Equals, message)
}

func (s *BrokerSuite) TestPublishNetworkMessage(c *C) {
    testpid := "42"
    //testvcstring := "[1, 2]"
    testmessage := "This is a network test message"
    testvc := vclock.New()
    testvc.Update(testpid, 1)
    s.gopub.PublishLocalMessage(testmessage, testpid, *testvc)

    // Assert that line is the last line of the log file.
    message := testpid + " " + testvc.ReturnVCString() + testmessage
    logpath := brokerlogfile + "-Log.txt"
    result, err := readLines(logpath)
    c.Assert(err, IsNil)
    c.Check(result, Equals, message)
}

func (s *BrokerSuite) TestConnectSubscriber(c *C) {
    nonce, ws, err := connectSubscriber("TestConnect", brokersubport)
    defer ws.Close()
    
    c.Assert(err, IsNil)
    c.Assert(ws, NotNil)
    c.Assert(nonce, Not(Equals), "")
}

func (s *BrokerSuite) TestConnectSubscriberToWrongPort(c *C) {
    // Try to connect to open but wrong port.
    _, _, err := connectSubscriber("TestConnectErr", brokerpubport)    
    c.Assert(err, NotNil)
    
    // Try to connect to random port.
    _, _, err2 := connectSubscriber("TestConnectErr", "8010")    
    c.Assert(err2, NotNil)
}

func (s *BrokerSuite) TestSubscriberReceive(c *C) {
    nonce, ws, err := connectSubscriber("TestReceive", brokersubport)
    
    c.Assert(err, IsNil)
    c.Assert(ws, NotNil)    
    c.Assert(nonce, Not(Equals), "")
    
    defer ws.Close()
    
    // Send a message
    testpid := "324"
    testmessage := "This is a test message for subscribers"
    testvc := vclock.New()
    testvc.Update(testpid, 1)
    s.gopub.PublishLocalMessage(testmessage, testpid, *testvc)

    // Receive the message
    var reply string
    websocket.Message.Receive(ws, &reply)
    expectedreply := "324 {\"324\":1}\nThis is a test message for subscribers\n"
    c.Assert(reply, Equals, expectedreply)
}

func (s *BrokerSuite) TestSubscriberRPC(c *C) {
    nonce, ws, err := connectSubscriber("TestSubscriberRPC", brokersubport)
    c.Assert(err, IsNil)
    c.Assert(ws, NotNil)    
    c.Assert(nonce, Not(Equals), "")
    
    defer ws.Close()
    
    jrpc := jsonrpc.NewClient(ws)
    defer jrpc.Close()
    c.Assert(jrpc, NotNil)
    
    message := brokervec.FilterMessage{
        Regex: "Fake Regex",
        Nonce: nonce}
        
    var reply string
    jerr := jrpc.Call("SubManager.AddFilter", message, &reply)

    c.Assert(jerr, IsNil)
    expected_reply := "Added filter: Fake Regex"
    c.Check(reply, Equals, expected_reply)

    jerr = jrpc.Call("SubManager.AddFilter", message, &reply)

    c.Assert(jerr, IsNil)
    c.Check(reply, Equals, expected_reply)
}

func (s *BrokerSuite) TestGetOldMessages(c *C) {
    testpid := "4"
    //testvcstring := "[6, 6]"
    testmessage := "This is an old message"
    testvc := vclock.New()
    testvc.Update(testpid, 1)
    s.gopub.PublishLocalMessage(testmessage, testpid, *testvc)
    
    nonce, ws, err := connectSubscriber("TestOldMessage", brokersubport)
    c.Assert(err, IsNil)
    c.Assert(ws, NotNil)    
    c.Assert(nonce, Not(Equals), "")
    
    defer ws.Close()
    
    jrpc := jsonrpc.NewClient(ws)
    defer jrpc.Close()
    c.Assert(jrpc, NotNil)

    message := nonce
        
    var reply string
    jerr := jrpc.Call("SubManager.SendOldMessages", message, &reply)
    
    c.Assert(jerr, IsNil)
    numMessages, cerr := strconv.Atoi(reply)
    c.Check(cerr, IsNil)

    result := numMessages >= 1
    c.Check(result, Equals, true)

}

func connectSubscriber(testname string, port string) (nonce string, ws *websocket.Conn, err error) {
    origin := "http://" + brokeraddr + "/"
    url := "ws://" + brokeraddr + ":" + port + "/ws"
    
    ws, err = websocket.Dial(url, "", origin)
    
    if err != nil {
        return "", nil, err
    }
    
    ws.SetDeadline(time.Now().Add(time.Duration(3e9)))

    websocket.Message.Send(ws, testname)
    
    nonce = ""
    websocket.Message.Receive(ws, &nonce)
    
    nonce = strings.Replace(nonce, "\"", "", -1)
    
    return nonce, ws, err
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



    
//Package vrpc provides support for automatically logging RPC Calls
//from a RPC Client to a RPC Server
package vrpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/rpc"

	"github.com/DistributedClocks/GoVector/govec"
)

//RPCDial connects to a RPC server at the specified network address. The
//logger is provided to be used by the RPCClientCodec for message
//capture.
func RPCDial(network, address string, logger *govec.GoLog) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return rpc.NewClientWithCodec(newClientCodec(conn, logger)), err
}

//ServerRPCCon is a convenience function that accepts connections for a
//given listener and starts a new goroutine for the server to serve a
//new connection. The logger is provided to be used by the
//RPCServerCodec for message capture.
func ServeRPCConn(server *rpc.Server, l net.Listener, logger *govec.GoLog) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go server.ServeCodec(newServerCodec(conn, logger))
	}
}

//RPCClientCodec is an extension of the default rpc codec which uses a
//logger of type GoLog to capture all the calls to a RPC Server as
//well as responses from a RPC server.
type RPCClientCodec struct {
	C      io.Closer
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *govec.GoLog
}

//RPCServerCodec is an extension of the default rpc codec which uses a
//logger of type of GoLog to capture all the requests made from the
//client to a RPC server as well as the server's to the clients.
type RPCServerCodec struct {
	Rwc    io.ReadWriteCloser
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *govec.GoLog
	Closed bool
}

//NewClient returs an rpc.Client insturmented with vector clocks.
func NewClient(conn io.ReadWriteCloser, logger *govec.GoLog) *rpc.Client {
	return rpc.NewClientWithCodec(newClientCodec(conn, logger))
}

func newClientCodec(conn io.ReadWriteCloser, logger *govec.GoLog) rpc.ClientCodec {
	encBuf := bufio.NewWriter(conn)
	return &RPCClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf, logger}
}

//WriteRequest marshalls and sends an rpc request, and it's associated
//parameters to an RPC server
func (c *RPCClientCodec) WriteRequest(req *rpc.Request, param interface{}) (err error) {
	if err = c.Enc.Encode(req); err != nil {
		return
	}
	opts := govec.GetDefaultLogOptions()
	buf := c.Logger.PrepareSend("Making RPC call", param, opts)
	if err = c.Enc.Encode(buf); err != nil {
		return
	}

	return c.EncBuf.Flush()
}

//ReadResponseHeader deacodes an RPC response header on the client
func (c *RPCClientCodec) ReadResponseHeader(resp *rpc.Response) error {
	return c.Dec.Decode(resp)
}

//ReadResponseBody decodes a response body and updates it's local
//vector clock with that of the server.
func (c *RPCClientCodec) ReadResponseBody(body interface{}) (err error) {
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	opts := govec.GetDefaultLogOptions()
	c.Logger.UnpackReceive("Received RPC Call response from server", buf, body, opts)
	return nil
}

//Close closes an RPCClientCodecs internal TCP connection
func (c *RPCClientCodec) Close() error {
	return c.C.Close()
}

func newServerCodec(conn io.ReadWriteCloser, logger *govec.GoLog) rpc.ServerCodec {
	buf := bufio.NewWriter(conn)
	srv := &RPCServerCodec{
		Rwc:    conn,
		Dec:    gob.NewDecoder(conn),
		Enc:    gob.NewEncoder(buf),
		EncBuf: buf,
		Logger: logger,
	}
	return srv
}

//ReadRequestHeader decodes a server rpc request header
func (c *RPCServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.Dec.Decode(r)
}

//ReadRequestBody decodes a clinet request and updates the servers
//local vector clock with the clients values
func (c *RPCServerCodec) ReadRequestBody(body interface{}) (err error) {
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	opts := govec.GetDefaultLogOptions()
	c.Logger.UnpackReceive("Received RPC request", buf, body, opts)
	return nil
}

//WriteResponse sends an rpc response, and it's associated result back
//to the client
func (c *RPCServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	Encode(c, r)
	opts := govec.GetDefaultLogOptions()
	buf := c.Logger.PrepareSend("Sending response to RPC request", body, opts)
	Encode(c, buf)
	return c.EncBuf.Flush()
}

//Encode is a convience function which writes to the wire and handels
//RPC errors
func Encode(c *RPCServerCodec, payload interface{}) {
	if err := c.Enc.Encode(payload); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			c.Close()
			panic(err)
		}
	}
}

//Close ends the underlying server TCP session
func (c *RPCServerCodec) Close() error {
	if c.Closed {
		return nil
	}
	c.Closed = true
	return c.Rwc.Close()
}

//GoVector provides support for automatically logging RPC Calls from a
//RPC Client to a RPC Server
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

//Convenience function that accepts connections for a given listener and
//starts a new goroutine for the server to serve a new connection. The
//logger is provided to be used by the RPCServerCodec for message
//capture.
func ServeRPCConn(server *rpc.Server, l net.Listener, logger *govec.GoLog) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go server.ServeCodec(newServerCodec(conn, logger))
	}
}

//An extension of the default rpc codec which uses a logger of type
//GoLog to capture all the calls to a RPC Server as well as responses
//from a RPC server.
type RPCClientCodec struct {
	C      io.Closer
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *govec.GoLog
}

//An extension of the default rpc codec which uses a logger of type of
//GoLog to capture all the requests made from the client to a RPC server
//as well as the server's to the clients.
type RPCServerCodec struct {
	Rwc    io.ReadWriteCloser
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *govec.GoLog
	Closed bool
}

func NewClient(conn io.ReadWriteCloser, logger *govec.GoLog) *rpc.Client {
	return rpc.NewClientWithCodec(newClientCodec(conn, logger))
}

func newClientCodec(conn io.ReadWriteCloser, logger *govec.GoLog) rpc.ClientCodec {
	encBuf := bufio.NewWriter(conn)
	return &RPCClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf, logger}
}

func (c *RPCClientCodec) WriteRequest(req *rpc.Request, param interface{}) (err error) {
	if err = c.Enc.Encode(req); err != nil {
		return
	}
	buf := c.Logger.PrepareSend("Making RPC call", param)
	if err = c.Enc.Encode(buf); err != nil {
		return
	}

	return c.EncBuf.Flush()
}

func (c *RPCClientCodec) ReadResponseHeader(resp *rpc.Response) error {
	return c.Dec.Decode(resp)
}

func (c *RPCClientCodec) ReadResponseBody(body interface{}) (err error) {
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	c.Logger.UnpackReceive("Received RPC Call response from server", buf, body)
	return nil
}

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

func (c *RPCServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.Dec.Decode(r)
}

func (c *RPCServerCodec) ReadRequestBody(body interface{}) (err error) {
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	c.Logger.UnpackReceive("Received RPC request", buf, body)
	return nil
}

func (c *RPCServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	Encode(c, r)
	Encode(c, body)
	return c.EncBuf.Flush()
}

func Encode(c *RPCServerCodec, payload interface{}) {
	if err := c.Enc.Encode(payload); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			c.Close()
			panic(err)
		}
	}
}

func (c *RPCServerCodec) Close() error {
	if c.Closed {
		return nil
	}
	c.Closed = true
	return c.Rwc.Close()
}

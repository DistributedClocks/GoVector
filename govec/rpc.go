package govec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
	"net/rpc"
)

/* Functions to add Codecs to Client RPC calls */
func RPCDial(network, address string, logger *GoLog) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return rpc.NewClientWithCodec(NewClientCodec(conn, logger)), err
}

func ServeRPCConn(server *rpc.Server, l net.Listener, logger *GoLog) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go server.ServeCodec(NewServerCodec(conn, logger))
	}
}

type RPCClientCodec struct {
	C      io.Closer
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *GoLog
}

type RPCServerCodec struct {
	Rwc    io.ReadWriteCloser
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Logger *GoLog
	Closed bool
}

func NewClient(conn io.ReadWriteCloser, logger *GoLog) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn, logger))
}

func NewClientCodec(conn io.ReadWriteCloser, logger *GoLog) rpc.ClientCodec {
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

func NewServerCodec(conn io.ReadWriteCloser, logger *GoLog) rpc.ServerCodec {
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

package capture

import (
	"bufio"
	"bitbucket.org/bestchai/dinv/instrumenter"
	"net"
	"net/rpc"
	"fmt"
	"io"
	"encoding/gob"
)


/*
Functions for appending vector clocks to the standard net package
*/
func Read(read func([]byte) (int, error), b []byte) (int, error) {
	buf := make([]byte,len(b))
	n, err := read(buf)
	instrumenter.Unpack(buf,&b)
	return n, err
}

//func WrapReadFrom(readFrom func([]byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
func ReadFrom(readFrom func([]byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
	buf := make([]byte,len(b))
	n, addr, err := readFrom(buf)
	instrumenter.Unpack(buf,&b)
	return n, addr, err
}

func Write(write func (b []byte) (int, error), b []byte) (int, error) {
	buf := instrumenter.Pack(b)
	n, err := write(buf)
	return n, err
}

func WriteTo(writeTo func ([]byte,net.Addr) (int, error), b []byte, addr net.Addr) (int, error) {
	buf := instrumenter.Pack(b)
	n, err := writeTo(buf, addr)
	return n, err
}

/* Functions to add Codecs to Client RPC calls */
func Dial(dial func (string,string) (*rpc.Client, error), network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network,address)
	if err != nil {
		return nil, err
	}
	return rpc.NewClientWithCodec(NewClientCodec(conn)), err
}

//TODO
func DialHTTP(dialHttp func (string,string) (*rpc.Client, error), network, address string) (*rpc.Client, error) {
	return nil, fmt.Errorf("Dinv has yet to implement an rpc.DialHTTP wrapper\n")
}

//TODO
func DialHTTPPath(dialHttpPath func (string,string,string) (*rpc.Client, error), network, address, path string) (*rpc.Client, error) {
	return nil, fmt.Errorf("Dinv has yet to implement an rpc.DialHTTP wrapper\n")
}

func NewClient(newClient func (io.ReadWriteCloser) (*rpc.Client), conn io.ReadWriteCloser) (*rpc.Client) {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

//TODO
func NewClientWithCodec(newClientWithCodec func(rpc.ClientCodec) (*rpc.Client), codec rpc.ClientCodec) (*rpc.Client) {
	return nil
}

/* Functions for adding codecs to Server RPC */

//TODO
func ServeCodec(serveCodec func(rpc.ServerCodec), codec rpc.ServerCodec) {
	return
}

func ServeConn(serveConn func(io.ReadWriteCloser), conn io.ReadWriteCloser) {
	rpc.DefaultServer.ServeCodec(NewServerCodec(conn))
}

//TODO
func ServeRequest(serveRequest func(rpc.ServerCodec) error, codec rpc.ServerCodec) error {
	return fmt.Errorf("Dinv has yet to implement an rpc.ServeRequest wrapper\n")
}

//TODO the interalfunction for the rpc.Server have yet to be
//implemented

/*
	Client Codec for instrumenting RPC calls with vector clocks
*/

type ClientClockCodec struct {
	C io.Closer
	Dec *gob.Decoder
	Enc *gob.Encoder
	EncBuf *bufio.Writer
}

func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	encBuf := bufio.NewWriter(conn)
	return &ClientClockCodec{conn,gob.NewDecoder(conn),gob.NewEncoder(encBuf),encBuf}
}

func (c *ClientClockCodec) WriteRequest(req *rpc.Request, param interface{}) (err error) {
	fmt.Println("WriteRequest")
	if err = c.Enc.Encode(req); err != nil {
		return
	}
	buf := instrumenter.Pack(param)
	if err = c.Enc.Encode(buf); err != nil {
		return
	}

	return c.EncBuf.Flush()
}

func (c *ClientClockCodec) ReadResponseHeader( resp * rpc.Response) error {
	fmt.Println("ReadResponseHeader")
	return c.Dec.Decode(resp)
}

func (c * ClientClockCodec) ReadResponseBody( body interface{}) (err error) {
	fmt.Println("ReadResponseBody")
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	instrumenter.Unpack(buf,body)
	return nil
}

func (c *ClientClockCodec) Close() error {
	return c.C.Close()
}

/*
	Client Codec for instrumenting RPC calls with vector clocks
*/

type ServerClockCodec struct {
	Rwc io.ReadWriteCloser
	Dec *gob.Decoder
	Enc *gob.Encoder
	EncBuf *bufio.Writer
	Closed bool
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	buf := bufio.NewWriter(conn)
	srv := &ServerClockCodec{
		Rwc:	conn,
		Dec:	gob.NewDecoder(conn),
		Enc:	gob.NewEncoder(buf),
		EncBuf: buf,
	}
	return srv
}

func (c *ServerClockCodec) ReadRequestHeader(r *rpc.Request) error {
	fmt.Println("ReadRequestHeader")
	return c.Dec.Decode(r)
}
func (c *ServerClockCodec) ReadRequestBody(body interface{}) (err error) {
	fmt.Println("ReadResponseBody")
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	instrumenter.Unpack(buf,body)
	return nil
	
}
func (c *ServerClockCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.Enc.Encode(r); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			fmt.Println("RPC Error encoding response:",err)
			c.Close()
		}
		return
	}
	buf := instrumenter.Pack(body)
	if err = c.Enc.Encode(buf); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			fmt.Println("RPC Error encoding body (This is likely Dinv's fault):",err)
			c.Close()
		}
		return
	}
	return c.EncBuf.Flush()
}

func (c *ServerClockCodec) Close() error {
	if c.Closed {
		return nil
	}
	c.Closed = true
	return c.Rwc.Close()
}



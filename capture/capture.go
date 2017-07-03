package capture

import (
	"bitbucket.org/bestchai/dinv/dinvRT"
	"bufio"

	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/rpc"
)

/*
Functions for appending vector clocks to the standard net package
*/
func Read(read func([]byte) (int, error), b []byte) (int, error) {
	buf := make([]byte, len(b))
	n, err := read(buf)
	dinvRT.Unpack(buf, &b)
	return n, err
}

func ReadFrom(readFrom func([]byte) (int, net.Addr, error), b []byte) (int, net.Addr, error) {
	buf := make([]byte, len(b))
	n, addr, err := readFrom(buf)
	dinvRT.Unpack(buf, &b)
	return n, addr, err
}

func Write(write func(b []byte) (int, error), b []byte) (int, error) {
	buf := dinvRT.Pack(b)
	n, err := write(buf)
	return n, err
}

func WriteTo(writeTo func([]byte, net.Addr) (int, error), b []byte, addr net.Addr) (int, error) {
	buf := dinvRT.Pack(b)
	n, err := writeTo(buf, addr)
	return n, err
}

//Protocol specific funcions
func ReadFromIP(readfromip func([]byte) (int, *net.IPAddr, error), b []byte) (int, *net.IPAddr, error) {
	buf := make([]byte, len(b))
	n, addr, err := readfromip(buf)
	dinvRT.Unpack(buf, &b)
	return n, addr, err
}

func ReadMsgIP(readmsgip func([]byte, []byte) (int, int, int, *net.IPAddr, error), b, oob []byte) (int, int, int, *net.IPAddr, error) {
	buf := make([]byte, len(b))
	n, oobn, flags, addr, err := readmsgip(buf, oob)
	dinvRT.Unpack(buf, &b)
	return n, oobn, flags, addr, err
}

func WriteMsgIP(writemsgip func([]byte, []byte, *net.IPAddr) (int, int, error), b, oob []byte, addr *net.IPAddr) (int, int, error) {
	buf := dinvRT.Pack(b)
	n, oobn, err := writemsgip(buf, oob, addr)
	return n, oobn, err
}

func WriteToIP(writetoip func([]byte, *net.IPAddr) (int, error), b []byte, addr *net.IPAddr) (int, error) {
	buf := dinvRT.Pack(b)
	n, err := writetoip(buf, addr)
	return n, err
}

func ReadFromUDP(readfromudp func([]byte) (int, *net.UDPAddr, error), b []byte) (int, *net.UDPAddr, error) {
	buf := make([]byte, len(b))
	n, addr, err := readfromudp(buf)
	dinvRT.Unpack(buf, &b)
	return n, addr, err
}

func ReadMsgUDP(readmsgudp func([]byte, []byte) (int, int, int, *net.UDPAddr, error), b, oob []byte) (int, int, int, *net.UDPAddr, error) {
	buf := make([]byte, len(b))
	n, oobn, flags, addr, err := readmsgudp(b, oob)
	dinvRT.Unpack(buf, &b)
	return n, oobn, flags, addr, err
}

func WriteMsgUDP(writemsgudp func([]byte, []byte, *net.UDPAddr) (int, int, error), b, oob []byte, addr *net.UDPAddr) (int, int, error) {
	buf := dinvRT.Pack(b)
	n, oobn, err := writemsgudp(buf, oob, addr)
	return n, oobn, err
}

func WriteToUDP(writetoudp func([]byte, *net.UDPAddr) (int, error), b []byte, addr *net.UDPAddr) (int, error) {
	buf := dinvRT.Pack(b)
	n, err := writetoudp(buf, addr)
	return n, err
}

func ReadFromUnix(readfromunix func([]byte) (int, *net.UnixAddr, error), b []byte) (int, *net.UnixAddr, error) {
	buf := make([]byte, len(b))
	n, addr, err := readfromunix(buf)
	dinvRT.Unpack(buf, &b)
	return n, addr, err
}

func ReadMsgUnix(readmsgunix func([]byte, []byte) (int, int, int, *net.UnixAddr, error), b, oob []byte) (int, int, int, *net.UnixAddr, error) {
	buf := make([]byte, len(b))
	n, oobn, flags, addr, err := readmsgunix(b, oob)
	dinvRT.Unpack(buf, &b)
	return n, oobn, flags, addr, err
}

func WriteMsgUnix(writemsgunix func([]byte, []byte, *net.UnixAddr) (int, int, error), b, oob []byte, addr *net.UnixAddr) (int, int, error) {
	buf := dinvRT.Pack(b)
	n, oobn, err := writemsgunix(buf, oob, addr)
	return n, oobn, err
}

func WriteToUnix(writetounix func([]byte, *net.UnixAddr) (int, error), b []byte, addr *net.UnixAddr) (int, error) {
	buf := dinvRT.Pack(b)
	n, err := writetounix(buf, addr)
	return n, err
}

/* Functions to add Codecs to Client RPC calls */
func Dial(dial func(string, string) (*rpc.Client, error), network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return rpc.NewClientWithCodec(NewClientCodec(conn)), err
}

//TODO
func DialHTTP(dialHttp func(string, string) (*rpc.Client, error), network, address string) (*rpc.Client, error) {
	return nil, fmt.Errorf("Dinv has yet to implement an rpc.DialHTTP wrapper\n")
}

//TODO
func DialHTTPPath(dialHttpPath func(string, string, string) (*rpc.Client, error), network, address, path string) (*rpc.Client, error) {
	return nil, fmt.Errorf("Dinv has yet to implement an rpc.DialHTTP wrapper\n")
}

func NewClient(newClient func(io.ReadWriteCloser) *rpc.Client, conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

//TODO
func NewClientWithCodec(newClientWithCodec func(rpc.ClientCodec) *rpc.Client, codec rpc.ClientCodec) *rpc.Client {
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
	C      io.Closer
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
}

func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	encBuf := bufio.NewWriter(conn)
	return &ClientClockCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
}

func (c *ClientClockCodec) WriteRequest(req *rpc.Request, param interface{}) (err error) {
	fmt.Println("WriteRequest")
	if err = c.Enc.Encode(req); err != nil {
		return
	}
	buf := dinvRT.Pack(param)
	if err = c.Enc.Encode(buf); err != nil {
		return
	}

	return c.EncBuf.Flush()
}

func (c *ClientClockCodec) ReadResponseHeader(resp *rpc.Response) error {
	fmt.Println("ReadResponseHeader")
	return c.Dec.Decode(resp)
}

func (c *ClientClockCodec) ReadResponseBody(body interface{}) (err error) {
	fmt.Println("ReadResponseBody")
	var buf []byte
	if err = c.Dec.Decode(&buf); err != nil {
		return
	}
	dinvRT.Unpack(buf, body)
	return nil
}

func (c *ClientClockCodec) Close() error {
	return c.C.Close()
}

/*
	Client Codec for instrumenting RPC calls with vector clocks
*/

type ServerClockCodec struct {
	Rwc    io.ReadWriteCloser
	Dec    *gob.Decoder
	Enc    *gob.Encoder
	EncBuf *bufio.Writer
	Closed bool
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	buf := bufio.NewWriter(conn)
	srv := &ServerClockCodec{
		Rwc:    conn,
		Dec:    gob.NewDecoder(conn),
		Enc:    gob.NewEncoder(buf),
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
	dinvRT.Unpack(buf, body)
	return nil

}
func (c *ServerClockCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.Enc.Encode(r); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			fmt.Println("RPC Error encoding response:", err)
			c.Close()
		}
		return
	}
	buf := dinvRT.Pack(body)
	if err = c.Enc.Encode(buf); err != nil {
		if c.EncBuf.Flush() == nil {
			//Gob Encoding Error
			fmt.Println("RPC Error encoding body (This is likely Dinv's fault):", err)
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

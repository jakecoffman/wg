package wg

import "net/http"

type FakeConn struct {
	Req *http.Request
	Closed bool
}

func (c *FakeConn) Send(v interface{}) {

}

func (c *FakeConn) Recv(v interface{}) error {
	return nil
}

func (c *FakeConn) Close() error {
	c.Closed = true
	return nil
}

func (c *FakeConn) Request() *http.Request {
	return c.Req
}

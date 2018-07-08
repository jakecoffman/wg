package wg

import "net/http"

type FakeConn struct {
	FakeIp string
	Closed bool

	Msgs chan interface{}
}

func NewFakeConn(ip string) *FakeConn {
	return &FakeConn{FakeIp: ip, Msgs: make(chan interface{}, 1000)}
}

func (c *FakeConn) Send(v interface{}) {
	c.Msgs <- v
}

func (c *FakeConn) Recv(v interface{}) error {
	return nil
}

func (c *FakeConn) SendRaw(v []byte) {

}

func (c *FakeConn) RecvRaw(v []byte) error {
	return nil
}

func (c *FakeConn) Close() error {
	c.Closed = true
	return nil
}

func (c *FakeConn) Ip() string {
	return c.FakeIp
}

func (c *FakeConn) Cookie(name string) (*http.Cookie, error) {
	return &http.Cookie{}, nil
}

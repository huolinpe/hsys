package gredis

import (
	"bufio"
	"context"
	"net"
)

type Conn struct {
	// usedAt  int64 // atomic
	netConn net.Conn

	br *bufio.Reader
	bw *bufio.Writer
	// rd *proto.Reader
	// bw *bufio.Writer
	// wr *proto.Writer

	// Inited    bool
	// pooled    bool
	// createdAt time.Time
}

func NewConn(netConn net.Conn) *Conn {
	cn := &Conn{
		netConn: netConn,
	}
	cn.br = bufio.NewReader(netConn)
	cn.bw = bufio.NewWriter(netConn)
	return cn
}

func (cn *Conn) WithWriter(ctx context.Context, fn func(wr *bufio.Writer) error) error {
	 fn(cn.bw)
	 cn.bw.Flush()
	 return nil
}

func (cn *Conn) WithReader(ctx context.Context, fn func(wr *bufio.Reader) error) error {
	return fn(cn.br)
}

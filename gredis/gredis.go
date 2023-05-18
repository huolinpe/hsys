package gredis

import (
	"net"
	"time"
)

type Client struct {
	*Conn
	cmdable
}

// type baseClient struct {
// 	connPool pool.Pooler
// }

func NewClient(network, addr string) *Client {
	netDialer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 5 * time.Minute,
	}

	rwc, err := netDialer.Dial(network, addr)
	if err != nil {
		return nil
	}

	conn := NewConn(rwc)
	cmdfunc := NewCmdfunc(conn)

	return &Client{
		Conn:    conn,
		cmdable: cmdfunc,
	}
}

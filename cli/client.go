package main

import (
	"log"
	"net"
	"time"

	hrpc "hurpc/rpc"
)

var remote_ip string = "9.134.*.*"
var remote_port string = "7080"
var rs *hrpc.RemoteStream = new(hrpc.RemoteStream)

func main() {
	for {
		log.Println("connect to server start!")
		conn, err := net.DialTimeout("tcp", remote_ip+":"+remote_port, 5*time.Second)
		if err != nil {

		}

		rs.InitConn(conn, 10, 64*1024, 64*1024)
		// rs.Handlers["getRemoteIp"] = (hrpc.RPCHandler)GetRemoteIp
		rs.Run()

		time.Sleep(50 * time.Second)
	}
}

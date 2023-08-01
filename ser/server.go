package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	hrpc "hurpc/rpc"
)

var listenPort int = 7080

// type AgentMgr struct {
// 	listen int
// 	// configs         []agentConfig
// 	// server          *webServer

// 	// supportCommands commandSupport
// 	agentTypeDesc map[string]string
// }

var allAgents map[string]*hrpc.RemoteStream = make(map[string]*hrpc.RemoteStream)

func main() {
	var listener *net.TCPListener
	var err error

	listenAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(listenPort))
	if err != nil {
		log.Println("Error listening:", err)
		os.Exit(1)
	}

	listener, err = net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Println("Error listening:", err)
		os.Exit(1)
	}

	defer listener.Close()
	log.Println("Listening at ", listenAddr)

	for {
		listener.SetDeadline(time.Now().Add(50 * time.Millisecond))
		conn, err := listener.Accept()
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				continue
			}
			os.Exit(1)
		}

		remoteAddr := conn.RemoteAddr()
		remoteIp := remoteAddr.(*net.TCPAddr).IP.String()

		// if _, ok := allAgents[remoteIp]; ok {
		// 	log.Printf("ip already in allAgents table: %s\n", remoteIp)
		// 	continue
		// }

		log.Printf("agent connected: %s\n", remoteIp)

		agent := new(hrpc.RemoteStream)
		agent.InitConn(conn, 10, 64*1024, 64*1024)
		// agent.Setup(m, *cfg, conn)
		// agent.loadContext()

		go agent.Run()

		go func() {
			var return_val string
			agent.Call(&return_val, "getRemoteIp") //服务器触发一次rpc调用
			log.Printf("rpc get remote ip:%s", return_val)
		}()

		allAgents[remoteIp] = agent
	}
}

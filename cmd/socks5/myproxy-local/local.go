package main

import (
	"fmt"
	"github.com/abusizhishen/myproxy/src"
	"log"
	"net"
)

func main() {
	proxy := src.Init()
	addr,err := net.ResolveTCPAddr("tcp",fmt.Sprintf("%s:%d",proxy.LocalAddr,proxy.LocalPort))
	if err != nil{
		panic(err)
	}

	listen,err := net.ListenTCP("tcp",addr)

	if err != nil{
		panic(err)
	}

	log.Printf("listen: %s",addr)

	for {
		userConn,err := listen.AcceptTCP()
		if err != nil{
			log.Printf("读取错误:%s",err)
			continue
		}


		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		userConn.SetLinger(0)
		go proxy.RemoteHandler(src.GetTCPConn(userConn))
	}
}
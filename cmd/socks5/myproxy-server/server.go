package main

import (
	"fmt"
	"github.com/abusizhishen/myproxy/src"
	"log"
	"net"
)

func main() {
	proxy := src.Init()
	addr,err := net.ResolveTCPAddr("tcp",fmt.Sprintf(":%d",proxy.RemotePort))
	if err != nil{
		panic(err)
	}

	s := addr.Network()
	log.Print(s)
	listen,err := net.ListenTCP("tcp",addr)
	if err != nil{
		panic(err)
	}
	defer listen.Close()
	log.Printf("listen: %s",addr)

	for {
		clientConn,err := listen.AcceptTCP()
		if err != nil{
			log.Printf("读取错误:%s",err)
			continue
		}

		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		clientConn.SetLinger(0)
//		log.Printf("new conn:%s", clientConn.RemoteAddr())
		go src.DoRequestAndReturn(src.GetTCPConn(clientConn))
	}
}

//func handler(conn net.Conn, p *src.Proxy)  {
//
//	defer conn.Close()
//	var b = make([]byte,255)
//
//	for {
//		count,err := conn.Read(b)
//		if err != nil{
//			if err != io.EOF{
//				log.Println(err)
//				return
//			}
//			return
//		}
//
//		go p.RemoteHandler(conn)
//		log.Printf("read: %s",b[:count])
//	}
//}

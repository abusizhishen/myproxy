package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/myproxy/src"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	config string
	proxy src.Proxy
)

func init() {
	log.SetFlags(log.Lshortfile)

	flag.StringVar(&config, "config", "config.json","配置文件")
	flag.Parse()

	f,err := os.Open(config)
	if err != nil{
		log.Printf("打开配置文件错误：%s",err)
		os.Exit(0)
	}
	b,err := ioutil.ReadAll(f)
	if err != nil{
		log.Printf("读取配置文件错误：%s",err)
		os.Exit(0)
	}
	err = json.Unmarshal(b,&proxy)
	if err != nil{
		log.Printf("配置文件解析错误：%s",err)
		os.Exit(0)
	}

	src.Coder.Init(proxy.Password)
	log.Printf("listen on:%d\npassword:%s",proxy.RemotePort,proxy.Password)
}

func main() {
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
		//clientConn.SetLinger(0)
		log.Printf("new conn:%s", clientConn.RemoteAddr())
		//var buf = make([]byte,256)
		//n,err := clientConn.Read(buf)
		//log.Printf("read len%d,content:%s,err:%v",n,buf[:n],err)
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

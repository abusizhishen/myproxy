package main

import (
	"encoding/json"
	"flag"
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

	proxy.ParsePwd()
	log.Printf("listen on:%s\npassword:%s",proxy.GetRemoteAddr(),proxy.Password)
}

func main() {
	addr,err := net.ResolveTCPAddr("tcp",proxy.GetRemoteAddr())
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
		conn,err := listen.Accept()
		if err != nil{
			log.Printf("读取错误:%s",err)
		}

		//go handler(nn, &proxy)
		log.Printf("new conn:%s", conn.RemoteAddr())
		//var b = make([]byte,256)
		//n,err := conn.Read(b)
		//log.Printf("[success]: read remote:content\n %s,len:%d",b[:n],n)
		go proxy.DoRequestAndReturn(&conn)
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

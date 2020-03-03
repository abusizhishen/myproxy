package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/abusizhishen/myproxy/src"
	"io/ioutil"
	"log"
	"net"
	"os"
)

//appid : 1001407
//appkey: b47fnVaxOoSIGhXZ
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
}

func main() {
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
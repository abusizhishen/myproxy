package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/myproxy/src"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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
}

func main() {
	addr,err := net.ResolveTCPAddr("tcp",fmt.Sprintf("%s:%d",proxy.LocalAddr,proxy.LocalPort))
	if err != nil{
		panic(err)
	}

	s := addr.Network()
	log.Print(s)
	listen,err := net.ListenTCP("tcp",addr)
	if err != nil{
		panic(err)
	}

	log.Printf("listen: %s",addr)

	for {
		conn,err := listen.Accept()
		if err != nil{
			log.Printf("读取错误:%s",err)
		}

		go proxy.RemoteHandler(&conn)
	}
}

func x()  {
	http.Handle("/test/", http.FileServer(http.Dir("/home/work/"))) ///home/work/test/中必须有内容
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir("/home/work/"))))
	http.Handle("/tmpfiles/", http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp")))) //127.0.0.1:9999/tmpfiles/访问的本地文件/tmp中的内容
	http.ListenAndServe(":9999", nil)
}



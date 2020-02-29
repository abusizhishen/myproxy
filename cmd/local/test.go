package main

import (
	"flag"
	"github.com/myproxy/src"
	"log"
	"os"
)

func main() {
	var pwd,str string
	flag.StringVar(&pwd,"pwd","","加密密码")
	flag.StringVar(&str,"str","","加密字符串")
	flag.Parse()

	if pwd == "" {
		log.Printf("密码为空")
		os.Exit(0)
	}

	var proxy = src.Proxy{
		Local:    nil,
		Server:   nil,
		Password: pwd,
	}

	log.Println(src.StrToByte256(pwd))
	proxy.ParsePwd()
	data := []byte(str)
	proxy.Encode(data)
	log.Println("encode:",string(data))
	proxy.Decode(data)
	//
	log.Println("decode:",string(data))
}

package main

import (
	"flag"
	"github.com/abusizhishen/myproxy/src"
	"log"
)

func main() {
	var pwd,str string
	flag.StringVar(&pwd,"pwd","","加密密码")
	flag.StringVar(&str,"str","","加密字符串")
	flag.Parse()

	if pwd == "" {
		//log.Printf("密码为空")
		//os.Exit(0)
	}

	var proxy = src.Proxy{Password: pwd}

	log.Println(src.StrToByte256(pwd))
	proxy.GeneratePwd()
	data := []byte(str)
	proxy.Encode(data)
	log.Println("encode:",string(data))
	proxy.Decode(data)
	//
	log.Println("decode:",string(data))
}

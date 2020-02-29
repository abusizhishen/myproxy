package main

import (
	"encoding/base64"
	"flag"
	"log"
	"math/rand"
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

	b := rand.Perm()
	s := []byte(str)
	var m = make(map[int]int)
	for i:=0;i<256;i++{
		m[i] = i
	}

	var i int
	for _,v := range m{
		b[i] = byte(v)
		i++
	}

	log.Printf("key:%v",b)
	log.Printf("before encode:%v",s)
	base64.StdEncoding.Encode(b,s)
	log.Printf("after encode:%v",s)
	base64.StdEncoding.Encode(b,s)
	log.Printf("decode:%s",s)
}

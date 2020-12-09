package src

import (
	base "encoding/base64"
	"flag"
	"io/ioutil"
	"os"
	"log"
	"encoding/json"
	"path"
)

type Code struct {
	encode *password
	decode *password
}

func (c *Code)Encode(b []byte)  {
	for i,v := range b{
		b[i] = c.encode[v]
	}
}

func (c *Code)Decode(b []byte)  {
	for i,v := range b{
		b[i] = c.decode[v]
	}
}

func (c *Code)Init(passwd string)  {
	b,err := base.StdEncoding.DecodeString(passwd)
	if err != nil{
		log.Printf("密码解析错误：%s, passwd:%s",err,passwd)
		os.Exit(0)
	}

	c.decode = &password{}
	c.encode = &password{}
	for i,v := range b{
		c.encode[i] = v
		c.decode[v] = byte(i)
	}

	//log.Printf("password:\n%v,encode\n%v,decode\n%v",b,c.encode,c.decode)
}

var Coder = Code{
	encode: nil,
	decode: nil,
}

func Decode(data []byte)  {
	Coder.Decode(data)
}

func Encode(data []byte)  {
	Coder.Encode(data)
}

func Init() Proxy{
	var proxy Proxy
	var config string
	log.SetFlags(log.Lshortfile)
	homeDir,_ := os.UserHomeDir()
	defaultConfig := path.Join(homeDir,"myproxy.json")
	flag.StringVar(&config, "config", defaultConfig,"配置文件")
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

	Coder.Init(proxy.Password)
	return proxy
}

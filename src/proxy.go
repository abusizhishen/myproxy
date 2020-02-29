package src

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	//"encoding/base64"
	base "encoding/base64"
	"log"
	"math/rand"
	"net"
)

type Proxy struct {
	Local *net.TCPAddr
	Server *net.TCPAddr
	LocalAddr string `json:"local_addr"`
	LocalPort int16 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort int16 `json:"remote_port"`
	//Method string
	Password string `json:"password"`
	encodeKey password
	decodeKey password
}

func (p *Proxy)Decode(data []byte)  {
	for i,v := range data{
		data[i] = p.decodeKey[v]
	}
}

func (p *Proxy)Encode(data []byte)  {
	for i,v := range data{
		data[i] = p.encodeKey[v]
	}
}

func (p *Proxy)GeneratePwd() {
	var randInt = rand.Perm(256)
	var b []byte
	for _,v := range randInt{
		b = append(b,byte(v))
	}
	key := base.StdEncoding.EncodeToString(b)
	p.Password = key
	log.Println(base.StdEncoding.DecodeString(key))
	log.Printf("密码: %v",key)
	p.decodeKey = password{}
	p.encodeKey = password{}
	for i:=0;i<len(b);i++{
		v := b[i]
		p.encodeKey[i] = byte(v)
		p.decodeKey[v] = byte(i)
	}
}

func (p *Proxy)ParsePwd() {
	b,err := base.StdEncoding.DecodeString(p.Password)
	if err != nil{
		log.Printf("密码解析错误：%s, passwd:%s",err,p.Password)
		os.Exit(0)
	}
	p.decodeKey = password{}
	p.encodeKey = password{}
	for i:=0;i<len(b);i++{
		v := b[i]
		p.encodeKey[i] = v
		p.decodeKey[v] = byte(i)
	}
}

type password [256]byte

func StrToByte256(string2 string) [32]byte {

	s := sha256.Sum256([]byte(string2))
	return s
}

var addr = "%s:%d"
func (p Proxy)GetRemoteAddr()string {
	return fmt.Sprintf(addr, p.RemoteAddr,p.RemotePort)
}

func (p Proxy)GetLocalAddr()string {
	return fmt.Sprintf(addr, p.LocalAddr,p.LocalPort)
}

func (p Proxy)RemoteHandler(conn *net.Conn)  {
	log.Println("remote close")
	defer (*conn).Close()
	connAddr,err := net.ResolveTCPAddr("tcp",p.GetRemoteAddr())
	if err != nil{
		log.Printf("connert remote add%s",err)
		return
	}

	log.Printf("创建远程连接")
	rmt,err := net.DialTCP("tcp",nil,connAddr)
	if err != nil{
		log.Printf("connert remote err%s",err)
		return
	}
	defer func() {
		log.Println("remote close")
	}()
	defer rmt.Close()

	go func() {
		var b = make([]byte,1024)
		for{
			n,err := (*conn).Read(b)
			if err != nil{
				if err == io.EOF{
					return
				}
				log.Printf("[err]: read client:err %s",err)
				return
			}


			//log.Printf("before :%s",b[:n])
			p.Encode(b[:n])
			//log.Printf("after :%s",b[:n])
			_,err = rmt.Write(b[:n])
			if err != nil{
				log.Printf("[err]: write remote:err %s",err)
				continue
			}
		}
	}()

	var b = make([]byte,1024)
	for{
		n,err := rmt.Read(b)
		if err != nil{
			if err == io.EOF{
				return
			}
			log.Printf("[err]: read server:err %s",err)
			return
		}

		p.Decode(b[:n])
		_,err = (*conn).Write(b[:n])
		if err != nil{
			log.Printf("[err]: write remote:err %s",err)
			continue
		}
	}
}

func (p Proxy)DoRequestAndReturn(conn *net.Conn)  {
	//connAddr,err := net.ResolveTCPAddr("tcp",p.GetRemoteAddr())
	//if err != nil{
	//	log.Printf("connert remote add%s",err)
	//	return
	//}
	//
	//rmt,err := net.DialTCP("tcp",nil,connAddr)
	//if err != nil{
	//	log.Printf("connert remote add%s",err)
	//	return
	//}
	//defer rmt.Close()

	go func() {
		defer func() {
			log.Println("read end close")
		}()
		defer (*conn).Close()
		var b = make([]byte,1024)
		for{
			n,err := (*conn).Read(b)
			if err != nil{
				if err == io.EOF{
					(*conn).Read([]byte("bye byeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"))
					return
				}
				log.Printf("[err]: read client :err %s",err)
				return
			}

			//type
			//var bys = bufio.NewReader()
			//http.ReadRequest(bys)
			p.Decode(b[:n])
			log.Printf("[success]: read remote:content\n %s",b[:n])
			n,err = (*conn).Write(b[:n])

			log.Printf("write client len:%d,err:%s",n,err)


			//_,err = rmt.Write(b)
			//if err != nil{
			//	log.Printf("[err]: write remote:err %s",err)
			//	continue
			//}
		}
	}()

	//var b = make([]byte,1024)
	//for{
	//	n,err := rmt.Read(b)
	//	if err != nil{
	//		if err == io.EOF{
	//			return
	//		}
	//		log.Printf("[err]: read real server:err %s",err)
	//		return
	//	}
	//
	//	p.Encode(b[:n])
	//	_,err = conn.Write(b)
	//	if err != nil{
	//		log.Printf("[err]: write remote:err %s",err)
	//		continue
	//	}
	//}
}

type Reader struct {
	Bytes []byte
}

//func (r Reader)Read(b []byte)([]byte,error)  {
//
//}
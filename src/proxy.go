package src

import (
	"crypto/sha256"
	"encoding/binary"
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

func (p Proxy)RemoteHandler(conn *TCPConn)  {
	defer conn.Close()
	connAddr,err := net.ResolveTCPAddr("tcp",p.GetRemoteAddr())
	if err != nil{
		log.Printf("connert remote add%s",err)
		return
	}

	log.Printf("创建远程连接")
	drs,err := net.DialTCP("tcp",nil,connAddr)
	if err != nil{
		log.Printf("connert remote err%s",err)
		return
	}

	defer drs.Close()
	rmt := &TCPConn{drs, &Coder,}

	go func() {
		var b = make([]byte,1024)
		for{
			n,err := conn.Read(b)
			if err != nil{
				if err == io.EOF{
					return
				}
				log.Printf("[err]: read client:err %s",err)
				return
			}

			log.Printf("content:%s",b[:n])
			_,err = rmt.EncodeWrite(b[:n])
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

		log.Printf("remote receive:%s",b[:n])
		_,err = conn.DecodeWrite(b[:n])
		if err != nil{
			log.Printf("[err]: write remote:err %s",err)
			continue
		}
	}
}

func DoRequestAndReturn(conn *TCPConn)  {
	defer conn.Close()
	buf := make([]byte, 256)

	/**
	   The conn connects to the dstServer, and sends a ver
	   identifier/method selection message:
		          +----+----------+----------+
		          |VER | NMETHODS | METHODS  |
		          +----+----------+----------+
		          | 1  |    1     | 1 to 255 |
		          +----+----------+----------+
	   The VER field is set to X'05' for this ver of the protocol.  The
	   NMETHODS field contains the number of method identifier octets that
	   appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := conn.DecodeRead(buf)

	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		log.Printf(" err != nil || buf[0] != 0x05 %s,bug[0]:%d",err,buf[0])
		log.Printf("%d,%d",conn.encode[buf[0]],conn.decode[buf[0]])
		log.Printf("%v\n%v",conn.encode,conn.decode)
		return
	}

	/**
	   The dstServer selects from one of the methods given in METHODS, and
	   sends a METHOD selection message:

		          +----+--------+
		          |VER | METHOD |
		          +----+--------+
		          | 1  |   1    |
		          +----+--------+
	*/
	// 不需要验证，直接验证通过
	conn.EncodeWrite([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	n, err := conn.DecodeRead(buf)
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

	// 连接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		return
	} else {
		defer dstServer.Close()
		// Conn被关闭时直接清除所有数据 不管没有发送的数据
		dstServer.SetLinger(0)

		// 响应客户端连接成功
		/**
		  +----+-----+-------+------+----------+----------+
		  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		  +----+-----+-------+------+----------+----------+
		  | 1  |  1  | X'00' |  1   | Variable |    2     |
		  +----+-----+-------+------+----------+----------+
		*/
		// 响应客户端连接成功
		conn.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	rmt := &TCPConn{dstServer, &Coder,
	}
	conn.Write([]byte{0x05, 0x00})

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
					//(*conn).Read([]byte("bye byeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"))
					return
				}
				log.Printf("[err]: read client :err %s",err)
				return
			}

			if n > 0 {
				_,err = rmt.DecodeWrite(b[:n])
				if err != nil{
					log.Printf("write remote err:%s",err)
					return
				}
			}
		}
	}()

	var b = make([]byte,1024)
	for{
		n,err := rmt.Read(b)
		if err != nil{
			if err == io.EOF{
				log.Printf("rmt read end")
				return
			}
			log.Printf("[err]: read real server:err %s",err)
			return
		}

		if n < 0 {
			return
		}
		log.Printf("get resp: %s",b[:n])
		_,err = conn.EncodeWrite(b[:n])
		if err != nil{
			log.Printf("[err]: write remote:err %s",err)
			return
		}
	}
}


type TCPConn struct {
	io.ReadWriteCloser
	*Code
}

func (t *TCPConn)EncodeWrite(data []byte)(int,error)  {
	t.Encode(data)
	return t.Write(data)
}

func (t *TCPConn)DecodeWrite(data []byte)(int,error)  {
	t.Decode(data)
	return t.Write(data)
}

func (t *TCPConn)DecodeRead(data []byte) (n int,err error) {
	n,err = t.Read(data)
	if err != nil{
		return
	}
	t.Decode(data[:n])

	return
}

func GetTCPConn(conn *net.TCPConn) *TCPConn {
	return &TCPConn{conn, &Coder,
	}
}

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
		b[i] = c.encode[v]
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
	for i:=0;i<len(b);i++{
		v := b[i]
		c.encode[i] = v
		c.decode[v] = byte(i)
	}
}

var Coder = Code{
	encode: nil,
	decode: nil,
}
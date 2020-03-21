package src

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

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
	LocalPort int `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort int `json:"remote_port"`
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
func (p Proxy)GetRemoteAddr()(*TCPConn,error) {
	connAddr,err := net.ResolveTCPAddr("tcp",fmt.Sprintf(addr,p.RemoteAddr,p.RemotePort))
	if err != nil{
		return nil,fmt.Errorf("connert remote add%s",err)
	}

	dst,err := net.DialTCP("tcp",nil,connAddr)
	if err != nil{
		return nil,err
	}

	return GetTCPConn(dst),nil
}

func (p Proxy)GetTargetAddr(addr string)(*TCPConn,error) {
	connAddr,err := net.ResolveTCPAddr("tcp",addr)
	if err != nil{
		log.Printf("connert remote add%s",err)
		return nil,err
	}

	drs,err := net.DialTCP("tcp",nil,connAddr)
	if err != nil{
		log.Printf("connert remote err%s",err)
		return nil,err
	}

	return GetTCPConn(drs),nil
}

func (p Proxy)GetLocalAddr()string {
	return fmt.Sprintf(addr, p.LocalAddr,p.LocalPort)
}

func (p Proxy)RemoteHandler(userConn *TCPConn)  {
	defer userConn.Close()
	drs,err := p.GetRemoteAddr()
	if err != nil{
		log.Printf("connert remote err%s",err)
		return
	}

	//log.Printf("远程连接成功：%v")
	defer drs.Close()
	proxy := &TCPConn{drs, &Coder,}

	//log.Printf("异步读取客户端数据")
	go func() {
		proxy.CopyTo(userConn,OperateDecode)
	}()

	//log.Printf("客户端数据写入远程代理")
	err = userConn.CopyTo(proxy,OperateEncode)
}

func DoRequestAndReturn(clientConn *TCPConn)  {
	defer clientConn.Close()
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
	_, err := clientConn.DecodeRead(buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		log.Printf(" err %v,buf[0] :%d",err,buf[0])
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
	clientConn.EncodeWrite([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	n, err := clientConn.DecodeRead(buf)
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		log.Println("eeee")
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
		log.Printf("unknown buf[3]:%v",buf[3])
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

//	log.Printf("remote:%s",dstAddr)
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
		clientConn.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	rmt := &TCPConn{dstServer, &Coder,}

	go func() {
		rmt.CopyTo(clientConn,OperateEncode)
		//rmt.Close()
		//clientConn.Close()
	}()

	err = clientConn.CopyTo(rmt,OperateDecode)
	if err != nil{
		log.Printf("真是请求出错:%s",err)
	}
}


type TCPConn struct {
	io.ReadWriteCloser
	*Code
}

func (from *TCPConn)EncodeWrite(data []byte)(int,error)  {
	from.Encode(data)
	return from.Write(data)
}

func (from *TCPConn)DecodeWrite(data []byte)(int,error)  {
	from.Decode(data)
	return from.Write(data)
}

func (from *TCPConn)DecodeRead(data []byte) (n int,err error) {
	n,err = from.Read(data)
	if err != nil{
		return
	}
//	log.Printf("%v",data)
	from.Decode(data[:n])
	return
}

func GetTCPConn(conn *net.TCPConn) *TCPConn {
	return &TCPConn{conn, &Coder,}
}

const (
	OperateDefault = 0
	OperateEncode = 1
	OperateDecode = 2
)

func (from *TCPConn)CopyTo(to *TCPConn,operateType int)error  {
	var b = make([]byte,1024)

	for{
		readCount,err := from.Read(b)
		if err != nil{
			if err == io.EOF{
				return nil
			}
			return fmt.Errorf("[err]: read err to err %s",err)
		}

		if readCount > 0 {
			var writeCount int
			var err error
			switch operateType {
			case OperateDefault:
				writeCount,err = to.Write(b[:readCount])
			case OperateEncode:
				writeCount,err = to.EncodeWrite(b[:readCount])
			case OperateDecode:
				writeCount,err = to.DecodeWrite(b[:readCount])
			}
			if err != nil{
				return fmt.Errorf("[err]: write remote:err %s",err)
			}

			if readCount != writeCount{
				return fmt.Errorf("[err]: write remote读写不一致")
			}
		}
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
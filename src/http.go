package src

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/url"
	"strings"
)

type HttpProxy struct {
	Addr
}

func (p *HttpProxy)Run(fun func(clientConn *net.TCPConn),port int) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}

		go fun(conn)
	}
}

type Addr struct {
	LocalPort,ServerPort int
	ServerIp string
}

func NewHttpServer(addr Addr)  {
	var proxy = HttpProxy{addr}
	proxy.Run(proxy.serverHandle, addr.ServerPort)
}

func NewHttpClient(addr Addr)  {
	var proxy = HttpProxy{addr}
	proxy.Run(proxy.clientHandle, addr.LocalPort)
}

func HandlerSocks(clientConn net.TCPConn) {
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
	_, err := clientConn.Read(buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		log.Printf(" err %v,buf[0] :%d", err, buf[0])
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
	n, err := clientConn.Write([]byte{0x05, 0x00})
	log.Printf("len,%d,err:%v", n, err)

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	n, err = clientConn.Read(buf)
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
		log.Printf("unknown buf[3]:%v", buf[3])
		return
	}
	dPort := buf[n-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

	log.Println(dstAddr)
	buf = make([]byte, 1024)
	for {
		var n, err = clientConn.Read(buf)
		if err != nil {
			return
		}

		log.Printf("read content\n%s", buf[:n])
	}
}

func (p *HttpProxy)serverHandle(client *net.TCPConn) {
	defer client.Close()
	log.Printf("remote addr: %v\n", client.RemoteAddr())

	// 用来存放客户端数据的缓冲区
	var b [1024]byte
	//从客户端获取数据
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(b[:n]))
	var method, URL, address string
	// 从客户端数据读入method，url
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &URL)
	hostPortURL, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("url:%s\n,method:%s\n,address:%s\n", URL, method, address)
	// 如果方法是CONNECT，则为https协议
	if method == "CONNECT" {
		address = hostPortURL.Scheme + ":" + hostPortURL.Opaque
	} else { //否则为http协议
		address = hostPortURL.Host
		// 如果host不带端口，则默认为80
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			address = hostPortURL.Host + ":80"
		}
	}

	//获得了请求的host和port，向服务端发起tcp连接
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()

	//如果使用https协议，需先向客户端表示连接建立完毕
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else { //如果使用http协议，需将从客户端得到的http请求转发给服务端
		server.Write(b[:n])
	}

	//将客户端的请求转发至服务端，将服务端的响应转发给客户端。io.Copy为阻塞函数，文件描述符不关闭就不停止
	go io.Copy(server, client)
	io.Copy(client, server)
}
func (p *HttpProxy)clientHandle(client *net.TCPConn) {
	defer client.Close()
	log.Printf("remote addr: %v\n", client.RemoteAddr())
	//获得了请求的host和port，向服务端发起tcp连接
	server, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.ServerIp,p.ServerPort))
	if err != nil {
		log.Println(err)
		return
	}

	defer server.Close()
	go io.Copy(server, client)
	io.Copy(client, server)
}
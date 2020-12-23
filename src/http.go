package src

import (
	"bytes"
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
	addrStr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Infof("listen:%v", addrStr)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
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
	newHostPort,ServerPort int
	ServerIp string
}

func NewHttpServer(addr Addr)  {
	var proxy = HttpProxy{addr}
	proxy.Run(proxy.serverHandle, addr.ServerPort)
}

//特殊服务器， 支持针对某个host进行替换
func NewSpecialHttpServer(addr Addr)  {
	var proxy = HttpProxy{addr}
	proxy.Run(proxy.specialServerHandle, addr.ServerPort)
}

func NewHttpClient(addr Addr)  {
	var proxy = HttpProxy{addr}
	proxy.Run(proxy.clientHandle, addr.newHostPort)
}

func (p *HttpProxy)serverHandle(client *net.TCPConn) {
	defer client.Close()
	log.Infof("remote addr: %v\n", client.RemoteAddr())

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

func (p *HttpProxy)specialServerHandle(client *net.TCPConn) {
	defer client.Close()
	log.Infof("remote addr: %v\n", client.RemoteAddr())

	// 用来存放客户端数据的缓冲区
	var b [1024]byte
	//从客户端获取数据
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	//log.Println("\n\n-----",string(b[:n]),"----\n\n")
	var method, URL, address string
	// 从客户端数据读入method，url
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &URL)
	hostPortURL, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return
	}

	// 如果方法是CONNECT，则为https协议

	// 如果host不带端口，则默认为80
	var oldHost string
	var newHost string

	if strings.Index(hostPortURL.Host,oldHost) != -1{
		hostPortURL.Host= strings.Replace(hostPortURL.Host, oldHost,newHost,1)
		count := strings.Count(string(b[:n]),oldHost)
		ss := strings.ReplaceAll(string(b[:n]), oldHost, newHost)
		n += (len(newHost)-len(oldHost))*count
		copy(b[:], ss)
	}

	address = hostPortURL.Host

	if method == "CONNECT" {
		address = hostPortURL.Scheme + ":" + hostPortURL.Opaque
	} else { //否则为http协议
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
package src

import (
	"bufio"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:9090")
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			continue
		}

		//go handlerSocks(conn)
		go handlerHttp(conn)
	}
}

func handlerSocks(clientConn *net.TCPConn) {
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

func handlerHttp(clientConn *net.TCPConn) {
	buf := make([]byte, 1024)
	defer clientConn.Close()
	for {
		var n, err = clientConn.Read(buf)
		if err != nil {
			return
		}

		var s = bufio.NewReader(strings.NewReader(string(buf[:n])))
		req, err := http.ReadRequest(s)
		if err != nil {
			log.Println(err)
			return
		}

		req.RequestURI = ""
		var c = &http.Client{}
		resp, err := c.Do(req)
		if err != nil {
			log.Println(err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		resp.Body.Close()
		clientConn.Write(body)
		break
	}
}

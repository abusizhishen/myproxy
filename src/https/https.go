package https

import (
	"bufio"
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/autotls"
	"log"
	"net"
	"net/http"
)

func Server()  {
	g := gin.New()
	g.GET("/", serverHandler)

	autotls.Run(g, "btmeeting.com")
}

func Local()  {
	tcpAddr,err := net.ResolveTCPAddr("tcp",":8004")
	if err != nil{
		panic(err)
	}

	listen,err := net.ListenTCP("tcp", tcpAddr)
	if err != nil{
		panic(err)
	}

	for {
		clientConn,err := listen.AcceptTCP()
		if err != nil{
			continue
		}

		go localHandler(clientConn)
	}
}

type HttpSProxy struct {

}

func localHandler(clientConn *net.TCPConn)  {
	defer clientConn.Close()
	var ch = make(chan []byte,1024)

	go func() {
		var b = make([]byte,1024)
		for {
			n,err := clientConn.Read(b)
			if err != nil{
				break
			}

			ch <- b[:n]
		}
		close(ch)
	}()

	for data := range ch{
		var url = "http://btmeeting.com:8005"
		req, err := http.NewRequest("GET", url, bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
		var c = &http.Client{}
		resp,err := c.Do(req)
		if err != nil{
			log.Println(err)
			break
		}

		resp.Write(clientConn)
		resp.Body.Close()
	}
}

func serverHandler(c *gin.Context)  {
	b := bufio.NewReader(c.Request.Body)
	req,err := http.ReadRequest(b)
	if err != nil{
		log.Printf("read request from client err: %s",err)
		return
	}

	defer c.Request.Body.Close()
	req.RequestURI = ""
	resp,err := (&http.Client{}).Do(req)
	if err != nil{
		log.Printf("read resp from read server err: %s",err)
		return
	}

	resp.Write(c.Writer)
}
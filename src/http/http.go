package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

func Server()  {
	g := gin.New()
	g.GET("/", serverHandler)

	log.Printf("config %+v",*config)
	g.Run(config.GetServerListenAddr())
}

func Local()  {
	listen,err := net.Listen("tcp", config.GetLocalUrl())
	if err != nil{
		panic(err)
	}

	for {
		clientConn,err := listen.Accept()
		if err != nil{
			continue
		}

		go localHandler(clientConn)
	}
}

type HttpSProxy struct {

}

func localHandler(clientConn net.Conn)  {
	defer clientConn.Close()
	//var ch = make(chan []byte,1024)

	tp := textproto.NewReader(bufio.NewReader(clientConn))
	_,isHttps,_,host := readRequest(tp)
	var err error
	if isHttps {
		// if https, should sent 200 to client
		_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			log.Printf("errr: %s",err)
			return
		}
	}

	log.Printf("host: %s",host)
	var b bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		//var b = make([]byte,1024)

		for {
			log.Println("read")
			n,err := tp.ReadLineBytes()
			if err != nil{
				log.Println(err)
				break
			}

			log.Println("read dd ")
			b.Write(n)
			//log.Println("read cc")
			//if err != nil{
			//	break
			//}
			//
			//ch <- b[:n]
			//req,err := http.ReadRequest(bufio.NewReader(bytes.NewReader(b[:n])))
			//log.Printf("data: %s",b[:n])
			//if err == nil{
			//	log.Printf("method: %s",req.Method)
			//	log.Printf("request: %s",req.URL)
			//	req.Body.Close()
			//}else{
			//	log.Println(err)
			//}
		}
		//close(ch)
		wg.Done()
	}()

	wg.Wait()
	log.Println("waiting")
	//for data := range ch{
	//	log.Println("read data")
	//	//var url = config.GetServerUrl()
	//	//req, err := http.NewRequest("GET", url, bytes.NewReader(data))
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//	//req.Header.Set("Content-Type", "application/json")
	//	//var c = &http.Client{
	//	//	Transport:     nil,
	//	//	CheckRedirect: nil,
	//	//	Jar:           nil,
	//	//	//Timeout:       5,
	//	//}
	//
	//	//if config.Timeout > 0{
	//	//	c.Timeout = config.Timeout*time.Second
	//	//}
	//	//resp,err := c.Do(req)
	//	//if err != nil{
	//	//	log.Println(err)
	//	//	break
	//	//}
	//	//
	//	//resp.Write(clientConn)
	//	//resp.Body.Close()
	//	b.Write(data)
	//}

	var url = config.GetServerUrl()
	log.Println("read all before")
	//data,err := ioutil.ReadAll(clientConn)
	//if err != nil {
	//	panic(err)
	//}
	log.Println("read all after")
	req, err := http.NewRequest("GET", url, bytes.NewReader(b.Bytes()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("remote_host", host)
	var c = &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		//Timeout:       5,
	}

	//if config.Timeout > 0{
	//	c.Timeout = config.Timeout*time.Second
	//}
	log.Println("send:")
	resp,err := c.Do(req)
	if err != nil{
		log.Println(err)
	}

	log.Println("receive:")
	resp.Write(clientConn)
	resp.Body.Close()

	//go func() {
	//	bufio.NewReader(clientConn).WriteTo()
	//}()
	//
	//io.Copy()
}

func serverHandler(c *gin.Context)  {
	log.Println("method:", c.Request.Method)
	b := bufio.NewReader(c.Request.Body)
	req,err := http.ReadRequest(b)
	if err != nil{
		log.Printf("read request from client err: %s",err)
		return
	}

	defer c.Request.Body.Close()
	log.Printf("remote: %s", req.RequestURI)
	checkGinRequest(req)
	req.RequestURI = ""
	resp,err := (&http.Client{}).Do(req)
	if err != nil{
		log.Printf("read resp from read server err: %s",err)
		return
	}

	log.Println("send:")
	resp.Write(c.Writer)
}

type Config struct {
	ServerAddr string `json:"server_addr"`
	ServerPort int16 `json:"server_port"`
	LocalAddr string `json:"local_addr"`
	LocalPort int16 `json:"local_port"`
	Timeout time.Duration `json:"timeout"`
}

func (c Config)GetLocalUrl() string {
	return fmt.Sprintf("%s:%d",config.LocalAddr,config.LocalPort)
}

func (c Config)GetServerUrl() string {
	return fmt.Sprintf("http://%s:%d",config.ServerAddr,config.ServerPort)
}

func (c Config)GetServerListenAddr() string {
	return fmt.Sprintf(":%d",config.ServerPort)
}

var config  = &Config{}
func init(){
	var configPath string
	log.SetFlags(log.Lshortfile)
	homeDir,err := os.UserHomeDir()
	if err != nil{
		log.Printf("用户home目录获取出错:%s",err)
		os.Exit(0)
	}
	defaultConfig := path.Join(homeDir,"http_config.json")
	flag.StringVar(&configPath, "config", defaultConfig,"配置文件")
	flag.Parse()

	log.Printf("配置文件目录:%s",configPath)
	f,err := os.Open(configPath)
	if err != nil{
		log.Printf("打开配置文件错误：%s",err)
		os.Exit(0)
	}
	b,err := ioutil.ReadAll(f)
	if err != nil{
		log.Printf("读取配置文件错误：%s",err)
		os.Exit(0)
	}
	log.Println(string(b))
	err = json.Unmarshal(b,config)
	if err != nil{
		log.Printf("配置文件解析错误：%s",err)
		os.Exit(0)
	}
}

func checkGinRequest(req *http.Request)  {
	hostLen := len(req.Host)
	if hostLen > 3 && req.Host[hostLen-3:hostLen] == "443"{
		req.URL.Scheme = "https"
	}else{
		req.URL.Scheme = "http"
	}
}

func readRequest(tp *textproto.Reader) (rawReqHeader bytes.Buffer, isHttps bool, credential,host string) {
	var err error
	var requestLine string
	if requestLine, err = tp.ReadLine(); err != nil {
		return
	}

	method, requestURI, _, ok := parseRequestLine(requestLine)
	if !ok {
		err = &BadRequestError{"malformed HTTP request"}
		return
	}

	// https request
	if method == "CONNECT" {
		isHttps = true
		requestURI = "http://" + requestURI
	}

	// get remote host
	uriInfo, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return
	}

	// Subsequent lines: Key: value.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}

	credential = mimeHeader.Get("Proxy-Authorization")

	if uriInfo.Host == "" {
		host = mimeHeader.Get("Host")
	} else {
		if strings.Index(uriInfo.Host, ":") == -1 {
			host = uriInfo.Host + ":80"
		} else {
			host = uriInfo.Host
		}
	}

	// rebuild http request header
	rawReqHeader.WriteString(requestLine + "\r\n")
	for k, vs := range mimeHeader {
		for _, v := range vs {
			rawReqHeader.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	rawReqHeader.WriteString("\r\n")
	return
}

func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

type BadRequestError struct {
	what string
}

func (b *BadRequestError) Error() string {
	return b.what
}

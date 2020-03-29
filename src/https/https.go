package https

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

func Server()  {
	g := gin.New()
	g.GET("/", serverHandler)

	log.Printf("config %+v",*config)
	autotls.Run(g, config.Domain)
}

func Local()  {
	tcpAddr,err := net.ResolveTCPAddr("tcp",config.GetLocalUrl())
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
			req,err := http.ReadRequest(bufio.NewReader(bytes.NewReader(b[:n])))
			if err == nil{
				log.Printf("request: %s",req.URL)
				req.Body.Close()
			}else{
				log.Println(err)
			}
		}
		close(ch)
	}()

	for data := range ch{
		var url = config.GetServerUrl()
		req, err := http.NewRequest("GET", url, bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
		var c = &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       5,
		}

		if config.Timeout > 0{
			c.Timeout = config.Timeout*time.Second
		}
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
	body,err := ioutil.ReadAll(c.Request.Body)
	log.Println(err)
	log.Printf("body:%s",body)
	return
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

	resp.Write(c.Writer)
}

type Config struct {
	Domain string `json:"domain"`
	LocalAddr string `json:"local_addr"`
	LocalPort int16 `json:"local_port"`
	Timeout time.Duration `json:"timeout"`
}

func (c Config)GetLocalUrl() string {
	return fmt.Sprintf("%s:%d",config.LocalAddr,config.LocalPort)
}

func (c Config)GetServerUrl() string {
	return fmt.Sprintf("https://%s",config.Domain)
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
	defaultConfig := path.Join(homeDir,"https_config.json")
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
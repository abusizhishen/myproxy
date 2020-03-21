package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	g := gin.New()
	g.POST("/", func(context *gin.Context) {
		body,err := HttpProxyDo(context.Request,"127.0.0.1:81","www.baidu.com")
		if err != nil{
			context.String(200,err.Error())
			return
		}
		context.String(200,string(body))
	})

	g.Run(":81")
}

func HttpProxyDo(req *http.Request, from,to string) (body []byte,err error) {
	log.Printf("%+v",req)
	log.Println()
	//log.Println(req.Host)
	//req.URL.Scheme = "http"
	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	log.Printf("%+v",req.URL.Scheme)
	log.Printf("origin:%s,to %s, find %s",req.URL.Host, to, strings.Replace(req.URL.Host,from, to,1))
	req.URL.Host = strings.Replace(req.URL.Host,from, to,1)
	req.Host = strings.Replace(req.Host, from, to,1)
	req.RequestURI = ""

	log.Printf("%+v",req)
	var c = &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	return
}
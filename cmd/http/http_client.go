package main

import (
	"github.com/abusizhishen/myproxy/src"
)

func main() {
	proxy := src.Init()
	src.NewHttpClient(src.Addr{
		LocalPort:  proxy.LocalPort,
		ServerPort: proxy.RemotePort,
		ServerIp:   proxy.RemoteAddr,
	})
}

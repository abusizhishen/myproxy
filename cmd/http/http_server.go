package main

import (
	"fmt"
	"github.com/abusizhishen/myproxy/src"
)

func main() {
	proxy := src.Init()
	src.NewHttpClient(fmt.Sprintf("%s:%d", proxy.RemoteAddr,proxy.RemotePort))
}

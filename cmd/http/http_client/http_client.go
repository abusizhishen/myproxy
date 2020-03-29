package main

import (
	"github.com/abusizhishen/myproxy/src/http"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//proxy()
	http.Local()
}

func proxy()  {
	g := gin.Default()
	g.GET("/", func(context *gin.Context) {
		log.Printf("%+v",context.Request)
	})

	g.Run(":8004")
}

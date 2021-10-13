package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/ping", func(c *gin.Context) {
		c.Status(200)
	})
	return r
}

func main() {
	r := setupRouter()
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

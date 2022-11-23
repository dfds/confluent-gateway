package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello world from confluent gateway!")

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

//1. Ensure capability has service account
//2. Ensure service account has all acls
//3. Ensure service account has api keys
//4. Ensure api keys are stored in vault
//5. Ensure topic is created
//6. Done!

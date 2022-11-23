package main

import (
	"fmt"
	"github.com/dfds/confluent-gateway/confluent"
	"github.com/gin-gonic/gin"
)

type newTopicHasBeenRequested struct {
	capabilityRootId string // example => logistics-somecapability-abcd
	clusterId        string
	topicName        string // full name => pub.logistics-somecapability-abcd.foo
	partitions       int
	retention        int // in ms
}

type Topic struct {
	name       string
	partitions int
	retention  int
}

type CapabilityRootId string

type ApiKey struct {
	userName string
	password string
}

type process struct {
	capabilityRootId      CapabilityRootId
	clusterId             confluent.ClusterId
	topic                 Topic
	serviceAccountId      confluent.ServiceAccountId
	acl                   confluent.Acl
	apiKey                ApiKey
	isApiKeyStoredInVault bool

	//1. Ensure capability has cluster access
	//  1.2. Ensure capability has service account
	//	1.3. Ensure service account has all acls
	//	1.4. Ensure service account has api keys
	//	1.5. Ensure api keys are stored in vault
	//2. Ensure topic is created
	//3. Done!
}

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

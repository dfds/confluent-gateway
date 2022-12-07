package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	r := gin.Default()

	r.POST("/kafka/v3/clusters/:cluster_id/topics", func(c *gin.Context) {
		c.Status(204)
	})

	r.POST("/iam/v2/service-accounts", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id": fmt.Sprintf("sa-%s", time.Now().Format("150405")),
		})
	})

	r.POST("/kafka/v3/clusters/:cluster_id/acls", func(c *gin.Context) {
		c.Status(204)
	})

	r.POST("/iam/v2/api-keys", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id": fmt.Sprintf("username-%s", time.Now().Format("150405")),
			"spec": gin.H{
				"secret": fmt.Sprintf("password-%s", time.Now().Format("150405")),
				"resource": gin.H{
					"id": "fake-cluster-id-1",
				},
			},
		})
	})

	r.POST("/aws-ssm-put", func(c *gin.Context) {

		c.Data(200, "application/x-amz-json-1.1", []byte(`
			{
			   "Tier": "Standard",
			   "Version": 1
			}
		`))
	})

	fmt.Println("Fake Confluent Cloud!")
	r.Run() // listen and serve on 0.0.0.0:8080

}

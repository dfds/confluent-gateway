package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const serviceAccountId = "sa-150405"

func main() {
	r := gin.Default()

	r.POST("/kafka/v3/clusters/:cluster_id/topics", func(c *gin.Context) {
		c.Status(204)
	})

	r.DELETE("/kafka/v3/clusters/:cluster_id/topics/:topic_name", func(c *gin.Context) {
		c.Status(204)
	})

	r.POST("/iam/v2/service-accounts", func(c *gin.Context) {
		c.JSON(200, gin.H{
			//"id": fmt.Sprintf("sa-%s", time.Now().Format("150405")),
			"id": serviceAccountId,
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

	r.GET("/iam/v2/api-keys", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"api_version": "iam/v2",
			"kind":        "ApiKeyList",
			"metadata": gin.H{
				"total_size": 0,
			},
			"data": []gin.H{},
		})
	})

	r.GET("/api/service_accounts", func(c *gin.Context) {
		c.Data(200, "application/json", []byte(`{
"users": [
	{
	  "id": 7482,
	  "email": "0-devex-deploy@serviceaccounts.confluent.cloud",
	  "first_name": "",
	  "last_name": "",
	  "deactivated": false,
	  "verified": "1970-01-01T00:00:00Z",
	  "created": "2019-06-27T08:47:35.460259Z",
	  "modified": "2019-06-27T08:47:35.460259Z",
	  "service_name": "devex-deploy",
	  "service_description": "Development excellence deploy account for management",
	  "service_account": true,
	  "preferences": {},
	  "internal": false,
	  "resource_id": "`+serviceAccountId+`",
	  "deactivated_at": null,
	  "social_connection": "",
	  "auth_type": "AUTH_TYPE_UNKNOWN"
	}]
}`))
	})

	r.POST("/subjects/:subject/versions", func(c *gin.Context) {
		c.Data(200, "application/vnd.schemaregistry.v1+json", []byte(`
			{
			   "id":  100001
			}
		`))
	})

	r.POST("/aws-ssm-put", func(c *gin.Context) {

		c.Data(200, "application/x-amz-json-1.1", []byte(`
			{
			   "Tier": "Standard",
			   "Version": 1
			}
		`))
	})

	// [THFIS] trailing slash (!!) because of AWS SDK
	r.POST("/aws-ssm-put/", func(c *gin.Context) {

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

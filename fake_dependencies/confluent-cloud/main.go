package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const serviceAccountId = "sa-150405"
const clusterRegion = "eu-central-1"
const clusterId = "`lkc-4dzn8`"
const clusterName = "pkc-e8dzn"

func main() {
	r := gin.Default()

	r.GET("/kafka/v3/clusters/:cluster_id/topics", func(c *gin.Context) {
		c.Data(200, "application/json", []byte(`{
			{
				"kind": "KafkaTopicList",
				"metadata": {
					"self": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics",
					"next": null
				},
				"data": [
					{
					"kind": "KafkaTopic",
					"metadata": {
						"self": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-command",
						"resource_name": "crn:///kafka=`+clusterId+`/topic=_confluent-command"
					},
					"cluster_id": "`+clusterId+`",
					"topic_name": "_confluent-command",
					"is_internal": false,
					"replication_factor": 3,
					"partitions_count": 1,
					"partitions": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-command/partitions"
					},
					"configs": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-command/configs"
					},
					"partition_reassignments": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-command/partitions/-/reassignment"
					},
					"authorized_operations": []
					},
					{
					"kind": "KafkaTopic",
					"metadata": {
						"self": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog",
						"resource_name": "crn:///kafka=`+clusterId+`/topic=_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog"
					},
					"cluster_id": "`+clusterId+`",
					"topic_name": "_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog",
					"is_internal": false,
					"replication_factor": 3,
					"partitions_count": 3,
					"partitions": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog/partitions"
					},
					"configs": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog/configs"
					},
					"partition_reassignments": {
						"related": "https://`+clusterName+`.`+clusterRegion+`.aws.confluent.cloud/kafka/v3/clusters/`+clusterId+`/topics/_confluent-controlcenter-5-1-2-1-AlertHistoryStore-changelog/partitions/-/reassignment"
					},
					"authorized_operations": []
					}
				]
			}
		}`))

	})

	r.POST("/kafka/v3/clusters/:cluster_id/topics", func(c *gin.Context) {
		c.Status(204)
	})

	r.DELETE("/kafka/v3/clusters/:cluster_id/topics/:topic_name", func(c *gin.Context) {
		c.Status(204)
	})

	r.DELETE("/subjects/:subject/versions/:version", func(c *gin.Context) {
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

	r.POST("/iam/v2/role-bindings", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id": "fake-role-binding-id",
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

	r.GET("/cluster/:clusterId/schemas", func(c *gin.Context) {
		subjectPrefix := c.Query("subjectPrefix")

		if subjectPrefix == "" {
			c.Data(200, "application/json", []byte(`
				[
					{
						"subject": "asdasd-lala",
						"version": 1,
						"id": 100022,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"$id\":\"http://example.com/myURI.schema.json\",\"title\":\"SampleRecord\",\"description\":\"Sample schema to help you get started.\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"myField1\":{\"type\":\"integer\",\"description\":\"The integer type is used for integral numbers.\"},\"myField2\":{\"type\":\"number\",\"description\":\"The number type is used for any numeric type, either integers or floating point numbers.\"},\"myField3\":{\"type\":\"string\",\"description\":\"The string type is used for strings of text.\"}}}"
					},
					{
						"subject": "cloudengineering.selfservice.test-EnvelopeOfAdmin",
						"version": 1,
						"id": 100019,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"title\":\"EnvelopeOfAdmin\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"messageId\":{\"type\":\"string\"},\"type\":{\"type\":\"string\"},\"data\":{\"$ref\":\"#/definitions/Admin\"}},\"definitions\":{\"Admin\":{\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"name\"],\"properties\":{\"name\":{\"type\":\"string\"},\"description\":{\"type\":\"string\"}}}}}"
					},
					{
						"subject": "cloudengineering.selfservice.test-EnvelopeOfAdmin",
						"version": 2,
						"id": 100020,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"additionalProperties\":false,\"definitions\":{\"Admin\":{\"additionalProperties\":false,\"properties\":{\"description\":{\"type\":\"string\"},\"name\":{\"type\":\"string\"}},\"required\":[\"name\"],\"type\":\"object\"}},\"properties\":{\"data\":{\"$ref\":\"#/definitions/Admin\"},\"messageId\":{\"type\":\"string\"},\"type\":{\"type\":\"string\",\"pattern\":\"admin\"}},\"title\":\"EnvelopeOfAdmin\",\"type\":\"object\"}"
					},
					{
						"subject": "cloudengineering.selfservice.test-EnvelopeOfUser",
						"version": 1,
						"id": 100018,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"title\":\"EnvelopeOfUser\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"messageId\":{\"type\":\"string\"},\"type\":{\"type\":\"string\"},\"data\":{\"$ref\":\"#/definitions/User\"}},\"definitions\":{\"User\":{\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"name\",\"favorite_color\"],\"properties\":{\"name\":{\"type\":\"string\"},\"favorite_color\":{\"type\":\"string\"},\"favorite_number\":{\"type\":\"integer\",\"format\":\"int64\"}}}}}"
					},
					{
						"subject": "test-01",
						"version": 1,
						"id": 100001,
						"schema": "{\"type\":\"record\",\"name\":\"sampleRecord\",\"namespace\":\"com.mycorp.mynamespace\",\"doc\":\"Sample schema to help you get started.\",\"fields\":[{\"name\":\"my_field1\",\"type\":\"int\",\"doc\":\"The int type is a 32-bit signed integer.\"},{\"name\":\"my_field2\",\"type\":\"double\",\"doc\":\"The double type is a double precision (64-bit) IEEE 754 floating-point number.\"},{\"name\":\"my_field3\",\"type\":\"string\",\"doc\":\"The string is a unicode character sequence.\"}]}"
					},
					{
						"subject": "x-value",
						"version": 1,
						"id": 100001,
						"schema": "{\"type\":\"record\",\"name\":\"sampleRecord\",\"namespace\":\"com.mycorp.mynamespace\",\"doc\":\"Sample schema to help you get started.\",\"fields\":[{\"name\":\"my_field1\",\"type\":\"int\",\"doc\":\"The int type is a 32-bit signed integer.\"},{\"name\":\"my_field2\",\"type\":\"double\",\"doc\":\"The double type is a double precision (64-bit) IEEE 754 floating-point number.\"},{\"name\":\"my_field3\",\"type\":\"string\",\"doc\":\"The string is a unicode character sequence.\"}]}"
					}
				]
			`))
		} else {
			c.Data(200, "application/json", []byte(`
				[
					{
						"subject": "`+subjectPrefix+`.asdasd-lala",
						"version": 1,
						"id": 100022,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"$id\":\"http://example.com/myURI.schema.json\",\"title\":\"SampleRecord\",\"description\":\"Sample schema to help you get started.\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"myField1\":{\"type\":\"integer\",\"description\":\"The integer type is used for integral numbers.\"},\"myField2\":{\"type\":\"number\",\"description\":\"The number type is used for any numeric type, either integers or floating point numbers.\"},\"myField3\":{\"type\":\"string\",\"description\":\"The string type is used for strings of text.\"}}}"
					},
					{
						"subject": "`+subjectPrefix+`.selfservice.test-EnvelopeOfAdmin",
						"version": 1,
						"id": 100019,
						"schemaType": "JSON",
						"schema": "{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"title\":\"EnvelopeOfAdmin\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"messageId\":{\"type\":\"string\"},\"type\":{\"type\":\"string\"},\"data\":{\"$ref\":\"#/definitions/Admin\"}},\"definitions\":{\"Admin\":{\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"name\"],\"properties\":{\"name\":{\"type\":\"string\"},\"description\":{\"type\":\"string\"}}}}}"
					}
				]
	
			`))
		}

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

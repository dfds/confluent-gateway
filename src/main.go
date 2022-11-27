package main

import (
	"github.com/dfds/confluent-gateway/database"
	"github.com/dfds/confluent-gateway/models"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {
	pr, err := database.NewProcessRepository(dsn)
	if err != nil {
		panic(err)
	}

	process := models.NewTopicCreationProcess(pr)

	process.ProcessLogic(models.NewTopicHasBeenRequested{
		CapabilityRootId: "new-cap",
		ClusterId:        "cluster-id",
		TopicName:        "new-topic",
		Partitions:       1,
		Retention:        -1,
	})

	//r := gin.Default()
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})
	//_ = r.Run() // listen and serve on 0.0.0.0:8080
}

package main

import (
	"github.com/dfds/confluent-gateway/logging"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {
	//db, err := database.NewDatabase(dsn)
	//if err != nil {
	//	panic(err)
	//}
	//
	//process := models.NewTopicCreationProcess(db, &mocks.MockClient{
	//	ServiceAccountId: "sa-some",
	//	ApiKey: models.ApiKey{
	//		Username: "u-some",
	//		Password: "p-another",
	//	},
	//}, &mocks.MockAwsClient{})
	//
	//if err := process.ProcessLogic(context.TODO(), models.NewTopicHasBeenRequested{
	//	CapabilityRootId: "new-cap",
	//	ClusterId:        "cluster-id",
	//	TopicName:        "new-topic",
	//	Partitions:       1,
	//	Retention:        -1,
	//}); err != nil {
	//	panic(err)
	//}

	//r := gin.Default()
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "pong",
	//	})
	//})
	//_ = r.Run() // listen and serve on 0.0.0.0:8080

	log := logging.NewLogger(logging.LoggerOptions{
		IsProduction: false,
		AppName:      "lala",
	})

	log.Trace("simple message")
	log.Debug("one reference {Something}", "foo")
}

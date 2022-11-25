package main

import (
	"context"
	"fmt"
	"github.com/dfds/confluent-gateway/database"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

const dsn = "host=localhost user=postgres password=p dbname=db port=5432 sslmode=disable"

func main() {

	Example_GetProcessById()
	Example_GetNextUncompletedProcess()
	Example_GetAllClusters()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	_ = r.Run() // listen and serve on 0.0.0.0:8080
}

func Example_GetProcessById() {
	pr, err := database.NewProcessRepository(dsn)
	if err != nil {
		panic(err)
	}

	process, err := pr.FindNextIncomplete(context.TODO())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n\n", process)
}

func Example_GetNextUncompletedProcess() {

	pr, err := database.NewProcessRepository(dsn)
	if err != nil {
		panic(err)
	}

	id, _ := uuid.FromString("4bf89980-fe38-46cb-833e-b07bc108d61b")

	process, err := pr.FindById(context.TODO(), id)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", process)
}

func Example_GetAllClusters() {
	cr, err := database.NewClusterRepository(dsn)
	if err != nil {
		panic(err)
	}

	clusters, err := cr.GetAll(context.TODO())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n\n", clusters)
}

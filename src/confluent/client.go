package confluent

import "fmt"

func CreateTopic(clusterId ClusterId, name string, partitions int, retention int) {
	if clusterId == nil {
		fmt.Println("cluster id is nil")
		return
	}

	fmt.Println("cluster id: ", clusterId)
}

func CreateServiceAccount(name string, description string) {

}

func CreateApiKey(clusterId string, serviceAccountId string) {

}

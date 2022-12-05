package models

import "context"

type ClusterId string

type Cluster struct {
	ClusterId         ClusterId `gorm:"column:id;primarykey"`
	Name              string
	AdminApiEndpoint  string
	AdminApiKey       ApiKey `gorm:"embedded;embeddedPrefix:admin_api_key_"`
	BootstrapEndpoint string
}

func (*Cluster) TableName() string {
	return "cluster"
}

type ClusterRepository interface {
	Get(ctx context.Context, id ClusterId) (Cluster, error)
}

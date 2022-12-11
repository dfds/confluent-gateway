package models

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

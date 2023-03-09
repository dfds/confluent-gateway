package models

type ClusterId string

type Cluster struct {
	ClusterId                 ClusterId `gorm:"column:id;primarykey"`
	Name                      string
	AdminApiEndpoint          string
	AdminApiKey               ApiKey `gorm:"embedded;embeddedPrefix:admin_api_key_"`
	BootstrapEndpoint         string
	SchemaRegistryApiEndpoint string
	SchemaRegistryApiKey      ApiKey `gorm:"embedded;embeddedPrefix:schema_registry_api_key_"`
}

func (*Cluster) TableName() string {
	return "cluster"
}

type ApiKey struct {
	Username string
	Password string
}

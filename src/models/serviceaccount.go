package models

import (
	"time"
)

type ServiceAccountId string

type ServiceAccount struct {
	Id               ServiceAccountId
	CapabilityRootId CapabilityRootId
	ClusterId        ClusterId
	ApiKey           ApiKey `gorm:"embedded;embeddedPrefix:api_key_"`
	//Acl              []AclEntry
	CreatedAt time.Time
}

func (*ServiceAccount) TableName() string {
	return "service_account"
}

type AclEntry struct {
	Id               int `gorm:"primarykey"`
	ServiceAccountId ServiceAccountId
	CreatedAt        *time.Time
	AclDefinition
}

func (*AclEntry) TableName() string {
	return "acl"
}

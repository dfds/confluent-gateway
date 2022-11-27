package models

import (
	"fmt"
	"strings"
)

type AclDefinition struct {
	ResourceType   ResourceType
	ResourceName   string
	PatternType    PatternType
	OperationType  OperationType
	PermissionType PermissionType
}

func (ad AclDefinition) String() string {
	return strings.Join([]string{
		string(ad.ResourceType),
		ad.ResourceName,
		string(ad.PatternType),
		string(ad.OperationType),
		string(ad.PermissionType),
	}, " | ")
}

func NewAclDefinition(resourceType ResourceType, resourceName string, patternType PatternType, operationType OperationType, permissionType PermissionType) AclDefinition {
	return AclDefinition{
		ResourceType:   resourceType,
		ResourceName:   resourceName,
		PatternType:    patternType,
		OperationType:  operationType,
		PermissionType: permissionType,
	}
}

type OperationType string

const (
	OperationTypeCreate          OperationType = "CREATE"
	OperationTypeRead                          = "READ"
	OperationTypeWrite                         = "WRITE"
	OperationTypeDescribe                      = "DESCRIBE"
	OperationTypeDescribeConfigs               = "DESCRIBECONFIGS"
	OperationTypeAlter                         = "ALTER"
	OperationTypeAlterConfigs                  = "ALTERCONFIGS"
	OperationTypeClusterAction                 = "CLUSTERACTION"
)

type PatternType string

const (
	PatternTypeLiteral PatternType = "LITERAL"
	PatternTypePrefix              = "PREFIX"
)

type ResourceType string

const (
	ResourceTypeTopic   ResourceType = "TOPIC"
	ResourceTypeGroup   ResourceType = "GROUP"
	ResourceTypeCluster ResourceType = "CLUSTER"
)

type PermissionType string

const (
	PermissionTypeDeny  PermissionType = "DENY"
	PermissionTypeAllow                = "ALLOW"
)

func CreateAclDefinitions(capabilityRootId CapabilityRootId) []AclDefinition {
	const PublicTopicPrefix = "pub."
	const AllTopics = "'*'"
	const ClusterResourceName = "kafka-cluster"

	pubPrefix := fmt.Sprintf("%s%s", PublicTopicPrefix, string(capabilityRootId))
	connectPrefix := fmt.Sprintf("connect-%s", string(capabilityRootId))

	return []AclDefinition{
		// deny create operations on all resource types
		CreateAclForTopicPrefix(AllTopics, OperationTypeCreate, PermissionTypeDeny),

		// for all private topics
		CreateAclForTopicPrefix(string(capabilityRootId), OperationTypeRead, PermissionTypeAllow),
		CreateAclForTopicPrefix(string(capabilityRootId), OperationTypeWrite, PermissionTypeAllow),
		CreateAclForTopicPrefix(string(capabilityRootId), OperationTypeCreate, PermissionTypeAllow),
		CreateAclForTopicPrefix(string(capabilityRootId), OperationTypeDescribe, PermissionTypeAllow),
		CreateAclForTopicPrefix(string(capabilityRootId), OperationTypeDescribeConfigs, PermissionTypeAllow),

		// for all public topics
		CreateAclForTopicPrefix(PublicTopicPrefix, OperationTypeRead, PermissionTypeAllow),

		// for own public topics
		CreateAclForTopicPrefix(pubPrefix, OperationTypeWrite, PermissionTypeAllow),
		CreateAclForTopicPrefix(pubPrefix, OperationTypeCreate, PermissionTypeAllow),

		// for all connect groups
		CreateAclForGroupPrefix(connectPrefix, OperationTypeRead, PermissionTypeAllow),
		CreateAclForGroupPrefix(connectPrefix, OperationTypeWrite, PermissionTypeAllow),
		CreateAclForGroupPrefix(connectPrefix, OperationTypeCreate, PermissionTypeAllow),

		// for all capability groups
		CreateAclForGroupPrefix(string(capabilityRootId), OperationTypeRead, PermissionTypeAllow),
		CreateAclForGroupPrefix(string(capabilityRootId), OperationTypeWrite, PermissionTypeAllow),
		CreateAclForGroupPrefix(string(capabilityRootId), OperationTypeCreate, PermissionTypeAllow),

		// for cluster
		NewAclDefinition(ResourceTypeCluster, ClusterResourceName, PatternTypeLiteral, OperationTypeAlter, PermissionTypeDeny),
		NewAclDefinition(ResourceTypeCluster, ClusterResourceName, PatternTypeLiteral, OperationTypeAlterConfigs, PermissionTypeDeny),
		NewAclDefinition(ResourceTypeCluster, ClusterResourceName, PatternTypeLiteral, OperationTypeClusterAction, PermissionTypeDeny),
	}
}

func CreateAclForTopicPrefix(topicName string, operationType OperationType, permissionType PermissionType) AclDefinition {
	return NewAclDefinition(ResourceTypeTopic, topicName, PatternTypePrefix, operationType, permissionType)
}

func CreateAclForGroupPrefix(groupName string, operationType OperationType, permissionType PermissionType) AclDefinition {
	return NewAclDefinition(ResourceTypeGroup, groupName, PatternTypePrefix, operationType, permissionType)
}

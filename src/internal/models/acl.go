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

type OperationType string

const (
	OperationTypeCreate          OperationType = "CREATE"
	OperationTypeRead                          = "READ"
	OperationTypeWrite                         = "WRITE"
	OperationTypeDescribe                      = "DESCRIBE"
	OperationTypeDescribeConfigs               = "DESCRIBE_CONFIGS"
	OperationTypeAlter                         = "ALTER"
	OperationTypeAlterConfigs                  = "ALTER_CONFIGS"
	OperationTypeClusterAction                 = "CLUSTER_ACTION"
)

type PatternType string

const (
	PatternTypeLiteral PatternType = "LITERAL"
	PatternTypePrefix              = "PREFIXED"
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

func CreateAclDefinitions(capabilityId CapabilityId) []AclDefinition {
	const publicTopicPrefix = "pub."
	const allTopics = "'*'"
	const clusterResourceName = "kafka-cluster"

	capabilityPrefix := string(capabilityId)
	publicCapabilityPrefix := fmt.Sprintf("%s%s", publicTopicPrefix, capabilityPrefix)
	connectPrefix := fmt.Sprintf("connect-%s", capabilityPrefix)

	return []AclDefinition{
		// deny create operations on all resource types
		defineAcl(ResourceTypeTopic, allTopics, PatternTypePrefix, OperationTypeCreate, PermissionTypeDeny),

		// for all private topics
		defineAcl(ResourceTypeTopic, capabilityPrefix, PatternTypePrefix, OperationTypeRead, PermissionTypeAllow),
		defineAcl(ResourceTypeTopic, capabilityPrefix, PatternTypePrefix, OperationTypeWrite, PermissionTypeAllow),
		defineAcl(ResourceTypeTopic, capabilityPrefix, PatternTypePrefix, OperationTypeCreate, PermissionTypeAllow),
		defineAcl(ResourceTypeTopic, capabilityPrefix, PatternTypePrefix, OperationTypeDescribe, PermissionTypeAllow),
		defineAcl(ResourceTypeTopic, capabilityPrefix, PatternTypePrefix, OperationTypeDescribeConfigs, PermissionTypeAllow),

		// for all public topics
		defineAcl(ResourceTypeTopic, publicTopicPrefix, PatternTypePrefix, OperationTypeRead, PermissionTypeAllow),

		// for own public topics
		defineAcl(ResourceTypeTopic, publicCapabilityPrefix, PatternTypePrefix, OperationTypeWrite, PermissionTypeAllow),
		defineAcl(ResourceTypeTopic, publicCapabilityPrefix, PatternTypePrefix, OperationTypeCreate, PermissionTypeAllow),

		// for all connect groups
		defineAcl(ResourceTypeGroup, connectPrefix, PatternTypePrefix, OperationTypeRead, PermissionTypeAllow),
		defineAcl(ResourceTypeGroup, connectPrefix, PatternTypePrefix, OperationTypeWrite, PermissionTypeAllow),
		defineAcl(ResourceTypeGroup, connectPrefix, PatternTypePrefix, OperationTypeCreate, PermissionTypeAllow),

		// for all capability groups
		defineAcl(ResourceTypeGroup, capabilityPrefix, PatternTypePrefix, OperationTypeRead, PermissionTypeAllow),
		defineAcl(ResourceTypeGroup, capabilityPrefix, PatternTypePrefix, OperationTypeWrite, PermissionTypeAllow),
		defineAcl(ResourceTypeGroup, capabilityPrefix, PatternTypePrefix, OperationTypeCreate, PermissionTypeAllow),

		// for cluster
		defineAcl(ResourceTypeCluster, clusterResourceName, PatternTypeLiteral, OperationTypeAlter, PermissionTypeDeny),
		defineAcl(ResourceTypeCluster, clusterResourceName, PatternTypeLiteral, OperationTypeAlterConfigs, PermissionTypeDeny),
		defineAcl(ResourceTypeCluster, clusterResourceName, PatternTypeLiteral, OperationTypeClusterAction, PermissionTypeDeny),
	}
}

func defineAcl(resourceType ResourceType, resourceName string, patternType PatternType, operationType OperationType, permissionType PermissionType) AclDefinition {
	return AclDefinition{
		ResourceType:   resourceType,
		ResourceName:   resourceName,
		PatternType:    patternType,
		OperationType:  operationType,
		PermissionType: permissionType,
	}
}

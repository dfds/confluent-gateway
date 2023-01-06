package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAclDefinition_String(t *testing.T) {
	acl := defineAcl(PatternTypePrefix, "pub.", PatternTypePrefix, OperationTypeRead, PermissionTypeAllow)

	assert.Equal(t, "PREFIXED | pub. | PREFIXED | READ | ALLOW", acl.String())
}

func TestCreateAclDefinitions(t *testing.T) {
	all := CreateAclDefinitions("some-capability")

	result := mapToString(all)

	assert.Equal(t, []string{
		"TOPIC | '*' | PREFIXED | CREATE | DENY",

		"TOPIC | some-capability | PREFIXED | READ | ALLOW",
		"TOPIC | some-capability | PREFIXED | WRITE | ALLOW",
		"TOPIC | some-capability | PREFIXED | CREATE | ALLOW",
		"TOPIC | some-capability | PREFIXED | DESCRIBE | ALLOW",
		"TOPIC | some-capability | PREFIXED | DESCRIBE_CONFIGS | ALLOW",

		"TOPIC | pub. | PREFIXED | READ | ALLOW",

		"TOPIC | pub.some-capability | PREFIXED | WRITE | ALLOW",
		"TOPIC | pub.some-capability | PREFIXED | CREATE | ALLOW",

		"GROUP | connect-some-capability | PREFIXED | READ | ALLOW",
		"GROUP | connect-some-capability | PREFIXED | WRITE | ALLOW",
		"GROUP | connect-some-capability | PREFIXED | CREATE | ALLOW",

		"GROUP | some-capability | PREFIXED | READ | ALLOW",
		"GROUP | some-capability | PREFIXED | WRITE | ALLOW",
		"GROUP | some-capability | PREFIXED | CREATE | ALLOW",

		"CLUSTER | kafka-cluster | LITERAL | ALTER | DENY",
		"CLUSTER | kafka-cluster | LITERAL | ALTER_CONFIGS | DENY",
		"CLUSTER | kafka-cluster | LITERAL | CLUSTER_ACTION | DENY",
	}, result)
}

func mapToString(definitions []AclDefinition) []string {
	var result = make([]string, len(definitions))

	for i, definition := range definitions {
		result[i] = definition.String()
	}
	return result
}

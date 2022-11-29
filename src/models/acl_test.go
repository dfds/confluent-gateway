package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAclDefinition_String(t *testing.T) {
	acl := defineAcl(PatternTypePrefix, "pub.", PatternTypePrefix, OperationTypeRead, PermissionTypeAllow)

	assert.Equal(t, "PREFIX | pub. | PREFIX | READ | ALLOW", acl.String())
}

func TestCreateAclDefinitions(t *testing.T) {
	all := CreateAclDefinitions("some-capability")

	result := mapToString(all)

	assert.Equal(t, []string{
		"TOPIC | '*' | PREFIX | CREATE | DENY",

		"TOPIC | some-capability | PREFIX | READ | ALLOW",
		"TOPIC | some-capability | PREFIX | WRITE | ALLOW",
		"TOPIC | some-capability | PREFIX | CREATE | ALLOW",
		"TOPIC | some-capability | PREFIX | DESCRIBE | ALLOW",
		"TOPIC | some-capability | PREFIX | DESCRIBECONFIGS | ALLOW",

		"TOPIC | pub. | PREFIX | READ | ALLOW",

		"TOPIC | pub.some-capability | PREFIX | WRITE | ALLOW",
		"TOPIC | pub.some-capability | PREFIX | CREATE | ALLOW",

		"GROUP | connect-some-capability | PREFIX | READ | ALLOW",
		"GROUP | connect-some-capability | PREFIX | WRITE | ALLOW",
		"GROUP | connect-some-capability | PREFIX | CREATE | ALLOW",

		"GROUP | some-capability | PREFIX | READ | ALLOW",
		"GROUP | some-capability | PREFIX | WRITE | ALLOW",
		"GROUP | some-capability | PREFIX | CREATE | ALLOW",

		"CLUSTER | kafka-cluster | LITERAL | ALTER | DENY",
		"CLUSTER | kafka-cluster | LITERAL | ALTERCONFIGS | DENY",
		"CLUSTER | kafka-cluster | LITERAL | CLUSTERACTION | DENY",
	}, result)
}

func mapToString(definitions []AclDefinition) []string {
	var result = make([]string, len(definitions))

	for i, definition := range definitions {
		result[i] = definition.String()
	}
	return result
}

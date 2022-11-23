package confluent

type AclEntry struct {
	resourceType   string
	resourceName   string
	patternType    string
	operationType  string
	permissionType string
}

type Acl struct {
	entries []AclEntry
}

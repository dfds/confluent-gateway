package models

type NewTopicHasBeenRequested struct {
	CapabilityRootId string // example => logistics-somecapability-abcd
	ClusterId        string
	TopicName        string // full name => pub.logistics-somecapability-abcd.foo
	Partitions       int
	Retention        int // in ms
}

type CapabilityRootId string
type ServiceAccountId string
type ClusterId string

type Cluster struct {
	ClusterId         ClusterId
	Name              string
	AdminApiEndpoint  string
	AdminApiKey       ApiKey
	BootstrapEndpoint string
}

type ClusterRepository interface {
	Get(id ClusterId) (Cluster, error)
}

type Acl struct {
	Entries []AclEntry
}

type CloudApiAccess struct {
	UserName    string
	Password    string
	ApiEndpoint string
}

type ServiceAccount struct {
	Id               ServiceAccountId
	CapabilityRootId CapabilityRootId
	ClusterAccess    []ClusterAccess
}

type ClusterAccess struct {
	ClusterId        ClusterId
	ServiceAccountId ServiceAccountId
	ApiKey           ApiKey
}

type Topic struct {
	Name       string
	Partitions int
	Retention  int
}

type AclEntry struct {
	ResourceType   string
	ResourceName   string
	PatternType    string
	OperationType  string
	PermissionType string
}

type ApiKey struct {
	UserName string
	Password string
}

type Process struct {
	CapabilityRootId      CapabilityRootId
	ClusterId             ClusterId
	Topic                 Topic
	ServiceAccountId      ServiceAccountId
	Acl                   Acl
	ApiKey                ApiKey
	IsApiKeyStoredInVault bool
}

func (p *Process) ProcessLogic() {
	//1. Ensure capability has cluster access
	//  1.2. Ensure capability has service account
	//	1.3. Ensure service account has all acls
	//	1.4. Ensure service account has api keys
	//	1.5. Ensure api keys are stored in vault
	//2. Ensure topic is created
	//3. Done!
}

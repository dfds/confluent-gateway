package models

// Reference represents a reference to another schema.
type Reference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

// Metadata contains tags, properties, and sensitive fields.
type Metadata struct {
	Tags       map[string][]string `json:"tags"`
	Properties map[string]string   `json:"properties"`
	Sensitive  []string            `json:"sensitive"`
}

// Params represents a map of parameters.
type Params struct {
	Property1 string `json:"property1"`
	Property2 string `json:"property2"`
}

// Rule represents a rule within a rule set.
type Rule struct {
	Name      string   `json:"name"`
	Doc       string   `json:"doc"`
	Kind      string   `json:"kind"`
	Mode      string   `json:"mode"`
	Type      string   `json:"type"`
	Tags      []string `json:"tags"`
	Params    Params   `json:"params"`
	Expr      string   `json:"expr"`
	OnSuccess string   `json:"onSuccess"`
	OnFailure string   `json:"onFailure"`
	Disabled  bool     `json:"disabled"`
}

// RuleSet contains migration and domain rules.
type RuleSet struct {
	MigrationRules []Rule `json:"migrationRules"`
	DomainRules    []Rule `json:"domainRules"`
}

// Schema represents the main schema structure.
type Schema struct {
	Subject    string      `json:"subject"`
	Version    int         `json:"version"`
	ID         int         `json:"id"`
	SchemaType string      `json:"schemaType"`
	References []Reference `json:"references"`
	Schema     string      `json:"schema"`
	Metadata   Metadata    `json:"metadata"`
	RuleSet    RuleSet     `json:"ruleSet"`
}

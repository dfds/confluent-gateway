package models

// ConfluentInternalUser represents a user internally in Confluent which can both represent a normal user or a service account
type ConfluentInternalUser struct {
	Id                 int    `json:"id"`
	Deactivated        bool   `json:"deactivated"`
	ServiceName        string `json:"service_name"`
	ServiceDescription string `json:"service_description"`
	ServiceAccount     bool   `json:"service_account"`
	Internal           bool   `json:"internal"`
	ResourceID         string `json:"resource_id"`
}

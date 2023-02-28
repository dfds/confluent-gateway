package models

type User struct {
	Id                 int    `json:"id"`
	Deactivated        bool   `json:"deactivated"`
	ServiceName        string `json:"service_name"`
	ServiceDescription string `json:"service_description"`
	ServiceAccount     bool   `json:"service_account"`
	Internal           bool   `json:"internal"`
	ResourceID         string `json:"resource_id"`
}

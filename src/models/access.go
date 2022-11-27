package models

type CloudApiAccess struct {
	ApiEndpoint string
	UserName    string
	Password    string
}

type ApiKey struct {
	UserName string
	Password string
}

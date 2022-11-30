package models

type AwsClient interface {
	PutApiKey(apikey ApiKey) error
}

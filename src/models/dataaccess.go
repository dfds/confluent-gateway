package models

import "context"

type DataAccess interface {
	NewSession(context.Context) DataSession
}

type DataSession interface {
	Transaction(func(DataSession) error) error
	ServiceAccounts() ServiceAccountRepository
	Processes() ProcessRepository
}

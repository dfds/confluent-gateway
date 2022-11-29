package models

type DataAccess interface {
	Transaction(f func(DataAccess) error) error
	ServiceAccounts() ServiceAccountRepository
	Processes() ProcessRepository
}

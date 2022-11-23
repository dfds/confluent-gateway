package confluent

type (
	ServiceAccountId interface {
		String() string
	}

	serviceAccountId struct {
		value string
	}
)

func (id *serviceAccountId) String() string {
	return id.value
}

func NewServiceAccountId(value string) (ServiceAccountId, error) {
	return &serviceAccountId{value: value}, nil
}

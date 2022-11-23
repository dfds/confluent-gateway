package confluent

type (
	ClusterId interface {
		String() string
	}
	clusterId struct {
		value string
	}
)

func (id *clusterId) String() string {
	return id.value
}

func NewClusterId(value string) (ClusterId, error) {
	return &clusterId{value: value}, nil
}

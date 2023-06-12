package serviceaccount

type ServiceAccountAccessRequested struct {
	CapabilityId   string `json:"capabilityId"`
	KafkaClusterId string `json:"kafkaClusterId"`
}

func (r *ServiceAccountAccessRequested) GetCapabilityId() string {
	return r.CapabilityId
}

func (r *ServiceAccountAccessRequested) GetClusterId() string {
	return r.KafkaClusterId
}

func (r *ServiceAccountAccessRequested) PartitionKey() string {
	return r.CapabilityId
}

package schema

import (
	"errors"
	"github.com/dfds/confluent-gateway/internal/confluent"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

const someCapabilityId = models.CapabilityId("some-capability-id")
const someClusterId = models.ClusterId("some-cluster-id")
const someMessageContractId = "e72d7a14-b240-4ace-a8e0-27ee0b0ccb25"
const someTopicId = "e72d7a14-b240-4ace-a8e0-27ee0b0ccb25"
const someTopicName = "some-topic-name"

var serviceError = errors.New("service error")
var confluentError = confluent.NewClientError("http://confluent", 422, "bad message")

func Test_createProcessState(t *testing.T) {
	tests := []struct {
		name            string
		mock            *mock
		input           ProcessInput
		topic           models.Topic
		wantIsCompleted bool
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			mock: &mock{},
			input: ProcessInput{
				MessageContractId: someMessageContractId,
				TopicId:           someTopicId,
				MessageType:       "message-type",
				Schema:            "{}",
				Description:       "description",
			},
			topic:           models.Topic{ClusterId: someClusterId, Name: someTopicName},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already exists",
			mock: &mock{ReturnProcessState: &models.SchemaProcess{
				ClusterId:         someClusterId,
				MessageContractId: someMessageContractId,
				TopicId:           someTopicId,
				MessageType:       "message-type",
				Subject:           someTopicName + "-message-type",
				Schema:            "{}",
				Description:       "description",
			}},
			input:           ProcessInput{MessageContractId: someMessageContractId},
			topic:           models.Topic{},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already finished",
			mock: &mock{ReturnProcessState: &models.SchemaProcess{
				ClusterId:         someClusterId,
				MessageContractId: someMessageContractId,
				TopicId:           someTopicId,
				MessageType:       "message-type",
				Subject:           someTopicName + "-message-type",
				Schema:            "{}",
				Description:       "description",
			}},
			input:           ProcessInput{MessageContractId: someMessageContractId},
			topic:           models.Topic{},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getOrCreateProcessState(tt.mock, tt.input, &tt.topic)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, someMessageContractId, got.MessageContractId)
			assert.Equal(t, someTopicId, got.TopicId)
			assert.Equal(t, "message-type", got.MessageType)
			assert.Equal(t, "{}", got.Schema)
			assert.Equal(t, "description", got.Description)
			assert.Equal(t, someTopicName+"-message-type", got.Subject)
			assert.Equal(t, someClusterId, got.ClusterId)
			assert.Equal(t, tt.wantIsCompleted, got.IsCompleted())
		})
	}
}

type mock struct {
	ReturnProcessState   *models.SchemaProcess
	ReturnServiceAccount *models.ServiceAccount
}

func (m *mock) GetSchemaProcessState(messageContractId string) (*models.SchemaProcess, error) {
	return m.ReturnProcessState, nil
}

func (m *mock) SaveSchemaProcessState(state *models.SchemaProcess) error {
	return nil
}

func Test_ensureSchemaIsRegistered(t *testing.T) {
	tests := []struct {
		name               string
		context            *mocks.StepContextMock
		wantErr            assert.ErrorAssertionFunc
		marked             bool
		okEventRaised      bool
		failureEventRaised bool
	}{
		{
			name:               "ok",
			context:            &mocks.StepContextMock{},
			wantErr:            assert.NoError,
			marked:             true,
			okEventRaised:      true,
			failureEventRaised: false,
		},
		{
			name:               "schema registration error",
			context:            &mocks.StepContextMock{OnRegisterSchemaError: serviceError},
			wantErr:            assert.Error,
			marked:             false,
			okEventRaised:      false,
			failureEventRaised: false,
		},
		{
			name:               "confluent client schema registration error",
			context:            &mocks.StepContextMock{OnRegisterSchemaError: confluentError},
			wantErr:            assert.NoError,
			marked:             true,
			okEventRaised:      false,
			failureEventRaised: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureSchemaIsRegisteredStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkAsCompletedWasCalled)
			assert.Equal(t, tt.okEventRaised, tt.context.SchemaRegisteredEventWasRaised)
			assert.Equal(t, tt.failureEventRaised, tt.context.SchemaRegistrationFailedWasRaised)
		})
	}
}

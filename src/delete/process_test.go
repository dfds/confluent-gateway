package create

import (
	"errors"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/dfds/confluent-gateway/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const someCapabilityRootId = models.CapabilityRootId("some-capability-root-id")
const someClusterId = models.ClusterId("some-cluster-id")
const someTopicName = "some-topic-name"

var serviceError = errors.New("service error")

func Test_createProcessState(t *testing.T) {
	tests := []struct {
		name            string
		mock            *mock
		input           ProcessInput
		wantIsCompleted bool
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			mock: &mock{},
			input: ProcessInput{
				CapabilityRootId: someCapabilityRootId,
				ClusterId:        someClusterId,
				TopicName:        someTopicName,
			},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already exists",
			mock: &mock{ReturnProcessState: &models.DeleteProcess{
				CapabilityRootId: someCapabilityRootId,
				ClusterId:        someClusterId,
				TopicName:        someTopicName,
				CompletedAt:      nil,
			}},
			input: ProcessInput{
				CapabilityRootId: someCapabilityRootId,
				ClusterId:        someClusterId,
				TopicName:        someTopicName,
			},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already finished",
			mock: &mock{ReturnProcessState: &models.DeleteProcess{
				CapabilityRootId: someCapabilityRootId,
				ClusterId:        someClusterId,
				TopicName:        someTopicName,
				CompletedAt:      &time.Time{},
			}},
			input: ProcessInput{
				CapabilityRootId: someCapabilityRootId,
				ClusterId:        someClusterId,
				TopicName:        someTopicName,
			},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getOrCreateProcessState(tt.mock, tt.input)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, someCapabilityRootId, got.CapabilityRootId)
			assert.Equal(t, someClusterId, got.ClusterId)
			assert.Equal(t, someTopicName, got.TopicName)
			assert.Equal(t, tt.wantIsCompleted, got.IsCompleted())
		})
	}
}

type mock struct {
	ReturnProcessState   *models.DeleteProcess
	ReturnServiceAccount *models.ServiceAccount
}

func (m *mock) GetDeleteProcessState(models.CapabilityRootId, models.ClusterId, string) (*models.DeleteProcess, error) {
	return m.ReturnProcessState, nil
}

func (m *mock) SaveDeleteProcessState(*models.DeleteProcess) error {
	return nil
}

func Test_ensureTopicIsDeleted(t *testing.T) {
	tests := []struct {
		name        string
		context     *mocks.StepContextMock
		wantErr     assert.ErrorAssertionFunc
		marked      bool
		eventRaised bool
	}{
		{
			name:        "ok",
			context:     &mocks.StepContextMock{},
			wantErr:     assert.NoError,
			marked:      true,
			eventRaised: true,
		},
		{
			name:        "create topic error",
			context:     &mocks.StepContextMock{OnDeleteTopicError: serviceError},
			wantErr:     assert.Error,
			marked:      false,
			eventRaised: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, ensureTopicIsDeletedStep(tt.context))

			assert.Equal(t, tt.marked, tt.context.MarkAsCompletedWasCalled)
			assert.Equal(t, tt.eventRaised, tt.context.TopicDeletedEventWasRaised)
		})
	}
}

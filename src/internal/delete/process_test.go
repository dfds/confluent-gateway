package delete

import (
	"errors"
	"github.com/dfds/confluent-gateway/internal/models"
	"github.com/dfds/confluent-gateway/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const someCapabilityId = models.CapabilityId("some-capability-id")
const someClusterId = models.ClusterId("some-cluster-id")
const someTopicId = "e72d7a14-b240-4ace-a8e0-27ee0b0ccb25"
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
				TopicId: someTopicId,
			},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already exists",
			mock: &mock{ReturnProcessState: &models.DeleteProcess{
				TopicId:     someTopicId,
				CompletedAt: nil,
			}},
			input: ProcessInput{
				TopicId: someTopicId,
			},
			wantIsCompleted: false,
			wantErr:         assert.NoError,
		},
		{
			name: "process already finished",
			mock: &mock{ReturnProcessState: &models.DeleteProcess{
				TopicId:     someTopicId,
				CompletedAt: &time.Time{},
			}},
			input: ProcessInput{
				TopicId: someTopicId,
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
			assert.Equal(t, someTopicId, got.TopicId)
			assert.Equal(t, tt.wantIsCompleted, got.IsCompleted())
		})
	}
}

type mock struct {
	ReturnProcessState   *models.DeleteProcess
	ReturnServiceAccount *models.ServiceAccount
}

func (m *mock) GetDeleteProcessState(string) (*models.DeleteProcess, error) {
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

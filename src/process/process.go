package process

import (
	"context"
	"github.com/dfds/confluent-gateway/models"
)

type Process struct {
	Context   context.Context
	Session   DataSession
	Confluent Confluent
	Vault     Vault
	State     *models.ProcessState
}

func (p *Process) CapabilityRootId() models.CapabilityRootId {
	return p.State.CapabilityRootId
}

func (p *Process) ClusterId() models.ClusterId {
	return p.State.ClusterId
}

func (p *Process) createTopic() error {
	topic := p.State.Topic()
	return p.Confluent.CreateTopic(p.Context, p.State.ClusterId, topic.Name, topic.Partitions, topic.Retention.Milliseconds())
}

func NewProcess(ctx context.Context, session DataSession, confluent Confluent, vault Vault, state *models.ProcessState) *Process {
	return &Process{
		Context:   ctx,
		Session:   session,
		Confluent: confluent,
		Vault:     vault,
		State:     state,
	}
}

func (p *Process) NewSession(session DataSession) *Process {
	return &Process{
		Session:   session,
		Context:   p.Context,
		State:     p.State,
		Confluent: p.Confluent,
		Vault:     p.Vault,
	}
}

func (p *Process) Execute(step Step) error {
	return p.Session.Transaction(func(session DataSession) error {
		np := p.NewSession(session)

		err := step(np)
		if err != nil {
			return err
		}

		return session.UpdateProcessState(p.State)
	})
}

func (p *Process) service() *AccountService {
	return NewAccountService(p.Context, p.Confluent, p.Session)
}

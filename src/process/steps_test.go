package process

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSteps_Run(t *testing.T) {
	c := NewCollector()

	err := PrepareSteps().
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Run(c.Execute)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, c.steps)
}

func TestSteps_Until(t *testing.T) {
	c := NewCollector()

	err := PrepareSteps().
		Step(c.DummyStep()).
		Step(c.DummyStep()).Until(c.Count(1)).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Run(c.Execute)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 2, 3, 4, 5}, c.steps)
}

type Collector struct {
	steps   []int
	cnt     int
	process *Process
}

func NewCollector() *Collector {
	return &Collector{steps: []int{}, process: &Process{}}
}

func (c *Collector) DummyStep() Step {
	c.cnt++
	cnt := c.cnt

	return func(p *Process) error {
		c.steps = append(c.steps, cnt)

		return nil
	}
}

func (c *Collector) Count(maxCnt int) Predicate {
	cnt := 0
	return func(*Process) bool {
		if cnt <= maxCnt {
			cnt++
			return false
		}
		return true
	}
}

func (c *Collector) Execute(step Step) error {
	return step(c.process)
}

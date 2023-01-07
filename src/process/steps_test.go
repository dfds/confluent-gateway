package process

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Context struct {
	calls int
}

func TestSteps_Run(t *testing.T) {
	c := NewCollector()

	err := PrepareSteps[*Context]().
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Run(c.Execute)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, c.steps)
	assert.Equal(t, 5, c.process.calls)
}

func TestSteps_Until(t *testing.T) {
	c := NewCollector()

	err := PrepareSteps[*Context]().
		Step(c.DummyStep()).
		Step(c.DummyStep()).Until(c.Count(2)).
		Step(c.DummyStep()).Until(c.Count(2)).
		Step(c.DummyStep()).
		Step(c.DummyStep()).
		Run(c.Execute)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 2, 3, 3, 4, 5}, c.steps)
	assert.Equal(t, 7, c.process.calls)
}

type Collector struct {
	steps   []int
	cnt     int
	process *Context
}

func NewCollector() *Collector {
	return &Collector{steps: []int{}, process: &Context{}}
}

func (c *Collector) DummyStep() func(*Context) error {
	c.cnt++
	cnt := c.cnt

	return func(p *Context) error {
		p.calls++
		c.steps = append(c.steps, cnt)

		return nil
	}
}

func (c *Collector) Count(maxCnt int) func(*Context) bool {
	cnt := 0
	return func(*Context) bool {
		if cnt < maxCnt {
			cnt++
			return false
		}
		return true
	}
}

func (c *Collector) Execute(step func(*Context) error) error {
	return step(c.process)
}

package process

type steps struct {
	steps []stepWrapper
}

type stepWrapper func(*Process) (bool, error)

type Step func(*Process) error
type Predicate func() bool

type NextStepBuilder interface {
	Step(Step) StepBuilder
	Run(Transaction) error
}

type StepBuilder interface {
	NextStepBuilder
	Until(Predicate) NextStepBuilder
}

type Transaction interface {
	Execute(Step) error
}

func PrepareSteps() NextStepBuilder {
	return &steps{steps: []stepWrapper{}}
}

func (s *steps) Step(step Step) StepBuilder {
	s.steps = append(s.steps, func(p *Process) (bool, error) {
		err := step(p)
		return true, err
	})
	return s
}

func (s *steps) Until(isDone Predicate) NextStepBuilder {
	lastStep := s.steps[len(s.steps)-1]

	s.steps[len(s.steps)-1] = func(p *Process) (bool, error) {
		if isDone() {
			return true, nil
		}

		_, err := lastStep(p)
		return false, err
	}

	return s
}

func (s *steps) Run(transaction Transaction) error {
	for _, step := range s.steps {
		for {
			done := true

			err := transaction.Execute(func(p *Process) error {
				var err error
				done, err = step(p)
				return err
			})

			if err != nil {
				return err
			}
			if done {
				break
			}
		}
	}

	return nil
}

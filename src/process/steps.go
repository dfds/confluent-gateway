package process

type stepWrapper func(*Process) (bool, error)
type Step func(*Process) error
type Predicate func(*Process) bool
type PerformStep func(Step) error

type StepBuilder interface {
	Step(Step) NextStepBuilder
}

type NextStepBuilder interface {
	StepBuilder
	Until(Predicate) StepBuilder
	Run(PerformStep) error
}

type steps struct {
	steps []stepWrapper
}

func PrepareSteps() StepBuilder {
	return &steps{steps: []stepWrapper{}}
}

func (s *steps) Step(step Step) NextStepBuilder {
	s.steps = append(s.steps, func(p *Process) (bool, error) {
		err := step(p)
		return true, err
	})
	return s
}

func (s *steps) Until(isDone Predicate) StepBuilder {
	lastStep := s.steps[len(s.steps)-1]

	s.steps[len(s.steps)-1] = func(p *Process) (bool, error) {
		if isDone(p) {
			return true, nil
		}

		_, err := lastStep(p)
		return false, err
	}

	return s
}

func (s *steps) Run(perform PerformStep) error {
	for _, step := range s.steps {
		for {
			done := true

			err := perform(func(p *Process) error {
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

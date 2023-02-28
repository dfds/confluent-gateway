package process

type StepBuilder[Context any] interface {
	Step(func(Context) error) NextStepBuilder[Context]
}

type NextStepBuilder[Context any] interface {
	StepBuilder[Context]
	Until(func(Context) bool) StepBuilder[Context]
	Run(func(func(Context) error) error) error
}

type Steps[Context any] struct {
	steps []func(Context) (bool, error)
}

func PrepareSteps[Context any]() StepBuilder[Context] {
	return &Steps[Context]{steps: []func(Context) (bool, error){}}
}

func (s *Steps[Context]) Step(step func(Context) error) NextStepBuilder[Context] {
	s.steps = append(s.steps, func(context Context) (bool, error) {
		err := step(context)
		return true, err
	})
	return s
}

func (s *Steps[Context]) Until(isDone func(Context) bool) StepBuilder[Context] {
	lastStep := s.steps[len(s.steps)-1]

	s.steps[len(s.steps)-1] = func(context Context) (bool, error) {
		if isDone(context) {
			return true, nil
		}

		_, err := lastStep(context)
		return false, err
	}

	return s
}

func (s *Steps[Context]) Run(perform func(func(Context) error) error) error {
	for _, step := range s.steps {
		for {
			done := true

			err := perform(func(context Context) error {
				var err error
				done, err = step(context)
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

package helpers

import "fmt"

type ErrorList struct {
	errors []error
}

func (e *ErrorList) AppendIfErr(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}
func (e *ErrorList) AppendErrors(errors ErrorList) {

	e.errors = append(e.errors, errors.errors...)
}
func (e *ErrorList) HasErrors() bool {
	return len(e.errors) > 0
}
func (e *ErrorList) String() string {

	output := ""
	for _, err := range e.errors {
		output += fmt.Sprintf("%s\n", err)
	}
	return output
}

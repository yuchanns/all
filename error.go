package all

import "strings"

type errors struct {
	errs []error
}

// Implement the error interface for the errors type.
func (e *errors) Error() string {
	var desc []string
	for _, err := range e.errs {
		desc = append(desc, err.Error())
	}
	return strings.Join(desc, ",")
}

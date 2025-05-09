package dbinterface

import "fmt"

type errNotFound struct {
	term string
}

func (e *errNotFound) Error() string {
	return fmt.Sprintf("Unable to find %q", e.term)
}

type errUnexpectedLanguage struct {
	expectedLanguage string
	term             string
}

func (e *errUnexpectedLanguage) Error() string {
	return fmt.Sprintf("%q is not in expected language of %s", e.term, e.expectedLanguage)
}

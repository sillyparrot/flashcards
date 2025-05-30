package dbinterface

import "fmt"

type ErrNotFound struct {
	term string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("Unable to find %q", e.term)
}

type ErrUnexpectedLanguage struct {
	expectedLanguage string
	term             string
}

func (e *ErrUnexpectedLanguage) Error() string {
	return fmt.Sprintf("%q is not in expected language of %s", e.term, e.expectedLanguage)
}

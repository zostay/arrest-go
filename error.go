package arrest

import (
	"errors"

	"github.com/zostay/go-std/set"
)

type ErrHandler interface {
	Err() error
	Errs() []error
	AddError(...error)
	AddHandler(...ErrHandler)
}

type ErrHelper struct {
	err       []error
	nestedErr set.Set[ErrHandler]
}

func (e *ErrHelper) Err() error {
	return errors.Join(e.Errs()...)
}

func (e *ErrHelper) Errs() []error {
	var errs []error
	if len(e.err) > 0 {
		errs = append(errs, e.err...)
	}

	for ne := range e.nestedErr {
		errs = append(errs, ne.Errs()...)
	}

	return errs
}

func (e *ErrHelper) AddHandler(handlers ...ErrHandler) {
	if e.nestedErr == nil {
		e.nestedErr = set.New[ErrHandler]()
	}

	for _, handler := range handlers {
		e.nestedErr.Insert(handler)
	}
}

func (e *ErrHelper) AddError(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		e.err = append(e.err, err)
	}
}

func withErr[T ErrHandler](e T, errs ...error) T {
	e.AddError(errs...)
	return e
}

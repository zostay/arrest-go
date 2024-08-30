package arrest

import (
	"errors"

	"github.com/zostay/go-std/set"
)

// ErrHandler is the interface that all DSL object implement to allow errors to
// flow upward to parent components.
type ErrHandler interface {
	Err() error
	Errs() []error
	AddError(...error)
	AddHandler(...ErrHandler)
}

// ErrHelper implements the ErrHandler to make implementation of the interface
// as easy as embedding ErrHelper into the struct.
type ErrHelper struct {
	err       []error
	nestedErr set.Set[ErrHandler]
}

// Err returns an error if there are any. This uses errors.JOin to join together
// all of the errors in this component as well as any in any child component.
func (e *ErrHelper) Err() error {
	return errors.Join(e.Errs()...)
}

// Errs returns all of the errors in this component as well as any in any child
// component. You probably went Err() instead.
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

// AddHandler is used internally to add child component errors to the parent.
func (e *ErrHelper) AddHandler(handlers ...ErrHandler) {
	if e.nestedErr == nil {
		e.nestedErr = set.New[ErrHandler]()
	}

	for _, handler := range handlers {
		e.nestedErr.Insert(handler)
	}
}

// AddError adds the given errors to the current component.
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

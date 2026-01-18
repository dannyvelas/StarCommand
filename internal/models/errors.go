package models

import "fmt"

type ErrAlreadyExists struct {
	Name     string
	Resource string
}

func NewErrAlreadyExists(name, resource string) *ErrAlreadyExists {
	return &ErrAlreadyExists{
		Name:     name,
		Resource: resource,
	}
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists", e.Resource)
}

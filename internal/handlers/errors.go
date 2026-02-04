package handlers

import "errors"

var (
	errConnectingSSH = errors.New("error connecting via ssh")
	errAlreadyExists = errors.New("resource already exists")
	errNotFound      = errors.New("not found")
)

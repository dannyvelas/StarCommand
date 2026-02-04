package handlers

import "context"

type Handler interface {
	GetConfig(hostAlias string) any
	Execute(ctx context.Context, config any, hostAlias string) (map[string]string, error)
}

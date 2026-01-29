package handlers

type Handler interface {
	GetConfig(hostAlias string) any
	Execute(config any, hostAlias string) (map[string]string, error)
}

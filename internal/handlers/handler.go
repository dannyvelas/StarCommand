package handlers

type Handler interface {
	GetConfig(hostAlias string) any
	Execute(config map[string]string, hostAlias string) (map[string]string, error)
}

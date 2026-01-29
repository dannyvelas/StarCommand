package app

type handler interface {
	getConfig(hostAlias string) any
	execute(config map[string]string, hostAlias string) (map[string]string, error)
}

package config

type unvalidatedReader interface {
	ReadUnvalidated() (unvalidatedResult, error)
}

type unvalidatedResult interface {
	getConfigMap() map[string]string
}

type simpleUnvalidatedResult struct {
	configMap map[string]string
}

func (r simpleUnvalidatedResult) getConfigMap() map[string]string {
	return r.configMap
}

type diagnosticUnvalidatedResult struct {
	configMap     map[string]string
	diagnosticMap map[string]string
}

func (r diagnosticUnvalidatedResult) getConfigMap() map[string]string {
	return r.configMap
}

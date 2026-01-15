package config

type Reader interface {
	read() (readResult, error)
}

type readResult interface {
	getConfigMap() map[string]string
}

type simpleReadResult struct {
	configMap map[string]string
}

func (r simpleReadResult) getConfigMap() map[string]string {
	return r.configMap
}

type diagnosticReadResult struct {
	configMap     map[string]string
	diagnostics map[string]string
}

func (r diagnosticReadResult) getConfigMap() map[string]string {
	return r.configMap
}

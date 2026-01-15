package config

type Reader interface {
	Read() (ReadResult, error)
}

type ReadResult interface {
	readResult()
	GetConfigMap() map[string]string
}

type SimpleReadResult struct {
	configMap map[string]string
}

func NewSimpleReadResult(configMap map[string]string) SimpleReadResult {
	return SimpleReadResult{configMap: configMap}
}

func (r SimpleReadResult) readResult() {}

func (r SimpleReadResult) GetConfigMap() map[string]string {
	return r.configMap
}

type DiagnosticReadResult struct {
	configMap   map[string]string
	diagnostics map[string]string
}

func NewDiagnosticReadResult(configMap, diagnostics map[string]string) DiagnosticReadResult {
	return DiagnosticReadResult{
		configMap:   configMap,
		diagnostics: diagnostics,
	}
}

func (r DiagnosticReadResult) readResult() {}

func (r DiagnosticReadResult) GetConfigMap() map[string]string {
	return r.configMap
}

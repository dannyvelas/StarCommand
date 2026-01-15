package config

// Reader is the interface that must be implemented
// if you want to define your own source of reading
// configuration data
type Reader interface {
	Read() (ReadResult, error)
}

// ReadResult is the expected return value from the
// Read() function of a Reader
// You cannot define a struct outside of this package
// that implements this interface. You must use either
// NewSimpleReadResult or NewDiagnosticReadResult to
// initialize a value that implements ReadResult
type ReadResult interface {
	readResult()
	GetConfigMap() map[string]string
}

// SimpleReadResult is what you should return from a
// Read() function if your configuration source doesn't
// have any diagnostics to report
type SimpleReadResult struct {
	configMap map[string]string
}

// NewSimpleReadResult creates a new SimpleReadResult
func NewSimpleReadResult(configMap map[string]string) SimpleReadResult {
	return SimpleReadResult{configMap: configMap}
}

func (r SimpleReadResult) readResult() {}

func (r SimpleReadResult) GetConfigMap() map[string]string {
	return r.configMap
}

// DiagnosticReadResult is what you should return from a
// Read() function if your configuration source has
// diagnostics to report.
// For example, the BitwardenSecretReader uses this as a
// return value. If the BitwardenSecretReader realizes
// that it doesn't have enough configuration values to
// authenticate to Bitwarden, it fills the "diagnostics" map
// with information about what fields were missing.
type DiagnosticReadResult struct {
	configMap   map[string]string
	diagnostics map[string]string
}

// NewDiagnosticReadResult creates a new DiagnosticReadResult
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

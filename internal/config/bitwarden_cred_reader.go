package config

var _ unvalidatedReader = bitwardenCredReader{}

type bitwardenCredReader struct {
	configMap map[string]string
}

func newBitwardenCredReader(configMap map[string]string) bitwardenCredReader {
	return bitwardenCredReader{
		configMap: configMap,
	}
}

func (r bitwardenCredReader) ReadUnvalidated() (unvalidatedResult, error) {
	return simpleUnvalidatedResult(r), nil
}

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

func (p bitwardenCredReader) ReadUnvalidated() (unvalidatedResult, error) {
	return simpleUnvalidatedResult{configMap: p.configMap}, nil
}

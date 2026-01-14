package config

var _ Reader = bitwardenCredReader{}

type bitwardenCredReader struct {
	configMap map[string]string
}

func newBitwardenCredReader(configMap map[string]string) bitwardenCredReader {
	return bitwardenCredReader{
		configMap: configMap,
	}
}

func (r bitwardenCredReader) read() (readResult, error) {
	return simpleReadResult(r), nil
}

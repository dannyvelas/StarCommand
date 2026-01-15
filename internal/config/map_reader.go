package config

var _ Reader = mapReader{}

type mapReader struct {
	configMap map[string]string
}

func newMapReader(configMap map[string]string) mapReader {
	return mapReader{
		configMap: configMap,
	}
}

func (r mapReader) read() (readResult, error) {
	return simpleReadResult(r), nil
}

package config

import (
	"fmt"
)

var _ provider = bitwardenCredProvider{}

type bitwardenCredProvider struct {
	configMap map[string]string
}

func newBitwardenCredProvider(configMap map[string]string) bitwardenCredProvider {
	return bitwardenCredProvider{
		configMap: configMap,
	}
}

func (p bitwardenCredProvider) UnmarshalInto(target any) error {
	if err := decode(p.configMap, target); err != nil {
		return fmt.Errorf("error reading bitwarden config into a map: %v", err)
	}

	return nil
}

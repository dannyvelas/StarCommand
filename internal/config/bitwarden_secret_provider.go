package config

import (
	"fmt"

	"github.com/dannyvelas/homelab/internal/client"
)

var _ unvalidatedReader = bitwardenSecretReader{}

type bitwardenSecretReader struct {
	bitwardenCredReader bitwardenCredReader
}

func newBitwardenSecretReader(configMap map[string]string) bitwardenSecretReader {
	return bitwardenSecretReader{
		bitwardenCredReader: newBitwardenCredReader(configMap),
	}
}

func (p bitwardenSecretReader) ReadUnvalidated() (map[string]string, error) {
	config := newBitwardenConfig()
	if err := UnmarshalInto(p.bitwardenCredReader, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling bitwarden creds: %v", err)
	}

	results, ok, err := validateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error validating bitwarden config: %v", err)
	} else if !ok {
		return nil, fmt.Errorf("error: invalid bitwarden configs: %s", fmtTable(results))
	}

	bitwardenClient, err := client.NewBitwardenClient(
		config.APIURL,
		config.IdentityURL,
		config.AccessToken,
		config.OrganizationID,
		config.ProjectID,
		config.StateFilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	bitwardenSecrets, err := bitwardenClient.ReadSecrets()
	if err != nil {
		return nil, fmt.Errorf("error reading bitwarden secrets: %v", err)
	}

	return bitwardenSecrets, nil
}

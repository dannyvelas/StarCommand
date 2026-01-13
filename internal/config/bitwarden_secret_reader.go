package config

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/client"
)

var _ reader = (*bitwardenSecretReader)(nil)

type bitwardenSecretReader struct {
	bitwardenCredReader bitwardenCredReader
}

func newBitwardenSecretReader(configMap map[string]string) *bitwardenSecretReader {
	return &bitwardenSecretReader{
		bitwardenCredReader: newBitwardenCredReader(configMap),
	}
}

func (r *bitwardenSecretReader) read() (readResult, error) {
	config := newBitwardenConfig()

	if _, err := UnmarshalInto(r.bitwardenCredReader, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling bitwarden creds: %v", err)
	}

	results, err := validateConfig(config)
	if errors.Is(err, ErrInvalidFields) {
		return diagnosticReadResult{configMap: nil, diagnosticMap: results}, ErrInvalidFields
	} else if err != nil {
		return nil, fmt.Errorf("error validating bitwarden config: %v", err)
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

	return diagnosticReadResult{
		configMap:     bitwardenSecrets,
		diagnosticMap: results,
	}, nil
}

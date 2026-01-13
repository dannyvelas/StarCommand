package config

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/client"
)

var (
	_ unvalidatedReader = (*bitwardenSecretReader)(nil)
	_ diagnosticReader  = (*bitwardenSecretReader)(nil)
)

type bitwardenSecretReader struct {
	bitwardenCredReader bitwardenCredReader
	diagnosticMap       map[string]string
}

func newBitwardenSecretReader(configMap map[string]string) *bitwardenSecretReader {
	return &bitwardenSecretReader{
		bitwardenCredReader: newBitwardenCredReader(configMap),
	}
}

func (p *bitwardenSecretReader) ReadUnvalidated() (map[string]string, error) {
	config := newBitwardenConfig()

	if err := UnmarshalInto(p.bitwardenCredReader, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling bitwarden creds: %v", err)
	}

	results, err := validateConfig(config)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error validating bitwarden config: %v", err)
	}

	p.diagnosticMap = results
	if errors.Is(err, ErrInvalidFields) {
		return nil, ErrInvalidFields
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

func (p *bitwardenSecretReader) GetDiagnosticMap() map[string]string {
	return p.diagnosticMap
}

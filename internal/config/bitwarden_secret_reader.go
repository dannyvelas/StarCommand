package config

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/client"
)

var _ Reader = (*bitwardenSecretReader)(nil)

type bitwardenSecretReader struct {
	mapReader mapReader
}

func NewBitwardenSecretReader(configMap map[string]string) *bitwardenSecretReader {
	return &bitwardenSecretReader{
		mapReader: newMapReader(configMap),
	}
}

func (r *bitwardenSecretReader) Read() (ReadResult, error) {
	config := newBitwardenConfig()

	diagnostics, err := Unmarshal(r.mapReader, &config)
	if errors.Is(err, ErrInvalidFields) {
		return NewDiagnosticReadResult(nil, diagnostics), ErrInvalidFields
	} else if err != nil {
		return nil, fmt.Errorf("error unmarshalling bitwarden creds: %v", err)
	}

	bitwardenClient, err := client.NewBitwardenClient(
		config.APIURL,
		config.IdentityURL,
		config.AccessToken,
		config.OrganizationID,
		config.StateFilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	bitwardenSecrets, err := bitwardenClient.ReadSecrets()
	if err != nil {
		return nil, fmt.Errorf("error reading bitwarden secrets: %v", err)
	}

	return NewDiagnosticReadResult(bitwardenSecrets, diagnostics), nil
}

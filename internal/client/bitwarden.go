package client

import (
	"fmt"

	"github.com/bitwarden/sdk-go"
)

type BitwardenClient struct {
	organizationID string
	client         sdk.BitwardenClientInterface
}

func NewBitwardenClient(apiURL, identityURL, accessToken, organizationID, stateFile string) (BitwardenClient, error) {
	bitwardenClient, err := sdk.NewBitwardenClient(&apiURL, &identityURL)
	if err != nil {
		return BitwardenClient{}, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	if err := bitwardenClient.AccessTokenLogin(accessToken, &stateFile); err != nil {
		return BitwardenClient{}, fmt.Errorf("error logging in to bitwarden client: %v", err)
	}

	return BitwardenClient{
		organizationID: organizationID,
		client:         bitwardenClient,
	}, nil
}

func (c BitwardenClient) ReadSecrets() (map[string]string, error) {
	m := make(map[string]string)
	listResponse, err := c.client.Secrets().List(c.organizationID)
	if err != nil {
		return nil, fmt.Errorf("error listing secrets: %v", err)
	}

	for _, secret := range listResponse.Data {
		secretData, err := c.client.Secrets().Get(secret.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting secret: %v", err)
		}

		m[secret.Key] = secretData.Value
	}

	return m, nil
}

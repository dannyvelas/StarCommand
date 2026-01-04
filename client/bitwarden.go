package client

import (
	"fmt"

	"github.com/bitwarden/sdk-go"
)

type BitwardenClient struct {
	projectID string
	client    sdk.BitwardenClientInterface
}

func NewBitwardenClient(apiURL, identityURL, projectID string) (BitwardenClient, error) {
	bitwardenClient, err := sdk.NewBitwardenClient(&apiURL, &identityURL)
	if err != nil {
		return BitwardenClient{}, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	return BitwardenClient{
		projectID: projectID,
		client:    bitwardenClient,
	}, nil
}

func (c BitwardenClient) GetSecretByName(secretName string) (string, error) {
	listResponse, err := c.client.Secrets().List(c.projectID)
	if err != nil {
		return "", err
	}

	for _, secret := range listResponse.Data {
		if secret.Key == secretName {
			secretData, err := c.client.Secrets().Get(secret.ID)
			if err != nil {
				return "", err
			}
			return secretData.Value, nil
		}
	}

	return "", fmt.Errorf("secret with name %s not found", secretName)
}

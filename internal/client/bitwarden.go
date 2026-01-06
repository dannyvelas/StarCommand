package client

import (
	"fmt"
	"reflect"

	"github.com/bitwarden/sdk-go"
	"github.com/dannyvelas/homelab/internal/helpers"
)

type BitwardenClient struct {
	organizationID string
	projectID      string
	client         sdk.BitwardenClientInterface
}

func NewBitwardenClient(apiURL, identityURL, accessToken, organizationID, projectID, stateFile string) (BitwardenClient, error) {
	bitwardenClient, err := sdk.NewBitwardenClient(&apiURL, &identityURL)
	if err != nil {
		return BitwardenClient{}, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	if err := bitwardenClient.AccessTokenLogin(accessToken, &stateFile); err != nil {
		return BitwardenClient{}, fmt.Errorf("error logging in to bitwarden client: %v", err)
	}

	return BitwardenClient{
		organizationID: organizationID,
		projectID:      projectID,
		client:         bitwardenClient,
	}, nil
}

// FillStruct takes a struct as an argument and fills its fields
// with values found in c.organizationID
func (c BitwardenClient) FillStruct(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("error: expected pointer argument to FillStruct")
	}

	listResponse, err := c.client.Secrets().List(c.organizationID)
	if err != nil {
		return fmt.Errorf("error listing secrets: %v", err)
	}

	tagToFieldMap, err := helpers.GetTagToFieldMap(v, "bw", "json")
	if err != nil {
		return fmt.Errorf("error getting tag to field map: %v", err)
	}

	for _, secret := range listResponse.Data {
		field, ok := tagToFieldMap[secret.Key]
		if !ok {
			continue
		}

		secretData, err := c.client.Secrets().Get(secret.ID)
		if err != nil {
			return fmt.Errorf("error getting secret: %v", err)
		}

		field.Value.SetString(secretData.Value)
	}

	return nil
}

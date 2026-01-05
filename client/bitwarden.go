package client

import (
	"fmt"
	"reflect"

	"github.com/bitwarden/sdk-go"
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

	for _, secret := range listResponse.Data {
		field, ok := getFieldByTag(v, secret.Key)
		if !ok {
			continue
		}

		secretData, err := c.client.Secrets().Get(secret.ID)
		if err != nil {
			return fmt.Errorf("error getting secret: %v", err)
		}

		field.SetString(secretData.Value)
	}

	return nil
}

// getFieldByTag takes a struct and returns the value of a field with a yaml tag that equals `tag`.
// If no field was found, nil and false are returned.
func getFieldByTag(v any, tag string) (reflect.Value, bool) {
	rv := reflect.ValueOf(v)

	// If a pointer is passed, get the underlying element (the actual struct)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	// If it's not a struct, we can't look up tags
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		foundTag := field.Tag.Get("bw")

		// if "bw" tag is not present, fall back to json
		if foundTag == "" {
			foundTag = field.Tag.Get("json")
		}

		if foundTag == tag {
			return rv.Field(i), true
		}
	}

	return reflect.Value{}, false
}

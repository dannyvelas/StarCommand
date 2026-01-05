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
		field, ok := getFieldByYAMLTag(v, secret.Key)
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

// getFieldByYAMLTag takes a struct and returns the value of a field with a yaml tag that equals `tag`.
// If no field was found, nil and false are returned.
func getFieldByYAMLTag(v any, tag string) (reflect.Value, bool) {
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
		// Get the content of the "yaml" tag
		yamlTag := field.Tag.Get("yaml")

		// Tags can look like `yaml:"my_field,omitempty"`.
		// We only want the name part before the comma.
		name := yamlTag
		if commaIdx := findComma(yamlTag); commaIdx != -1 {
			name = yamlTag[:commaIdx]
		}

		if name == tag {
			return rv.Field(i), true
		}
	}

	return reflect.Value{}, false
}

// Helper to handle the "omitempty" or other options in tags
func findComma(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

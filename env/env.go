package env

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	BitwardenAccessToken   string
	BitwardenProjectID     string
	BitwardenStateFilePath string
}

// New returns an Env struct if all expected environmental variables are present
// otherwise it will return a zero-value and the list of missing environmental variables
func New() (Env, []string) {
	godotenv.Load()

	missing := make([]string, 0)

	bitwardenAccessToken := os.Getenv("BWS_ACCESS_TOKEN")
	if bitwardenAccessToken == "" {
		missing = append(missing, "BWS_ACCESS_TOKEN")
	}

	bitwardenProjectID := os.Getenv("BWS_PROJECT_ID")
	if bitwardenProjectID == "" {
		missing = append(missing, "BWS_PROJECT_ID")
	}

	bitwardenStateFilePath := os.Getenv("BWS_STATE_FILE_PATH")
	if bitwardenStateFilePath == "" {
		missing = append(missing, "BWS_STATE_FILE_PATH")
	}

	if len(missing) > 0 {
		return Env{}, missing
	}

	return Env{
		BitwardenAccessToken:   bitwardenAccessToken,
		BitwardenProjectID:     bitwardenProjectID,
		BitwardenStateFilePath: bitwardenStateFilePath,
	}, nil
}

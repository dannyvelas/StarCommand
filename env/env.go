package env

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	BitwardenAccessToken string
	BitwardenProjectID   string
}

// New returns an Env struct if all expected environmental variables are present
// otherwise it will return a zero-value and the list of missing environmental variables
func New() (Env, []string) {
	godotenv.Load()

	missing := make([]string, 0)

	token := os.Getenv("BWS_ACCESS_TOKEN")
	if token == "" {
		missing = append(missing, "BWS_ACCESS_TOKEN")
	}

	projectID := os.Getenv("BWS_PROJECT_ID")
	if projectID == "" {
		missing = append(missing, "BWS_PROJECT_ID")
	}

	if len(missing) > 0 {
		return Env{}, missing
	}

	return Env{
		BitwardenAccessToken: token,
		BitwardenProjectID:   projectID,
	}, nil
}

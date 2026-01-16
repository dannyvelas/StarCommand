package host

import "fmt"

var FallbackFile = "config/all.yml"

func GetConfigPath(hostAlias string) string {
	return fmt.Sprintf("config/%s.yml", hostAlias)
}

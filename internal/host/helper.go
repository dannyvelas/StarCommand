package host

import "fmt"

var FallbackFile = "config/all.yml"

func GetConfigPath(hostName string) string {
	return fmt.Sprintf("config/%s.yml", hostName)
}

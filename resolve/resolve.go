package resolve

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const fallbackConfigFile = "./configs/all.yml"

func ResolveConfig(verbose bool, hostName string) (map[string]string, error) {
	conf := map[string]string{
		"node_cidr_address":   "10.0.0.50/24",
		"gateway_address":     "10.0.0.1",
		"physical_nic":        "enx6c1ff7135975",
		"ssh_public_key_path": "~/.ssh/id_ed25519.pub",
	}

	hostConfigFile := fmt.Sprintf("./configs/%s.yml", hostName)
	for _, file := range []string{fallbackConfigFile, hostConfigFile} {
		data, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			if verbose {
				fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
			}
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error reading config file(%s): %v", file, err)
		}
		if err := yaml.Unmarshal(data, &conf); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
	}
	conf["node_ip"] = strings.Split(conf["node_cidr_address"], "/")[0]
	return conf, nil
}

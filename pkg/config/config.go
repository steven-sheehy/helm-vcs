package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/steven-sheehy/helm-vcs/pkg/chart"
	"github.com/steven-sheehy/helm-vcs/pkg/path"
)

type Config struct {
	APIVersion   string              `json:"apiVersion,omitempty"`
	Generated    time.Time           `json:"generated,omitempty"`
	Repositories []*chart.Repository `json:"repositories"`
}

func Load() (*Config, error) {
	configFile := configFile()
	_, err := os.Stat(configFile)

	if os.IsNotExist(err) {
		config := &Config{
			APIVersion:   "v1",
			Generated:    time.Now(),
			Repositories: []*chart.Repository{},
		}
		return config, nil
	}

	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configFile() string {
	return path.Home.ConfigFile()
}

func (c *Config) Repository(uri string) *chart.Repository {
	for _, r := range c.Repositories {
		if r.URI == uri || r.DisplayURI == uri {
			return r
		}
	}
	return nil
}

// Save writes the plugin configuration file
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile(), data, 0644)
}

package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/steven-sheehy/helm-vcs/pkg/chart"
)

type Config struct {
	APIVersion   string              `json:"apiVersion,omitempty"`
	Generated    time.Time           `json:"generated,omitempty"`
	Repositories []*chart.Repository `json:"repositories"`
	Path         string              `json:"-"`
}

func Load(path string) (*Config, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		config := &Config{
			APIVersion:   "v1",
			Generated:    time.Now(),
			Repositories: []*chart.Repository{},
			Path:         path,
		}
		return config, nil
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return nil, err
	}

	config.Path = path
	return config, nil
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
	return ioutil.WriteFile(c.Path, data, 0644)
}

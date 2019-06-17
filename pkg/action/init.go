package action

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/steven-sheehy/helm-vcs/pkg/chart"
	"github.com/steven-sheehy/helm-vcs/pkg/config"
)

type InitAction struct {
	Name   string
	Path   string
	Ref    string
	URI    string
	UseTag bool
}

func NewInitAction() *InitAction {
	action := &InitAction{}
	register(action)
	return action
}

func (a InitAction) Run() error {
	log.Infof("Running init action with args: %v", a)
	config, err := config.Load()
	if err != nil {
		return err
	}

	repository := config.Repository(a.URI)

	if repository == nil {
		repository, err = chart.NewRepository(a.Name, a.URI)
		if err != nil {
			return err
		}
		config.Repositories = append(config.Repositories, repository)
	}

	repository.Path = a.Path
	repository.Ref = a.Ref
	repository.UseTag = a.UseTag

	err = repository.Update()
	if err != nil {
		return err
	}

	return config.Save()
}

func (a InitAction) String() string {
	return fmt.Sprintf("{Name: '%v', URI: '%v', Path: '%v', Ref: '%v', UseTag: %v}", a.Name, a.URI, a.Path, a.Ref, a.UseTag)
}

func (a InitAction) Type() string {
	return "init"
}

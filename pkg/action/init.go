package action

import (
	"fmt"

	"github.com/steven-sheehy/helm-vcs/pkg/chart"
)

type InitAction struct {
	Action

	Name   string
	URI    string
	Path   string
	Ref    string
	UseTag bool
}

func NewInitAction() *InitAction {
	action := &InitAction{}
	register(action)
	return action
}

func (a InitAction) Run() error {
	repository, err := chart.NewRepository(a.Name, a.URI)
	if err != nil {
		return err
	}

	return repository.Update()
}

func (a InitAction) String() string {
	return fmt.Sprintf("{Name: '%v', URI: '%v', Path: '%v', Ref: '%v', UseTag: %v}", a.Name, a.URI, a.Path, a.Ref, a.UseTag)
}

func (a InitAction) Type() string {
	return "init"
}

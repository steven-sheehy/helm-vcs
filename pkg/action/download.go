package action

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/steven-sheehy/helm-vcs/pkg/config"
	"github.com/steven-sheehy/helm-vcs/pkg/path"
)

type DownloadAction struct {
	URI string
}

func NewDownloadAction() *DownloadAction {
	action := &DownloadAction{}
	register(action)
	return action
}

func (a DownloadAction) Run() error {
	log.SetLevel(log.ErrorLevel)
	config, err := config.Load(path.Home.ConfigFile())
	if err != nil {
		return err
	}

	uri := strings.TrimRight(a.URI, "/index.yaml")
	repository := config.Repository(uri)
	if repository == nil {
		return fmt.Errorf("Missing repository %v. Try first running `helm vcs init`", uri)
	}

	err = repository.Update()
	if err != nil {
		return err
	}

	index, err := repository.GetIndex()
	if err != nil {
		return err
	}

	fmt.Print(index)
	return nil
}

func (a DownloadAction) String() string {
	return fmt.Sprintf("{URI: %v}", a.URI)
}

func (a DownloadAction) Type() string {
	return "download"
}

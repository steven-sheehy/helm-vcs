package action

import (
	"fmt"
	"log"
	"os"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

type InitAction struct {
	Action

	Name string
	URI string
	Path string
	Ref string
	UseTag bool
}

func (a InitAction) Run() {
	err := a.addRepo()
        if err != nil {
                log.Fatal(err)
        }
}

func (a InitAction) addRepo() error {
	home := a.getHome()
	repoFile, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return errors.Wrap(err, "Unable to load repositories file")
	}

	repoEntry := &repo.Entry{
		Name: a.Name,
		URL: a.URI,
		Cache: fmt.Sprintf("%s-index.yaml", a.Name),
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(home.RepositoryFile(), 0644)

	if err != nil {
		return errors.Wrap(err, "Unable to write repositories file")
	}

	return nil
}

func (a InitAction) getHome() helmpath.Home {
	home := helmpath.Home(environment.DefaultHelmHome)

	envHome := os.Getenv("HELM_HOME")
	if envHome != "" {
		home = helmpath.Home(envHome)
	}

	return home
}


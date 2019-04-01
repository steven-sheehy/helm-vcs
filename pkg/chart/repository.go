package chart

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"github.com/Masterminds/vcs"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/plugin/cache"
	"k8s.io/helm/pkg/repo"
)

type Repository struct {
	Name     string
	vcsRepo  vcs.Repo
}

func NewRepository(name, uri string) (*Repository, error) {
	if uri == "" {
		return nil, errors.New("Missing required vcs URI")
	}

	if name == "" {
		projectName, err := getNameFromURI(uri)
		if err != nil {
			return nil, err
		}
		name = projectName
	}

	key, err := cache.Key(uri)
	if err != nil {
		return nil, err
	}

	local := getHome().Path("cache", "plugins", key, "repository", name)
	repo, err := vcs.NewRepo(uri, local)
	if err != nil {
		return nil, err
	}

	return &Repository{Name: name, vcsRepo: repo}, nil
}

func (r Repository) Update() error {
	home := getHome()
	repoFile, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return errors.Wrap(err, "Unable to load repositories file")
	}

	repoEntry := &repo.Entry{
		Name: r.Name,
		URL: r.vcsRepo.Remote(),
		Cache: fmt.Sprintf("%s-index.yaml", r.Name),
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(home.RepositoryFile(), 0644)

	if err != nil {
		return errors.Wrap(err, "Unable to write repositories file")
	}

	return nil
}

func getNameFromURI(uri string) (string, error) {
	name := filepath.Base(uri)

	if name == "" || name == "." || name == string(os.PathSeparator) {
		return "", errors.New("Unable to guess repository name from URI. Please explicitly provide a name")
	}

	if filepath.Ext(name) == ".git" {
		name = strings.TrimSuffix(name, ".git")
	}

	log.Printf("Extracted project name from URI: %v", name)
	return name, nil
}

func getHome() helmpath.Home {
	home := helmpath.Home(environment.DefaultHelmHome)

	envHome := os.Getenv("HELM_HOME")
	if envHome != "" {
		home = helmpath.Home(envHome)
	}

	return home
}


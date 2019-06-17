package chart

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	path "github.com/steven-sheehy/helm-vcs/pkg/path"

	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/repo"
)

var (
	chartFile    = "Chart.yaml"
	skippedFiles = map[string]bool{".git": true, ".svn": true}
)

// Repository represents a VCS backed Helm chart repository
type Repository struct {
	DisplayURI string `json:"displayURI"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Ref        string `json:"ref"`
	URI        string `json:"uri"`
	UseTag     bool   `json:"useTag"`
	vcsRepo    vcs.Repo
}

// NewRepository creates a chart repository
func NewRepository(name, uri string) (*Repository, error) {
	r := &Repository{
		Name: name,
		URI:  uri,
	}

	localPath := path.Home.Vcs(r.Name)
	vcsRepo, err := vcs.NewRepo(uri, localPath)
	if err != nil {
		return nil, err
	}

	r.vcsRepo = vcsRepo
	r.setDisplayURI(uri)
	return r, nil
}

func (r *Repository) GetIndex() (string, error) {
	chartDir := path.Home.Chart(r.Name)
	indexFile := filepath.Join(chartDir, "index.yaml")
	data, err := ioutil.ReadFile(indexFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *Repository) Reset() error {
	chartPath := path.Home.Chart(r.Name)
	_, err := os.Stat(chartPath)

	if os.IsNotExist(err) {
		return nil
	}

	return os.RemoveAll(chartPath)
}

// Update the chart repository by syncing it with the upstream VCS repo
func (r *Repository) Update() error {
	if _, err := os.Stat(r.vcsRepo.LocalPath()); os.IsNotExist(err) {
		log.Infof("Cloning %v", r.vcsRepo.LocalPath())
		err = r.vcsRepo.Get()
		if err != nil {
			return nil
		}
	} else {
		log.Infof("Updating remote repository")
		err = r.vcsRepo.Update()
		if err != nil {
			return err
		}
	}

	versions, err := r.versions()
	if err != nil {
		return err
	}

	chartsPath := path.Home.Chart(r.Name)
	startPath := r.vcsRepo.LocalPath() + r.Path + string(filepath.Separator)
	log.Debugf("Search for charts at relative path: '%v'", r.Path)

	err = r.Reset()
	if err != nil {
		return err
	}

	err = os.MkdirAll(chartsPath, 0755)
	if err != nil {
		return err
	}

	for _, version := range versions {
		log.Infof("Checking out %v", version)
		err = r.vcsRepo.UpdateVersion(version.Original())
		if err != nil {
			return err
		}

		err = filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Errorf("Unable to search path: %v", path)
			}

			if _, skipped := skippedFiles[info.Name()]; skipped {
				return filepath.SkipDir
			}

			if !info.IsDir() && info.Name() == chartFile {
				chartPath := filepath.Dir(path)
				chart, err := chartutil.LoadDir(chartPath)

				if err != nil {
					log.Errorf("Skipping invalid chart: %v", err)
					return nil
				}

				_, err = chartutil.Save(chart, chartsPath)
				if err != nil {
					log.Errorf("Unable to save chart: %v", err)
					return nil
				}

				log.Infof("Added new chart at %v", strings.TrimPrefix(chartPath, startPath))
			}

			return nil
		})

		if err != nil {
			log.Errorf("Error searching for charts: %v", err)
		}
	}

	index, err := repo.IndexDirectory(chartsPath, chartsPath)
	if err != nil {
		return err
	}

	if err := index.WriteFile(filepath.Join(chartsPath, "index.yaml"), 0644); err != nil {
		return err
	}

	repoFile, err := repo.LoadRepositoriesFile(path.Home.RepositoryFile())
	if err != nil {
		return errors.Wrap(err, "Unable to load repositories file")
	}

	repoEntry := &repo.Entry{
		Name:  r.Name,
		URL:   r.DisplayURI,
		Cache: path.Home.CacheIndex(r.Name),
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(path.Home.RepositoryFile(), 0644)
	if err != nil {
		return errors.Wrap(err, "Unable to write repositories file")
	}

	return nil
}

// Versions lists the semantic versions found in this repository
func (r *Repository) versions() ([]*semver.Version, error) {
	tags, err := r.vcsRepo.Tags()
	if err != nil {
		return nil, err
	}

	var versions []*semver.Version
	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err == nil {
			versions = append(versions, v)
		}
	}

	sort.Sort(semver.Collection(versions))
	log.Debugf("Found versions: %v", versions)
	return versions, nil
}

func (r *Repository) setDisplayURI(uri string) {
	i := strings.Index(uri, "://")
	if i > -1 {
		uri = uri[i+3:]
	}
	r.DisplayURI = string(r.vcsRepo.Vcs()) + "://" + uri
}

// UnmarshalJSON callback to load internal objects after unmarshalling
func (r *Repository) UnmarshalJSON(b []byte) error {
	type repositoryJSON Repository
	if err := json.Unmarshal(b, (*repositoryJSON)(r)); err != nil {
		return err
	}

	localPath := path.Home.Vcs(r.Name)
	vcsRepo, err := vcs.NewRepo(r.URI, localPath)
	if err != nil {
		return err
	}

	r.vcsRepo = vcsRepo
	return nil
}

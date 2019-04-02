package chart

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	log "github.com/sirupsen/logrus"
)

var (
	chartFile       = "Chart.yaml"
	ignoredSuffixes = [5]string{"/", ".git", "/branches", "/tags", "/trunk"}
	pluginName      = "helm-vcs"
	skippedFiles    = map[string]int{".git": 1, ".svn": 1}
)

type Repository struct {
	Name    string
	Path    string
	vcsRepo vcs.Repo
}

func NewRepository(name, uri string) (*Repository, error) {
	if uri == "" {
		return nil, errors.New("Missing required vcs URI")
	}

	if name == "" {
		projectName, err := projectName(uri)
		if err != nil {
			return nil, err
		}
		name = projectName
	}

	local := PluginHome().Path("repository", name)
	repo, err := vcs.NewRepo(uri, local)
	if err != nil {
		return nil, err
	}

	return &Repository{Name: name, vcsRepo: repo}, nil
}

func (r Repository) Update() error {
	home := HelmHome()
	repoFile, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return errors.Wrap(err, "Unable to load repositories file")
	}

	repoEntry := &repo.Entry{
		Name: r.Name,
		URL: r.vcsRepo.Remote(),
		Cache: home.CacheIndex(r.Name),
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(home.RepositoryFile(), 0644)

	if err != nil {
		return errors.Wrap(err, "Unable to write repositories file")
	}

	if _, err = os.Stat(r.vcsRepo.LocalPath()); os.IsNotExist(err) {
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

	versions, err := r.Versions()
	if err != nil {
		return err
	}

	charts := make(map[string]*chart.Chart)
	startPath := r.vcsRepo.LocalPath() + r.Path + string(filepath.Separator)
	log.Infof("Search for charts at path: %v", startPath)

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

				chartKey := chart.GetMetadata().GetName() + chart.GetMetadata().GetVersion()
				if _, exists := charts[chartKey]; !exists {
					charts[chartKey] = chart
					log.Infof("Added new chart at %v", strings.TrimPrefix(path, startPath))
				}
			}

			return nil
		})

		if err != nil {
			log.Errorf("Error searching for charts: %v", version, err)
		}
	}

//	if err := idx.WriteFile(repoEntry.Cache, 0644); err != nil {
//		return err
//	}

	return nil
}

func (r Repository) Versions() ([]*semver.Version, error) {
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
	log.Infof("Found versions: %v", versions)
	return versions, nil
}

func projectName(uri string) (string, error) {
	name := uri

	for _, ignoredSuffix := range ignoredSuffixes {
		name = strings.TrimSuffix(name, ignoredSuffix)
	}

	name = filepath.Base(name)

	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", errors.New("Unable to get repository name from URI. Please explicitly provide a name")
	}

	log.Infof("Extracted project name '%v' from URI", name)
	return name, nil
}

func HelmHome() helmpath.Home {
	home := helmpath.Home(environment.DefaultHelmHome)

	envHome := os.Getenv("HELM_HOME")
	if envHome != "" {
		home = helmpath.Home(envHome)
	}

	return home
}

func PluginHome() helmpath.Home {
	return helmpath.Home(helmpath.Home(HelmHome().Plugins()).Path(pluginName))
}


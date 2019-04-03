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

	r := &Repository{}

	if name == "" {
		projectName, err := projectName(uri)
		if err != nil {
			return nil, err
		}
		name = projectName
	}
	r.Name = name

	localPath := r.path("vcs")
	repo, err := vcs.NewRepo(uri, localPath)
	if err != nil {
		return nil, err
	}
	r.vcsRepo = repo

	return r, nil
}

func (r Repository) path(subPath string) string {
	return HelmHome().Path("plugins", pluginName, "repository", r.Name, subPath)
}

func (r Repository) Reset() error {
	chartPath := r.path("chart")
	_, err := os.Stat(chartPath)

	if os.IsNotExist(err) {
		return nil
	}

	return os.RemoveAll(chartPath)
}

func (r Repository) Update() error {
	home := HelmHome()

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

	versions, err := r.Versions()
	if err != nil {
		return err
	}

	chartsPath := r.path("chart")
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

	repoFile, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return errors.Wrap(err, "Unable to load repositories file")
	}

	repoURL := r.vcsRepo.Remote()
	protocol := string(r.vcsRepo.Vcs()) + "://"
	if !strings.HasPrefix(repoURL, protocol) {
		repoURL = protocol + repoURL
	}

	repoEntry := &repo.Entry{
		Name: r.Name,
		URL: repoURL,
		Cache: home.CacheIndex(r.Name),
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(home.RepositoryFile(), 0644)
	if err != nil {
		return errors.Wrap(err, "Unable to write repositories file")
	}

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
	log.Debugf("Found versions: %v", versions)
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


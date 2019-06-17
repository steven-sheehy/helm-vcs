package path

import (
	"os"

	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
)

type Path struct {
	helmpath.Home
}

const plugin = "helm-vcs"

var Home = newHome()

func newHome() Path {
	home := helmpath.Home(environment.DefaultHelmHome)

	envHome := os.Getenv("HELM_HOME")
	if envHome != "" {
		home = helmpath.Home(envHome)
	}

	return Path{home}
}

func (p Path) ConfigFile() string {
	return p.Path("plugins", plugin, "vcs.yaml")
}

func (p Path) Chart(name string) string {
	return p.Path("plugins", plugin, "repository", name, "chart")
}

func (p Path) Vcs(name string) string {
	return p.Path("plugins", plugin, "repository", name, "vcs")
}

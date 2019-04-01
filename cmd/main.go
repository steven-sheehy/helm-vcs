package main

import (
	"os"
	"github.com/steven-sheehy/helm-vcs/pkg/action"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("helm vcs", "Allows any version control system (bzr, git, hg, svn) to be used as a Helm chart repository")

	initAction := action.InitAction{}
	init := app.Command("init", "Initialize the chart repository using the VCS repository as its source")
	init.Arg("uri", "The VCS URI").Required().StringVar(&initAction.URI)
	init.Flag("name", "The chart repository name. By default it will guess it from the URI").StringVar(&initAction.Name)
	init.Flag("path", "A path within the repository that contains charts").StringVar(&initAction.Path)
	init.Flag("ref", "A specific tag, branch or commit to checkout").StringVar(&initAction.Ref)
	init.Flag("use-tag", "Override the Chart.yaml version with the VCS tag").BoolVar(&initAction.UseTag)

	downloadAction := action.DownloadAction{}
	download := app.Command("download", "Download a file from the VCS repo. This is an internal command for use by Helm")
	download.Arg("certificate", "The certificate file to use").Required().String()
	download.Arg("key", "The private key to use").Required().String()
	download.Arg("CA", "The Certificate Authority to use").Required().String()
	download.Arg("uri", "The URI to download").Required().StringVar(&downloadAction.URI)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
		case init.FullCommand():
			initAction.Run()
		case download.FullCommand():
			downloadAction.Run()
	}
}


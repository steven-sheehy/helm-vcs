package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/steven-sheehy/helm-vcs/pkg/action"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	initAction := action.NewInitAction()
	downloadAction := action.NewDownloadAction()

	// Workaround Helm downloader not supporting sub-commands
	if len(os.Args) == 5 && action.Find(os.Args[1]) == nil {
		os.Args = append(os.Args[:1], append([]string{downloadAction.Type()}, os.Args[1:]...)...)
	}

	app := kingpin.New("helm vcs", "Turns any existing version control repository into a chart repository")

	init := app.Command(initAction.Type(), "Initialize the chart repository using the VCS repository as its source")
	init.Arg("name", "The chart repository name").Required().StringVar(&initAction.Name)
	init.Arg("uri", "The VCS URI").Required().StringVar(&initAction.URI)
	init.Flag("path", "A path within the repository that contains charts").StringVar(&initAction.Path)
	init.Flag("ref", "A specific tag, branch or commit to checkout").StringVar(&initAction.Ref)
	init.Flag("use-tag", "Override the Chart.yaml version with the VCS tag").BoolVar(&initAction.UseTag)

	download := app.Command(downloadAction.Type(), "Download a file from the VCS repo. This is an internal command for use by Helm")
	download.Arg("certificate", "The certificate file to use").Required().String()
	download.Arg("key", "The private key to use").Required().String()
	download.Arg("ca", "The Certificate Authority to use").Required().String()
	download.Arg("uri", "The URI to download").Required().StringVar(&downloadAction.URI)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	action := action.Find(command)
	if action == nil {
		log.Fatalf("Unknown command: %s", command)
	}

	err := action.Run()
	if err != nil {
		log.Fatal(err)
	}
}

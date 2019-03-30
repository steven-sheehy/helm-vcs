package main

import (
	"fmt"
	"os"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("helm vcs", "Allows any version control system (bzr, git, hg, svn) to be used as a Helm chart repository")
	action := kingpin.MustParse(app.Parse(os.Args[1:]))
	fmt.Printf("action: %s", action)
}


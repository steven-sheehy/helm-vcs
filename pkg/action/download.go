package action

import (
	"fmt"
)

type DownloadAction struct {
	Action

	URI string
}

func (a DownloadAction) Run() error {
	fmt.Printf("apiVersion: v1\nentries: {}\n")
	return nil
}

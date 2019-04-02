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

func (a DownloadAction) String() string {
	return fmt.Sprintf("{URI: %v}", a.URI)
}


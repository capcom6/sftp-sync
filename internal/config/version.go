package config

import (
	"fmt"
	"os"
)

const notSet string = "not set"

// these information will be collected when build, by `-ldflags "-X main.appVersion=0.1"`.
//
//nolint:gochecknoglobals // build metadata
var (
	appVersion = notSet
	buildTime  = notSet
	gitCommit  = notSet
	gitRef     = notSet
)

func printVersion() {
	fmt.Fprintf(os.Stdout, "Version:    %s\n", appVersion)
	fmt.Fprintf(os.Stdout, "Build Time: %s\n", buildTime)
	fmt.Fprintf(os.Stdout, "Git Commit: %s\n", gitCommit)
	fmt.Fprintf(os.Stdout, "Git Ref:    %s\n", gitRef)
}

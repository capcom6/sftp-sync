package config

import "fmt"

const notSet string = "not set"

// these information will be collected when build, by `-ldflags "-X main.appVersion=0.1"`
var (
	appVersion = notSet
	buildTime  = notSet
	gitCommit  = notSet
	gitRef     = notSet
)

func printVersion() {
	fmt.Printf("Version:    %s\n", appVersion)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Git Commit: %s\n", gitCommit)
	fmt.Printf("Git Ref:    %s\n", gitRef)
}

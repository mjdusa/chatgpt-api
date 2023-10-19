package version

import (
	"fmt"
	"os"
)

var (
	// AppVersion contains the current version in SemVer format.
	AppVersion string //nolint:gochecknoglobals  // only used on main for usage

	// Branch is the name of the branch referenced by HEAD.
	Branch string //nolint:gochecknoglobals  // only used on main for usage

	// BuildTime is the compiled build time.
	BuildTime string //nolint:gochecknoglobals  // only used on main for usage

	// Commit contains the hash of the latest commit on Branch.
	Commit string //nolint:gochecknoglobals  // only used on main for usage

	// GoVersion contains the the version of the go that performed the build.
	GoVersion string //nolint:gochecknoglobals  // only used on main for usage
)

func GetVersion() string {
	msg := fmt.Sprintf("%s version: [%s]\n", os.Args[0], AppVersion)
	msg += fmt.Sprintf("- Branch:     [%s]\n", Branch)
	msg += fmt.Sprintf("- Build Time: [%s]\n", BuildTime)
	msg += fmt.Sprintf("- Commit:     [%s]\n", Commit)
	msg += fmt.Sprintf("- Go Version: [%s]\n", GoVersion)

	return msg
}

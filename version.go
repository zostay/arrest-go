package arrest

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var versionFile string

// Version is the arrest-go release version, taken from version.txt at build
// time. It is the bare X.Y.Z form; the matching git tag is vX.Y.Z.
var Version = strings.TrimSpace(versionFile)

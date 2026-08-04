// Package testkit carries the authored conformance cases, embedded so
// that one kit binary is one complete artifact: the runner and the
// cases it judges by can never drift apart in a release.
package testkit

import (
	"embed"
	"io/fs"
)

//go:embed cases
var embedded embed.FS

// Cases is the authored case tree, rooted so that case names read as
// their paths under cases/.
func Cases() fs.FS {
	sub, err := fs.Sub(embedded, "cases")
	if err != nil {
		// The subtree is embedded at compile time; failing to root it is
		// unreachable short of a build defect.
		panic(err)
	}
	return sub
}

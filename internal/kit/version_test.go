package kit

import (
	"strings"
	"testing"
)

// The kit's version tracks the spec version its embedded cases target:
// KitVersion's x.y is SpecVersion, and the last digit counts the kit's
// own fixes within that target.
func TestKitVersionTracksSpecVersion(t *testing.T) {
	if !strings.HasPrefix(KitVersion, SpecVersion+".") {
		t.Fatalf("KitVersion %s does not target SpecVersion %s", KitVersion, SpecVersion)
	}
}

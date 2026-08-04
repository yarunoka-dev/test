package kit

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Case is one authored conformance case: a request, the expected
// response, and the metadata that FAIL reporting shows (what the case
// checks and which normative text it comes from).
type Case struct {
	Name        string
	Description string   `json:"description"`
	Spec        string   `json:"spec"`
	Request     Request  `json:"request"`
	Expected    Response `json:"response"`
}

// LoadCases reads every *.json under the root of fsys into a Case, named
// by its path without the extension, sorted by name so runs are
// deterministic.
func LoadCases(fsys fs.FS) ([]Case, error) {
	var cases []Case

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		c := Case{Name: strings.TrimSuffix(path, ".json")}
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("case %s: %w", path, err)
		}
		if c.Expected.Result == nil && !c.Expected.Invalid {
			return fmt.Errorf("case %s: the expected response carries neither a result nor invalid", path)
		}

		cases = append(cases, c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].Name < cases[j].Name })
	return cases, nil
}

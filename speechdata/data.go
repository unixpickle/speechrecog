package speechdata

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

const (
	IndexFilename = "index.json"
	IndexPerms    = 0755
)

// A Sample stores information about one audio sample.
type Sample struct {
	ID    string
	Label string
	File  string
}

// An Index is a listing of a bunch of samples and their
// enclosing directory.
type Index struct {
	Samples []Sample
	DirPath string `json:"-"`
}

// LoadIndex loads an index from a data directory.
func LoadIndex(dbPath string) (*Index, error) {
	contents, err := ioutil.ReadFile(filepath.Join(dbPath, IndexFilename))
	if err != nil {
		return nil, err
	}
	var ind Index
	if err := json.Unmarshal(contents, &ind); err != nil {
		return nil, err
	}
	ind.DirPath = dbPath
	return &ind, nil
}

// Save saves the index to its data directory.
func (i *Index) Save() error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(i.DirPath, IndexFilename), data, IndexPerms)
}

// Clone creates a copy of this index, which should be
// treated as a read-only copy useful for printing
// directory listings.
func (i *Index) Clone() *Index {
	res := &Index{DirPath: i.DirPath}
	for _, x := range i.Samples {
		res.Samples = append(res.Samples, x)
	}
	return res
}

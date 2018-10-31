package local

import (
	"io"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Store struct {
	logger  *log.Logger
	url     string
	root    string
	collMap map[string]*Coll
}

// NewStore returns a new Store for image collections stored
// on the local file system.
// A local store has a collection for each directory, where the
// name of the directory is the name of the collection.
func NewStore(url, root string, logger *log.Logger) (*Store, error) {
	s := Store{
		logger:  logger,
		url:     url,
		root:    root,
		collMap: make(map[string]*Coll),
	}

	rootDir, err := os.OpenFile(root, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return nil, err
	}

	for err == nil {
		var fis []os.FileInfo
		fis, err = rootDir.Readdir(10)
		for _, fi := range fis {
			if fi.IsDir() {
				name := fi.Name()
				s.collMap[name] = &Coll{
					path: filepath.Join(root, name),
					url:  path.Join(url, name),
				}
			}
		}
	}
	if err != io.EOF {
		return nil, err
	}

	return &s, nil
}

func (s *Store) Collections() (*CollList, error) {
	cl := CollList{colls: make([]*Coll, len(s.collMap))}
	var i int
	for _, c := range s.collMap {
		cl.colls[i] = c
		i++
	}
	return &cl, nil
}

type CollList struct {
	colls []*Coll
	i     int
}

func (l *CollList) ReadList(n int) ([]*Coll, error) {
	if l.i >= len(l.colls) {
		return nil, io.EOF
	}

	high := len(l.colls)
	if n > 0 && l.i+n < len(l.colls) {
		high = l.i + n
	}
	colls := make([]*Coll, high-l.i)
	copy(colls, l.colls[l.i:high])
	l.i = high

	if n <= 0 {
		return colls, nil
	}

	var err error
	if l.i >= len(l.colls) {
		err = io.EOF
	}
	return colls, err
}

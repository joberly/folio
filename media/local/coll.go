package local

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var extRegex = regexp.MustCompile(`(?i)\.((?:gif)|(?:jpg)|(?:png))$`)

type Coll struct {
	path string // OS path to collection directory
	url  string // URL to collection images
}

// Name returns the collection's name.
func (c *Coll) Name() string {
	return filepath.Base(c.path)
}

// Images returns an ImageList for iterating through
// images in the collection
func (c *Coll) Images() (*ImageList, error) {
	dir, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}

	il := ImageList{
		path:  c.path,
		url:   c.url,
		names: make([]string, 0, 10),
	}

	for err == nil {
		var fis []os.FileInfo
		fis, err = dir.Readdir(10)
		for _, fi := range fis {
			if !fi.IsDir() && extRegex.Match([]byte(fi.Name())) {
				il.names = append(il.names, fi.Name())
			}
		}
	}
	if err != io.EOF {
		return nil, err
	}

	return &il, nil
}

// imageList is the image collection directory or iterator.
type ImageList struct {
	path  string
	url   string
	names []string
	i     int
}

// ReadList returns a slice of images a collection.
func (l *ImageList) ReadList(n int) ([]*Image, error) {
	if l.i == len(l.names) {
		return nil, io.EOF
	}

	high := len(l.names)
	if n > 0 && l.i+n < len(l.names) {
		high = l.i + n
	}

	imgs := make([]*Image, high-l.i)
	for i := 0; i < len(imgs); i++ {
		imgs[i] = &Image{
			name: l.names[l.i],
			path: l.path,
			url:  l.url,
		}
		l.i++
	}

	if n <= 0 {
		return imgs, nil
	}

	var err error
	if l.i == len(l.names) {
		err = io.EOF
	}
	return imgs, err
}

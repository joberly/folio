package local

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

type Image struct {
	name string // Name of the image file
	path string // Path to the image file in the local filesystem not including the file name
	url  string // URL path to the image file not including the file name
}

func (img *Image) Name() string {
	return img.name
}

func (img *Image) URL() string {
	return path.Join(img.url, img.name)
}

func (img *Image) Reader() (io.ReadCloser, error) {
	return os.Open(filepath.Join(img.path, img.name))
}

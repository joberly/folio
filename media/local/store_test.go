package local

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var randomStores = flag.Int("randomstores", 10, "total number of randomly generated stores to test")
var randomColls = flag.Int("randomcolls", 5, "maximum number of randomly generated collections per store")
var randomFiles = flag.Int("randomfiles", 10, "maximum number of randomly generated files per collection")

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Test store info
type storeInfo struct {
	name  string
	url   string
	colls map[string]*collInfo
}

// Test collection info
type collInfo struct {
	name   string
	path   string
	url    string
	images []string
	others []string
}

func TestStore(t *testing.T) {
	// Create a temp directory for all of the test stores
	tdir, err := ioutil.TempDir("", "folio")
	if err != nil {
		t.Fatal(err)
		return
	}

	baseURL := "https://localhost/folio/test"

	// Test stores
	stores := make([]storeInfo, 1, 1+*randomStores)

	// Static store
	stores[0] = storeInfo{
		name: "static-1",
		url:  baseURL,
		colls: map[string]*collInfo{
			"coll-a": &collInfo{
				name:   "coll-a",
				images: []string{"a2i.jpg", "a3i.gif", "a5i.png"},
				others: []string{"a1n.txt", "a4n.json"},
			},
			"coll-b": &collInfo{
				name:   "coll-b",
				images: []string{"b3i.jpg", "b5i.gif", "b6i.jpg", "b7i.png"},
				others: []string{"b2n.md", "b4n", "b8n.data"},
			},
			"coll-c": &collInfo{
				name:   "coll-c",
				images: []string{"c1i.gif", "c4i.png", "c5i.jpg", "c8n.jpg", "c9i.png"},
				others: []string{"c2n", "c3n", "c6n.c", "c7n.go", "c10n.jp", "c11n.pn", "c12n.gig"},
			},
		},
	}
	for _, c := range stores[0].colls {
		c.path = filepath.Join(tdir, stores[0].name, c.name)
		c.url = path.Join(baseURL, c.name)
	}

	// Add randomly generated collections
	/*
		for s := 0; s < *randomStores; s++ {
			si := storeInfo{
				name:  fmt.Sprintf("random-%d", s),
				colls: make(map[string]collInfo),
			}
			for ci := 0; ci < rnd.Intn(*randomColls)+1; ci++ {
				numimg := rnd.Intn(*randomFiles)+1
				numoth := rnd.Intn(*randomFiles - numimg + 1)
				c := collInfo{
					name: randString(10),
					images: make([]string, 0, numimg),
					others: make([]string, 0, numoth),
				}
				for ii := 0; ii < numimg; ii++ {

				}
				for oi := 0; oi < numoth; oi++ {

				}
			}
		}
	*/

	// Run test on each store
	for _, si := range stores {
		t.Run(si.name, func(t *testing.T) {
			testSingleStore(t, tdir, &si)
		})
	}
}

func testSingleStore(t *testing.T, tdir string, si *storeInfo) {
	// Create store directory
	sdir := filepath.Join(tdir, si.name)
	err := os.Mkdir(sdir, 0770)
	// Create collections
	for _, ci := range si.colls {
		//t.Logf("create %s/%s\n", si.name, ci.name)
		// Create collection directory
		cdir := filepath.Join(sdir, ci.name)
		err = os.Mkdir(cdir, 0770)
		if err != nil {
			t.Fatal(err)
		}
		// Create fake image files
		for _, img := range ci.images {
			// Create image file
			f, err := os.Create(filepath.Join(cdir, img))
			if err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
		// Create all other fake files
		for _, oth := range ci.others {
			// Create other non-image file
			f, err := os.Create(filepath.Join(cdir, oth))
			if err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
	}

	// Create the store struct
	logger := log.New()
	logger.SetOutput(ioutil.Discard)
	var s *Store
	s, err = NewStore(si.url, sdir, logger)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Validate some internals of the Store
	if len(si.colls) != len(s.collMap) {
		t.Errorf("collMap size mismatch actual %d - expected %d",
			len(s.collMap), len(si.colls))
	}
	expPath := filepath.Join(tdir, si.name)
	if s.root != expPath {
		t.Errorf("root mismatch actual %s - expected %s",
			s.root, expPath)
	}

	var cl *CollList
	//t.Logf("%s len collMap %d\n", si.name, len(s.collMap))
	cl, err = s.Collections()
	if err != nil {
		t.Fatal(err)
		return
	}

	// Validate internal CollList data
	if len(cl.colls) != len(s.collMap) {
		t.Errorf("collection list slice size mismatch actual %d - expected %d",
			len(cl.colls), len(s.collMap))
	}

	cs, _ := cl.ReadList(0)
	if len(cs) != len(si.colls) {
		t.Errorf("collection list size %d mismatch (expected %d)",
			len(cs), len(si.colls))
	}

	for _, c := range cs {
		ci := si.colls[c.Name()]
		t.Run(c.Name(), func(t *testing.T) {
			testSingleColl(t, ci, c)
		})
	}
}

func testSingleColl(t *testing.T, ci *collInfo, c *Coll) {
	if ci == nil {
		t.Errorf("no test collInfo for collection")
		return
	}

	if c.Name() != ci.name {
		t.Errorf("unexpected name %s (expected %s)",
			c.Name(), ci.name)
	}

	if c.path != ci.path {
		t.Errorf("unexpected path %s (expected %s)",
			c.path, ci.path)
	}

	if c.url != ci.url {
		t.Errorf("unexpected URL %s (expected %s)",
			c.url, ci.url)
	}

	il, err := c.Images()
	if err != nil {
		t.Errorf("no image list")
		return
	}

	// Validate internals
	if len(il.names) != len(ci.images) {
		t.Errorf("image list names size mismatch actual %d expected %d",
			len(il.names), len(ci.images))
	}

	var imgs []*Image
	imgs, err = il.ReadList(0)
	if err != nil {
		t.Errorf("cannot get images from image list")
		return
	}

	if len(imgs) != len(ci.images) {
		t.Errorf("unexpected num images %d (expected %d)",
			len(imgs), len(ci.images))
	}

	// Create an image map for testing
	imgMap := make(map[string]bool)
	for _, name := range ci.images {
		imgMap[name] = true
	}

	for _, img := range imgs {
		// Look for entry in map
		name := img.Name()
		if _, ok := imgMap[name]; !ok {
			t.Errorf("unexpected image name %s", img.Name())
		}
		expURL := path.Join(ci.url, name)
		if img.URL() != expURL {
			t.Errorf("unexpected image URL %s (expected %s)",
				img.URL(), expURL)
		}
		if img.path != ci.path {
			t.Errorf("unexpected image path %s (expected %s)",
				img.path, ci.path)
		}
	}
}

const randStringRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randStringRunes[rnd.Intn(len(randStringRunes))]
	}
	return string(b)
}

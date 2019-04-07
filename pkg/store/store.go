package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

type Store interface {
	Get(key string) ([]byte, error)
	Put(key string, b []byte) error
}

type Storage struct {
	mutex     sync.Mutex
	mutexes   map[string]*sync.Mutex
	dir       string
	log       *logrus.Entry
	namespace string
}

var _ Store = &Storage{}

func New(log *logrus.Entry, dir, namespace string) (Store, error) {
	dir = filepath.Clean(dir)

	s := &Storage{
		dir:       dir,
		log:       log,
		mutexes:   make(map[string]*sync.Mutex),
		namespace: namespace,
	}

	// if the database already exists, just use it
	if _, err := os.Stat(dir); err == nil {
		s.log.Debugf("Using '%s' (database already exists)", dir)
		return s, nil
	}

	// if the database doesn't exist create it
	s.log.Debugf("Creating database at '%s'", dir)
	return s, os.MkdirAll(dir, 0755)
}

func (s *Storage) Put(key string, b []byte) error {
	if key == "" {
		return fmt.Errorf("missing key - unable to save")
	}

	mutex := s.getMutex(s.namespace)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(s.dir, s.namespace)
	fnlPath := filepath.Join(dir, key+".json")
	tmpPath := fnlPath + ".tmp"

	// create collection directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// write marshaled data to the temp file
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	// move final file into place
	return os.Rename(tmpPath, fnlPath)
}

// Get a record from the database
func (s *Storage) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("missing key - unable to read")
	}

	record := filepath.Join(s.dir, s.namespace, key)

	// check to see if file exists
	if _, err := stat(record); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(record + ".json")
}

func (s *Storage) Delete(key string) error {
	path := filepath.Join(s.namespace, key)
	//
	mutex := s.getMutex(s.namespace)
	mutex.Lock()
	defer mutex.Unlock()

	//
	dir := filepath.Join(s.dir, path)

	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v", path)
	case fi.Mode().IsDir():
		return os.RemoveAll(dir)
	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil
}

func stat(path string) (fi os.FileInfo, err error) {
	// check for dir, if path isn't a directory check to see if it's a file
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

func (s *Storage) getMutex(collection string) *sync.Mutex {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	m, ok := s.mutexes[collection]
	// if the mutex doesn't exist make it
	if !ok {
		m = &sync.Mutex{}
		s.mutexes[collection] = m
	}
	return m
}

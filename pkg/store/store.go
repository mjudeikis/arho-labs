package store

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/api"
)

type Store interface {
	Get() (*api.Credential, error)
}

type Storage struct {
	storageDir string
	sync.Mutex
	log *logrus.Entry
}

var _ Store = &Storage{}

var credentialsFile = "credentials.yaml"

func NewStore(log *logrus.Entry, storageDir string) Storage {
	return Storage{
		storageDir: storageDir,
		log:        log,
	}
}

func (s *Storage) Get() (*api.Credential, error) {
	s.Lock()
	defer s.Unlock()
	s.log.Debugf("storage: Get %s/%s", s.storageDir, credentialsFile)
	bytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", s.storageDir, credentialsFile))
	if err != nil {
		return nil, err
	}
	var cs api.CredentialsStore
	err = yaml.Unmarshal(bytes, &cs)
	if err != nil {
		return nil, err
	}

	var result *api.Credential
	for key, cred := range cs.Credentials {
		if !cred.Reserved {
			result = &cred
			cs.Credentials[key].Reserved = true
			break
		}
	}

	bytes, err = yaml.Marshal(cs)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", s.storageDir, credentialsFile), bytes, 0600)
	if err != nil {
		return nil, err
	}

	return result, nil
}

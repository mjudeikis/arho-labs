package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/api"
	"github.com/mjudeikis/osa-labs/pkg/store"
	"github.com/mjudeikis/osa-labs/pkg/workers"
)

var lock sync.Mutex

type Server struct {
	store         store.Store
	log           *logrus.Entry
	address       string
	hostname      string
	workerManager workers.Workers
	devMode       bool
}

type setup struct {
	hostname string
}

func New(log *logrus.Entry, devMode bool, hostname, address, workerImage string, workerNumber int) (*Server, error) {
	store, err := store.New(log, "storage", "credentials")
	if err != nil {
		return nil, err
	}

	wm, err := workers.New(log, workerImage, workerNumber)
	if err != nil {
		return nil, err
	}

	server := &Server{
		log:           log,
		address:       address,
		hostname:      hostname,
		store:         store,
		workerManager: wm,
		devMode:       devMode,
	}
	return server, nil
}

func (s *Server) Run() error {
	if s.devMode {
		s.dummyData()
	}

	// init workers
	err := s.workerManager.Create()
	if err != nil {
		return err
	}

	http.HandleFunc("/", s.index)
	http.HandleFunc("/setup", s.getSetup)
	http.HandleFunc("/credentials", s.getCredentials)
	http.HandleFunc("/worker", s.getWorker)

	log.Printf("Listening on %s", s.address)
	return http.ListenAndServe(s.address, nil)
}

func (s *Server) getCredentials(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("getCredentials")

	result, err := s.getUniqueCredential()
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
	}

	var res []byte
	res, err = json.Marshal(result)
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (s *Server) getWorker(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("getWorker")

	result, err := s.getUniqueWorker()
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
	}

	var res []byte
	res, err = json.Marshal(result)
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func (s *Server) getSetup(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("getSetup")
	t, err := template.New("setup.sh").ParseFiles("template/setup.sh")
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}

	tmpl := struct {
		Hostname string
	}{
		Hostname: s.hostname,
	}

	err = t.Execute(w, tmpl)
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}

}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("index")
	t, err := template.New("index.html").ParseFiles("template/index.html")
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}

	tmpl := struct {
		Hostname string
	}{
		Hostname: s.hostname,
	}

	err = t.Execute(w, tmpl)
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}
}

func (s *Server) getUniqueCredential() (*api.Credential, error) {
	lock.Lock()
	defer lock.Unlock()
	data, err := s.store.Get("credentials")
	if err != nil {
		return nil, err
	}
	var result *api.Credential
	var credentialStore api.CredentialsStore

	err = yaml.Unmarshal(data, &credentialStore)
	if err != nil {
		return nil, err
	}

	for key, cred := range credentialStore.Credentials {
		if !cred.Reserved {
			credentialStore.Credentials[key].Reserved = true
			result = &cred
			break
		}
	}

	data, err = yaml.Marshal(credentialStore)
	if err != nil {
		return nil, err
	}

	err = s.store.Put("credentials", data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) getUniqueWorker() (*api.Worker, error) {
	lock.Lock()
	defer lock.Unlock()
	data, err := s.store.Get("workers")
	if err != nil {
		return nil, err
	}
	var result *api.Worker
	var workerStore api.WorkersStore

	err = yaml.Unmarshal(data, &workerStore)
	if err != nil {
		return nil, err
	}

	for key, wk := range workerStore.Workers {
		if !wk.Reserved {
			workerStore.Workers[key].Reserved = true
			result = &wk
			break
		}
	}

	data, err = yaml.Marshal(workerStore)
	if err != nil {
		return nil, err
	}

	err = s.store.Put("workers", data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) dummyData() {
	// dummy code to produce credentials file
	var cs api.CredentialsStore
	for i := 1; i <= 50; i++ {
		cs.Credentials = append(cs.Credentials, api.Credential{
			Username: "username" + strconv.Itoa(i),
			Password: "password" + strconv.Itoa(i),
			Reserved: false,
		})
	}

	bytes, err := yaml.Marshal(cs)
	if err != nil {
		panic(err)
	}
	s.store.Put("credentials", bytes)

	// dummy code to produce credentials file
	var wk api.WorkersStore
	for i := 1; i <= 50; i++ {
		wk.Workers = append(wk.Workers, api.Worker{
			IP:       "1.1.1.1",
			SSHKey:   "dummy ssh key",
			Reserved: false,
		})
	}

	bytes, err = yaml.Marshal(wk)
	if err != nil {
		panic(err)
	}
	s.store.Put("workers", bytes)

}

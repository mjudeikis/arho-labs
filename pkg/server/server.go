package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/store"
)

type Server struct {
	store    store.Storage
	log      *logrus.Entry
	address  string
	hostname string
}

type setup struct {
	hostname string
}

func New(log *logrus.Entry, hostname string, address string) *Server {
	s := &Server{
		log:      log,
		address:  address,
		hostname: hostname,
		store:    store.NewStore(log, "storage"),
	}
	return s
}

func (s *Server) Run() {
	http.HandleFunc("/", s.index)
	http.HandleFunc("/setup", s.getSetup)
	http.HandleFunc("/credentials", s.getCredentials)

	log.Printf("Listening on %s", s.address)
	http.ListenAndServe(s.address, nil)
}

func (s *Server) getCredentials(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("getCredentials")

	cred, err := s.store.Get()
	if err != nil {
		s.log.Error(err)
		resp := fmt.Sprintf("500 Internal Error: %s", err)
		http.Error(w, resp, http.StatusInternalServerError)
	}

	var res []byte
	res, err = json.Marshal(cred)
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

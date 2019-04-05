package main

import (
	"flag"
	"io/ioutil"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/api"
	"github.com/mjudeikis/osa-labs/pkg/server"
)

var (
	devMode  = flag.Bool("dev-mode", false, "If set, dummy files will be produced on startup")
	hostname = flag.String("hostname", "", "Application hostname")
	address  = flag.String("address", ":8080", "Bind address")
)

// TODO:
// Add caching for repeatable request from the same host
// catch empty cred bash exception for python
// Make hostname configurable across the board
// manage credentials file in PVC!!!

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetReportCaller(true)
	log := logrus.NewEntry(logrus.StandardLogger())

	if *devMode {
		log.Debug("Dev mode - produce dummy files")
		dummyFiles()
	}
	if *hostname == "" {
		*hostname = "http://localhost:8080"
	}

	log.Info("starting the osa lab dispatcher")
	s := server.New(log, *hostname, *address)

	s.Run()
}

func dummyFiles() {
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
	err = ioutil.WriteFile("storage/credentials.yaml", bytes, 0600)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/server"
)

var (
	devMode      = flag.Bool("dev-mode", false, "If set, dummy files will be produced on startup")
	hostname     = flag.String("hostname", "", "Application hostname")
	address      = flag.String("address", ":8080", "Bind address")
	workerImage  = flag.String("worker-image", "quay.io/mangirdas/labs-worker", "Worker container image")
	workerNumber = flag.Int("worker-number", 5, "Number of workers")
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

	if *hostname == "" {
		*hostname = "http://localhost:8080"
	}

	log.Info("starting the osa lab dispatcher")
	s, err := server.New(log, *devMode, *hostname, *address, *workerImage, *workerNumber)
	if err != nil {
		panic(err)
	}

	err = s.Run()
	if err != nil {
		panic(err)
	}
}

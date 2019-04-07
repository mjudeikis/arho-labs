package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/mjudeikis/osa-labs/pkg/ssh"
)

var (
	port      = flag.String("port", "2222", "Bind address")
	publicKey = flag.String("public-key", "/data/id_rsa.pub", "Public key location")
)

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetReportCaller(true)
	log := logrus.NewEntry(logrus.StandardLogger())

	log.Info("starting the lab ssh server")
	s, err := ssh.New(log, *port, *publicKey)
	if err != nil {
		panic(err)
	}

	err = s.Run()
	if err != nil {
		panic(err)
	}
}

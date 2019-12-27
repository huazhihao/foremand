package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {

	flag.Parse()
	log.WithFields(log.Fields{
		"version": Version,
		"gitsha":  GitSHA,
		"runtime": runtime.Version(),
	}).Debug("version")
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	fmd, err := InitForemand(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	go fmd.Start(doneChan, errChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
		case sig := <-signalChan:
			log.WithFields(log.Fields{"signal": sig}).Info("Exiting...")
			close(doneChan)
			fmd.Stop()
			os.Exit(0)
		}
	}
}

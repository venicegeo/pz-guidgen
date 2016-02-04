package main

import (
	//"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-uuidgen/server"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"log"
)


func main() {

	var mode piazza.ConfigMode = piazza.ConfigModeCloud
	if piazza.IsLocalConfig() {
		mode = piazza.ConfigModeLocal
	}

	config, err := piazza.NewConfig("pz-uuidgen", mode)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerPkg.NewPzLoggerClient(sys)
	if err != nil {
		log.Fatal(err)
	}
	err = sys.WaitForService("pz-logger", 1000)
	if err != nil {
		log.Fatal(err)
	}

	err = server.RunUUIDServer(sys, logger)
	if err != nil {
		log.Fatal(err)
	}

	// not reached
	log.Fatal("not reached")
}

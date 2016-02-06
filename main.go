package main

import (
	//"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-uuidgen/server"
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

	loggerService, err := loggerPkg.NewPzLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	done := sys.StartServer(server.CreateHandlers(sys, loggerService))

	err = <- done
	if err != nil {
		log.Fatal(err)
	}
}

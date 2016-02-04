package main

import (
	//"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-uuidgen/server"
	"log"
)


func main() {

	local := piazza.IsLocalConfig()

	config, err := piazza.GetConfig("pz-uuidgen", local)
	if err != nil {
		log.Fatal(err)
	}

	discoverClient, err := piazza.NewDiscoverClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = discoverClient.RegisterServiceWithDiscover(config.ServiceName, config.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	err = discoverClient.WaitForService("pz-logger", 1000)
	if err != nil {
		log.Fatal(err)
	}

	err = server.RunUUIDServer(config)
	if err != nil {
		log.Fatal(err)
	}

	// not reached
	log.Fatal("not reached")
}

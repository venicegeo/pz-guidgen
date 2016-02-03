package main

import (
	//"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-uuidgen/server"
	"log"
	"os"
)

var pzService *piazza.PzService


func Main(done chan bool, local bool) int {

	var err error

	config, err := piazza.GetConfig("pz-uuidgen", local)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = config.RegisterServiceWithDiscover()
	if err != nil {
		log.Fatal(err)
		return 1
	}

	pzService, err = piazza.NewPzService(config, false)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = pzService.WaitForService("pz-logger", 1000)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	if done != nil {
		done <- true
	}

	err = server.RunUUIDServer(config.BindTo, pzService)
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main() {
	local := piazza.IsLocalConfig()
	os.Exit(Main(nil, local))
}

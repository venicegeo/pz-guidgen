package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var pzService *piazza.PzService

var debugCounter = 0

var numRequests = 0
var numUUIDs = 0

var startTime = time.Now()

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-uuidgen.")
}

func handleAdminGet(c *gin.Context) {
	respUuid := piazza.AdminResponseUuidgen{NumRequests: numRequests, NumUUIDs: numUUIDs}
	resp := piazza.AdminResponse{StartTime: startTime, Uuidgen: &respUuid}

	c.IndentedJSON(http.StatusOK, resp)
}

// request body is ignored
// we allow a count of zero, for testing
func handleUUIDService(c *gin.Context) {

	var count int
	var err error

	key := c.Query("count")
	if key == "" {
		count = 1
	} else {
		count, err = strconv.Atoi(key)
		if err != nil {
			c.String(http.StatusBadRequest, "query argument invalid: %s", key)
			return
		}
	}

	if count < 0 || count > 255 {
		c.String(http.StatusBadRequest, "query argument out of range: %d", count)
		return
	}

	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		if pzService.Debug {
			uuids[i] = fmt.Sprintf("%d", debugCounter)
			debugCounter++
		} else {
			uuids[i] = uuid.New()
		}
	}

	data := make(map[string]interface{})
	data["data"] = uuids

	numUUIDs += count
	numRequests++

	// @TODO ignore any failure here
	pzService.Log(piazza.SeverityInfo, fmt.Sprintf("uuidgen created %d", count))

	c.IndentedJSON(http.StatusOK, data)
}

func runUUIDServer() error {

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	router.GET("/uuid/admin", func(c *gin.Context) { handleAdminGet(c) })

	router.POST("/uuid", func(c *gin.Context) { handleUUIDService(c) })

	router.GET("/", func(c *gin.Context) { handleHealthCheck(c) })

	return router.Run(pzService.Address)
}

func app() int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	serviceAddress, discoverAddress, debug, err := piazza.NewDiscoverService("pz-uuidgen", "localhost:12340", "localhost:3000")
	if err != nil {
		log.Print(err)
		return 1
	}

	pzService, err = piazza.NewPzService("pz-logger", serviceAddress, discoverAddress, debug)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = runUUIDServer()
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app()
}

func main() {
	os.Exit(app())
}

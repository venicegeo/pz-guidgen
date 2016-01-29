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

var debugMode bool

func handleGetRoot(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-uuidgen.")
}

func handleGetAdminStats(c *gin.Context) {
	respUuid := piazza.AdminResponseUuidgen{NumRequests: numRequests, NumUUIDs: numUUIDs}
	resp := piazza.AdminResponse{StartTime: startTime, Uuidgen: &respUuid}

	c.IndentedJSON(http.StatusOK, resp)
}

// request body is ignored
// we allow a count of zero, for testing
func handlePostUuids(c *gin.Context) {

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

func handleGetAdminSettings(c *gin.Context) {
	s := "false"
	if debugMode {
		s = "true"
	}
	m := map[string]string{"debug": s}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminSettings(c *gin.Context) {
	m := map[string]string{}
	err := c.BindJSON(&m)
	if err != nil {
		c.Error(err)
		return
	}
	for k, v := range m {
		switch k {
		case "debug":
			switch v {
			case "true":
				debugMode = true
				break
			case "false":
				debugMode = false
			default:
				c.String(http.StatusBadRequest, "Illegal value for 'debug': %s", v)
				return
			}
		default:
			c.String(http.StatusBadRequest, "Unknown parameter: %s", k)
			return
		}
	}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminShutdown(c *gin.Context) {
	var reason string
	err := c.BindJSON(&reason)
	if err != nil {
		c.String(http.StatusBadRequest, "no reason supplied")
		return
	}
	pzService.Log(piazza.SeverityFatal, "Shutdown requested: "+reason)

	// TODO: need a graceful shutdown method
	os.Exit(0)
}

func runUUIDServer() error {

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) { handleGetRoot(c) })

	router.POST("/v1/uuids", func(c *gin.Context) { handlePostUuids(c) })

	router.GET("/v1/admin/stats", func(c *gin.Context) { handleGetAdminStats(c) })

	router.GET("/v1/admin/settings", func(c *gin.Context) { handleGetAdminSettings(c) })
	router.POST("/v1/admin/settings", func(c *gin.Context) { handlePostAdminSettings(c) })

	router.POST("/v1/admin/shutdown", func(c *gin.Context) { handlePostAdminShutdown(c) })

	return router.Run(pzService.Address)
}

func app(done chan bool) int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	serviceAddress, discoverAddress, debug, err := piazza.NewDiscoverService("pz-uuidgen", "localhost:12340", "localhost:3000")
	if err != nil {
		log.Print(err)
		return 1
	}

	pzService, err = piazza.NewPzService("pz-uuidgen", serviceAddress, discoverAddress, debug)
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

	err = runUUIDServer()
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string, done chan bool) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)

	return app(done)
}

func main() {
	os.Exit(app(nil))
}

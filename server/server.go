package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-uuidgen/client"
	"net/http"
	"strconv"
	"time"
	"log"
)

var debugCounter = 0

var numRequests = 0
var numUUIDs = 0

var startTime = time.Now()

var debugMode bool

func handleGetRoot(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-uuidgen.")
}

func handleGetAdminStats(c *gin.Context) {
	stats := client.UuidGenAdminStats{StartTime: startTime, NumRequests: numRequests, NumUUIDs: numUUIDs}
	c.IndentedJSON(http.StatusOK, stats)
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
		if debugMode {
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
	//pzService.Log(piazza.SeverityInfo, fmt.Sprintf("uuidgen created %d", count))
	log.Printf("INFO: uuidgen created %d", count)
	c.IndentedJSON(http.StatusOK, data)
}

func handleGetAdminSettings(c *gin.Context) {
	s := client.UuidGenAdminSettings{Debug: debugMode}
	c.JSON(http.StatusOK, s)
}

func handlePostAdminSettings(c *gin.Context) {
	settings := client.UuidGenAdminSettings{}
	err := c.BindJSON(&settings)
	if err != nil {
		c.Error(err)
		return
	}
	debugMode = settings.Debug
	c.String(http.StatusOK, "")
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func RunUUIDServer(config *piazza.ServiceConfig) error {

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

	return router.Run(config.BindTo)
}

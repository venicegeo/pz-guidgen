package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-uuidgen/client"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type LockedAdminSettings struct {
	sync.Mutex
	client.UuidGenAdminSettings
}

var settings LockedAdminSettings

type LockedAdminStats struct {
	sync.Mutex
	client.UuidGenAdminStats
}

var stats LockedAdminStats

func init() {
	stats.StartTime = time.Now()
}

func handleGetRoot(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-uuidgen.")
	log.Print("got health-check request")
}

func handleGetAdminStats(c *gin.Context) {
	stats.Lock()
	t := stats.UuidGenAdminStats
	stats.Unlock()
	c.JSON(http.StatusOK, t)
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
		if settings.Debug {
			stats.Lock()
			uuids[i] = fmt.Sprintf("%d", stats.DebugCount)
			stats.DebugCount++
			stats.Unlock()
		} else {
			uuids[i] = uuid.New()
		}
	}

	data := make(map[string]interface{})
	data["data"] = uuids

	stats.Lock()
	stats.NumUUIDs += count
	stats.NumRequests++
	stats.Unlock()

	// @TODO ignore any failure here
	//pzService.Log(piazza.SeverityInfo, fmt.Sprintf("uuidgen created %d", count))
	log.Printf("INFO: uuidgen created %d", count)
	c.IndentedJSON(http.StatusOK, data)
}

func handleGetAdminSettings(c *gin.Context) {
	settings.Lock()
	t := settings
	settings.Unlock()
	c.JSON(http.StatusOK, t)
}

func handlePostAdminSettings(c *gin.Context) {
	t := client.UuidGenAdminSettings{}
	err := c.BindJSON(&t)
	if err != nil {
		c.Error(err)
		return
	}
	settings.Lock()
	settings.UuidGenAdminSettings = t
	settings.Unlock()

	c.String(http.StatusOK, "")
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
}

func CreateHandlers(sys *piazza.System, logger loggerPkg.ILoggerService) http.Handler {

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

	return router
}

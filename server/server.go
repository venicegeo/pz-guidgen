// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/lib"
	"github.com/venicegeo/pz-uuidgen/client"
)

var logger loggerPkg.IClient

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
	//log.Print("got health-check request")
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
	var debug bool
	var prefix string
	var err error
	var key string

	// ?count=INT
	key = c.Query("count")
	if key == "" {
		count = 1
	} else {
		count, err = strconv.Atoi(key)
		if err != nil {
			c.String(http.StatusBadRequest, "query argument invalid: %s", key)
			return
		}
	}

	// ?debug=BOOL
	key = c.Query("debug")
	if key == "" {
		debug = false
	} else {
		debug, err = strconv.ParseBool(key)
		if err != nil {
			c.String(http.StatusBadRequest, "query argument invalid: %s", key)
			return
		}
	}

	// ?prefix=STR
	// valid only if debug is false
	prefix = c.Query("prefix")

	if !debug && (prefix != "") {
		c.String(http.StatusBadRequest, "\"?prefix\" query parameter only valid if \"?debug\" is true")
		return
	}

	if count < 0 || count > 255 {
		c.String(http.StatusBadRequest, "query argument out of range: %d", count)
		return
	}

	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		if debug {
			stats.Lock()
			uuids[i] = fmt.Sprintf("%s%d", prefix, stats.DebugCount)
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

	// @TODO handle failures
	if logger != nil {
		s := fmt.Sprintf("generated %d: %s", count, uuids[0])
		err = logger.Log(piazza.PzUuidgen, "0.0.0.0", loggerPkg.SeverityInfo, time.Now(), s)

		if err != nil {
			log.Printf("error writing to logger: %s", err)
		}
	}
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

func CreateHandlers(sys *piazza.SystemConfig, loggerp loggerPkg.IClient) http.Handler {

	logger = loggerp

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

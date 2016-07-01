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

package uuidgen

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/logger"
)

var logger loggerPkg.IClient

type LockedAdminStats struct {
	sync.Mutex
	UuidGenAdminStats
}

var stats LockedAdminStats

func Init(l loggerPkg.IClient) {
	stats.CreatedOn = time.Now()
	logger = l
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
	var specified bool

	// ?count=INT
	key, specified = c.GetQuery("count")

	countInQueryString := strings.Contains(
		strings.ToLower(c.Request.URL.RawQuery), "count=")

	// fmt.Printf("key='%s', specified=%v, countInQueryString=%v\n", key, specified, countInQueryString)

	if key == "" {
		if !specified && !countInQueryString {
			count = 1
		} else {
			c.String(http.StatusBadRequest, "query argument invalid: %s", key)
			return
		}
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

var Routes = []piazza.RouteData{
	{"GET", "/", handleGetRoot},
	{"GET", "/v1/admin/stats", handleGetAdminStats},
	{"POST", "/v1/uuids", handlePostUuids},
}

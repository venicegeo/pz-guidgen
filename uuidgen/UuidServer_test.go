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
	"log"
	"testing"
	"time"

	"github.com/pborman/uuid"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/logger"
)

const MOCKING = true

type UuidgenTester struct {
	suite.Suite
	sys           *piazza.SystemConfig
	total         int
	loggerClient  loggerPkg.IClient
	uuidClient    IClient
	genericServer *piazza.GenericServer
	uuidServer    *UuidServer
	uuidService   *UuidService
}

func (suite *UuidgenTester) SetupSuite() {
	var err error

	var required []piazza.ServiceName
	if MOCKING {
		required = []piazza.ServiceName{}
	} else {
		required = []piazza.ServiceName{
			piazza.PzElasticSearch,
			piazza.PzLogger,
		}
	}

	suite.sys, err = piazza.NewSystemConfig(piazza.PzUuidgen, required)
	if err != nil {
		log.Fatal(err)
	}

	if MOCKING {
		suite.loggerClient, err = loggerPkg.NewMockClient(suite.sys)
		if err != nil {
			log.Fatal(err)
		}
		suite.uuidClient, err = NewMockClient(suite.sys)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		suite.loggerClient, err = loggerPkg.NewClient(suite.sys)
		if err != nil {
			log.Fatal(err)
		}
		suite.uuidClient, err = NewClient(suite.sys)
		if err != nil {
			log.Fatal(err)
		}
	}

	suite.total = 0

	suite.uuidService = &UuidService{}
	err = suite.uuidService.Init(suite.loggerClient)
	if err != nil {
		log.Fatal(err)
	}

	suite.uuidServer = &UuidServer{}
	suite.uuidServer.Init(suite.uuidService)

	suite.genericServer = &piazza.GenericServer{Sys: suite.sys}
	err = suite.genericServer.Configure(suite.uuidServer.Routes)
	if err != nil {
		log.Fatal(err)
	}
	_, err = suite.genericServer.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *UuidgenTester) TearDownSuite() {
	suite.genericServer.Stop()
}

func TestRunSuite(t *testing.T) {
	s := new(UuidgenTester)
	suite.Run(t, s)
}

func (suite *UuidgenTester) checkValidStatsResponse(t *testing.T, stats *UuidGenAdminStats) {
	assert.WithinDuration(t, time.Now(), stats.CreatedOn, 5*time.Second)

	assert.Equal(t, suite.total, stats.NumUUIDs)
}

func (suite *UuidgenTester) checkValidResponse(t *testing.T, data *[]string, count int) []uuid.UUID {
	assert.Len(t, *data, count)

	values := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		values[i] = uuid.Parse((*data)[i])
		if values[i] == nil {
			t.Fatalf("returned uuid has invalid format: %v", values)
		}
	}

	return values
}

func (suite *UuidgenTester) Test01Okay() {
	t := suite.T()
	assert := assert.New(t)

	var err error
	var tmp []uuid.UUID

	values := []uuid.UUID{}

	var uuidClient = suite.uuidClient

	data, err := uuidClient.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, data, 1)
	values = append(values, tmp...)
	suite.total += 1

	data, err = uuidClient.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, data, 1)
	values = append(values, tmp...)
	suite.total += 1

	data, err = uuidClient.PostUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, data, 1)
	values = append(values, tmp...)
	suite.total += 1

	data, err = uuidClient.PostUuids(10)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, data, 10)
	values = append(values, tmp...)
	suite.total += 10

	data, err = uuidClient.PostUuids(255)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, data, 255)
	values = append(values, tmp...)
	suite.total += 255

	// uuids should be, umm, unique
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if uuid.Equal(values[j], values[i]) {
				t.Fatalf("returned uuids not unique")
			}
		}
	}

	stats, err := uuidClient.GetStats()
	assert.NoError(err, "GetStats")
	suite.checkValidStatsResponse(t, stats)

	s, err := uuidClient.GetUuid()
	assert.NoError(err, "pzService.GetUuid")
	assert.NotEmpty(s, "GetUuid failed - returned empty string")
	suite.total += 1
}

func (suite *UuidgenTester) Test02Bad() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	var uuidClient = suite.uuidClient

	// count out of range
	_, err = uuidClient.PostUuids(-1)
	assert.Error(err)

	// count out of range
	_, err = uuidClient.PostUuids(256)
	assert.Error(err)
}
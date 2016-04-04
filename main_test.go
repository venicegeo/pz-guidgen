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

package main

import (
	"log"
	"testing"
	"time"

	"github.com/pborman/uuid"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-uuidgen/server"
)

const MOCKING = true

type UuidgenTester struct {
	suite.Suite
	sys     *piazza.SystemConfig
	total   int
	logger  loggerPkg.ILoggerService
	uuidgen client.IUuidGenService
}

func (suite *UuidgenTester) SetupSuite() {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzUuidgen, required, true)
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	if MOCKING {
		suite.logger, err = loggerPkg.NewMockLoggerService(sys)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		suite.logger, err = loggerPkg.NewPzLoggerService(sys)
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = sys.StartServer(server.CreateHandlers(sys, suite.logger))

	suite.uuidgen, err = client.NewPzUuidGenService(sys)
	if err != nil {
		log.Fatal(err)
	}

	suite.total = 0
}

func (suite *UuidgenTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(UuidgenTester)
	suite.Run(t, s)
}

func (suite *UuidgenTester) checkValidStatsResponse(t *testing.T, stats *client.UuidGenAdminStats) {
	assert.WithinDuration(t, time.Now(), stats.StartTime, 5*time.Second, "service start time too long ago")

	assert.Equal(t, suite.total, stats.NumUUIDs)
}

func (suite *UuidgenTester) checkValidResponse(t *testing.T, resp *client.UuidGenResponse, count int) []uuid.UUID {
	assert.Len(t, resp.Data, count)

	values := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		values[i] = uuid.Parse(resp.Data[i])
		if values[i] == nil {
			t.Fatalf("returned uuid has invalid format: %v", values)
		}
	}

	return values
}

func (suite *UuidgenTester) checkValidDebugResponse(t *testing.T, resp *client.UuidGenResponse, count int) []string {

	assert.Len(t, resp.Data, count)

	return resp.Data
}

func (suite *UuidgenTester) TestOkay() {
	t := suite.T()
	assert := assert.New(t)

	var resp *client.UuidGenResponse
	var err error
	var tmp []uuid.UUID

	values := []uuid.UUID{}

	var uuidgen = suite.uuidgen

	resp, err = uuidgen.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, resp, 1)
	values = append(values, tmp...)
	suite.total += 1

	resp, err = uuidgen.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, resp, 1)
	values = append(values, tmp...)
	suite.total += 1

	resp, err = uuidgen.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, resp, 1)
	values = append(values, tmp...)
	suite.total += 1

	resp, err = uuidgen.PostToUuids(10)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, resp, 10)
	values = append(values, tmp...)
	suite.total += 10

	resp, err = uuidgen.PostToUuids(255)
	assert.NoError(err, "PostToUuids")
	tmp = suite.checkValidResponse(t, resp, 255)
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

	stats, err := uuidgen.GetFromAdminStats()
	assert.NoError(err, "PostToUuids")
	suite.checkValidStatsResponse(t, stats)

	s, err := uuidgen.GetUuid()
	assert.NoError(err, "pzService.GetUuid")
	assert.NotEmpty(s, "GetUuid failed - returned empty string")
	suite.total += 1
}

func (suite *UuidgenTester) TestDebugOkay() {
	t := suite.T()
	assert := assert.New(t)

	var resp *client.UuidGenResponse
	var err error
	var tmp []string

	values := []string{}

	var uuidgen = suite.uuidgen
	resp, err = uuidgen.PostToDebugUuids(1, "XYZZY")
	assert.NoError(err, "PostToDebugUuids")
	tmp = suite.checkValidDebugResponse(t, resp, 1)
	values = append(values, tmp...)
	suite.total += 1

	resp, err = uuidgen.PostToDebugUuids(1, "Yow.")
	assert.NoError(err, "PostToDebugUuids")
	tmp = suite.checkValidDebugResponse(t, resp, 1)
	values = append(values, tmp...)
	suite.total += 1

	resp, err = uuidgen.PostToDebugUuids(10, "A")
	assert.NoError(err, "PostToDebugUuids")
	tmp = suite.checkValidDebugResponse(t, resp, 10)
	values = append(values, tmp...)
	suite.total += 10

	// did prefix work?
	assert.EqualValues("XYZZY0", values[0][0:6])
	assert.EqualValues("Yow.1", values[1][0:5])
	for i := 2; i < len(values); i++ {
		assert.EqualValues('A', values[i][0])
	}

	// uuids should be, umm, unique
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			assert.False(values[j] == values[i], "returned uuids not unique %d %d %s %s", i, j, values[i], values[j])
		}
	}

	s, err := uuidgen.GetDebugUuid("PFX")
	assert.NoError(err, "pzService.GetDebugUuid")
	assert.NotEmpty(s, "GetDebugUuid failed - returned empty string")
	suite.total += 1
}

func (suite *UuidgenTester) TestBad() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	var uuidgen = suite.uuidgen

	// count out of range
	_, err = uuidgen.PostToUuids(-1)
	assert.Error(err)

	// count out of range
	_, err = uuidgen.PostToUuids(256)
	assert.Error(err)
}

func (suite *UuidgenTester) TestAdminSettings() {
	t := suite.T()
	assert := assert.New(t)

	var uuidgen = suite.uuidgen

	// no settings fields anymore, so this is kinda dumb

	settings, err := uuidgen.GetFromAdminSettings()
	assert.NoError(err, "GetFromAdminSettings")

	err = uuidgen.PostToAdminSettings(settings)
	assert.NoError(err, "PostToAdminSettings")
}

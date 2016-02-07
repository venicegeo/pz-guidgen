package main

import (
	"github.com/pborman/uuid"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-uuidgen/client"
	"github.com/venicegeo/pz-uuidgen/server"
	"log"
	"testing"
	"time"
)

type UuidgenTester struct {
	suite.Suite

	logger     loggerPkg.ILoggerService
	uuidgenner client.IUuidGenService
}

func (suite *UuidgenTester) SetupSuite() {
	//t := suite.T()

	config, err := piazza.NewConfig(piazza.PzUuidgen, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	suite.logger, err = loggerPkg.NewMockLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(server.CreateHandlers(sys, suite.logger))

	suite.uuidgenner, err = client.NewPzUuidGenService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *UuidgenTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(UuidgenTester)
	suite.Run(t, s)
}

func checkValidStatsResponse(t *testing.T, stats *client.UuidGenAdminStats) {
	assert.WithinDuration(t, time.Now(), stats.StartTime, 5*time.Second, "service start time too long ago")

	assert.True(t, stats.NumUUIDs == 268 || stats.NumUUIDs == 272, "num uuids: expected 268/272, actual %d", stats.NumUUIDs)
	assert.True(t, stats.NumRequests == 5 || stats.NumRequests == 7, "num requests: expected 5/7, actual %d", stats.NumRequests)
}

func checkValidResponse(t *testing.T, resp *client.UuidGenResponse, count int) []uuid.UUID {

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

func checkValidDebugResponse(t *testing.T, resp *client.UuidGenResponse, count int) []string {

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

	var uuidgenner = suite.uuidgenner

	//////////////////////
	{
		settings, err := uuidgenner.GetFromAdminSettings()
		assert.NoError(err, "GetFromAdminSettings")
		assert.False(settings.Debug, "settings.Debug")

		settings.Debug = true
		err = uuidgenner.PostToAdminSettings(settings)
		assert.NoError(err, "PostToAdminSettings")

		settings, err = uuidgenner.GetFromAdminSettings()
		assert.NoError(err, "GetFromAdminSettings")
		assert.True(settings.Debug, "settings.Debug")

		settings.Debug = false
		err = uuidgenner.PostToAdminSettings(settings)
		assert.NoError(err, "PostToAdminSettings")

		settings, err = uuidgenner.GetFromAdminSettings()
		assert.NoError(err, "GetFromAdminSettings")
		assert.False(settings.Debug, "settings.Debug")
	}
	////////////////////////

	resp, err = uuidgenner.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	resp, err = uuidgenner.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	resp, err = uuidgenner.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	resp, err = uuidgenner.PostToUuids(10)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidResponse(t, resp, 10)
	values = append(values, tmp...)

	resp, err = uuidgenner.PostToUuids(255)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidResponse(t, resp, 255)
	values = append(values, tmp...)

	// uuids should be, umm, unique
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if uuid.Equal(values[j], values[i]) {
				t.Fatalf("returned uuids not unique")
			}
		}
	}

	stats, err := uuidgenner.GetFromAdminStats()
	assert.NoError(err, "PostToUuids")
	checkValidStatsResponse(t, stats)

	s, err := uuidgenner.GetUuid()
	assert.NoError(err, "pzService.GetUuid")
	assert.NotEmpty(s, "GetUuid failed - returned empty string")
}

func (suite *UuidgenTester) TestBad() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	var uuidgenner = suite.uuidgenner

	// count out of range
	_, err = uuidgenner.PostToUuids(-1)
	assert.Error(err)

	// count out of range
	_, err = uuidgenner.PostToUuids(256)
	assert.Error(err)
}

func (suite *UuidgenTester) TestDebug() {
	t := suite.T()
	assert := assert.New(t)

	var uuidgenner = suite.uuidgenner

	var resp *client.UuidGenResponse
	var err error
	var tmp []string

	values := []string{}

	/////////////////
	settings := &client.UuidGenAdminSettings{Debug: true}
	err = uuidgenner.PostToAdminSettings(settings)
	assert.NoError(err, "PostToAdminSettings")

	resp, err = uuidgenner.PostToUuids(1)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidDebugResponse(t, resp, 1)
	values = append(values, tmp...)

	resp, err = uuidgenner.PostToUuids(3)
	assert.NoError(err, "PostToUuids")
	tmp = checkValidDebugResponse(t, resp, 3)
	values = append(values, tmp...)

	if values[0] != "0" || values[1] != "1" || values[2] != "2" || values[3] != "3" {
		t.Fatalf("invalid debug uuids returned: %v", values)
	}

	// set it back
	settings = &client.UuidGenAdminSettings{Debug: false}
	err = uuidgenner.PostToAdminSettings(settings)
	assert.NoError(err, "PostToAdminSettings")
}

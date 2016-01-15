package main

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	assert "github.com/stretchr/testify/assert"
	piazza "github.com/venicegeo/pz-gocommon"
	pztesting "github.com/venicegeo/pz-gocommon/testing"
	"io/ioutil"
	http "net/http"
	"testing"
	"time"
)

// @TODO: need to automate call to setup() and/or kill thread after each test
func setup(port string, debug bool) {
	s := fmt.Sprintf("-discovery http://localhost:3000 -port %s", port)
	if debug {
		s += " -debug"
	}

	go main2(s)

	time.Sleep(250 * time.Millisecond)
}

func checkValidAdminResponse(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "bad admin response")

	data := pztesting.HttpBody(t, resp)
	t.Log(string(data))

	var m piazza.AdminResponse

	err := json.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("unmarshall of admin response: %v", err)
	}

	if time.Since(m.StartTime).Seconds() > 5 {
		t.Fatalf("service start time too long ago: %f", time.Since(m.StartTime).Seconds())
	}

	uuidgen := m.UuidGen
	if uuidgen.NumUUIDs != 268 {
		t.Fatalf("num uuids: expected %d, actual %d", 268, uuidgen.NumUUIDs)
	}
	if uuidgen.NumRequests != 5 {
		t.Fatalf("num requests: expected %d, actual %d", 5, uuidgen.NumRequests)
	}
}

func checkValidResponse(t *testing.T, resp *http.Response, count int) []uuid.UUID {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	var data map[string][]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatalf("unmarshall failed: %s", err)
	}

	uuids, ok := data["data"]
	if !ok {
		t.Fatalf("returned data has invalid data type")
	}

	if len(uuids) != count {
		t.Fatalf("returned array wrong size")
	}

	values := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		values[i] = uuid.Parse(uuids[i])
		if values[i] == nil {
			t.Fatalf("returned uuid has invalid format: %v", values)
		}

		//t.Log(values[i])
	}

	return values
}

func checkValidDebugResponse(t *testing.T, resp *http.Response, count int) []string {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	var data map[string][]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatalf("unmarshall failed: %s", err)
	}

	uuids, ok := data["data"]
	if !ok {
		t.Fatalf("returned data has invalid data type")
	}

	if len(uuids) != count {
		t.Fatalf("returned array wrong size")
	}

	return uuids
}

func TestOkay(t *testing.T) {
	setup("12340", false)

	var resp *http.Response
	var err error
	var tmp []uuid.UUID

	values := []uuid.UUID{}

	// default url
	resp, err = http.Post("http://localhost:12340/uuid", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	// url with count=1
	resp, err = http.Post("http://localhost:12340/uuid?count=1", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	// ignore other keywords
	resp, err = http.Post("http://localhost:12340/uuid?count=1&foo=bar", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidResponse(t, resp, 1)
	values = append(values, tmp...)

	// url with count=10
	resp, err = http.Post("http://localhost:12340/uuid?count=10", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidResponse(t, resp, 10)
	values = append(values, tmp...)

	// url with count=255
	resp, err = http.Post("http://localhost:12340/uuid?count=255", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
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

	resp, err = http.Get("http://localhost:12340/uuid/admin")
	if err != nil {
		t.Fatalf("admin get failed: %s", err)
	}
	checkValidAdminResponse(t, resp)
}

func TestBad(t *testing.T) {
	setup("12350", false)

	var resp *http.Response
	var err error

	// bad url
	resp, err = http.Post("http://localhost:12350/guid", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("bad url was accepted")
	}

	// count out of range
	resp, err = http.Post("http://localhost:12350/uuid?count=-1", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad count was accepted")
	}

	// count out of range
	resp, err = http.Post("http://localhost:12350/uuid?count=256", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad count was accepted")
	}

	// bad count
	resp, err = http.Post("http://localhost:12350/uuid?count=fortyleven", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad count was accepted")
	}
}

func TestDebug(t *testing.T) {
	setup("12351", true)

	var resp *http.Response
	var err error
	var tmp []string

	values := []string{}

	resp, err = http.Post("http://localhost:12351/uuid", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidDebugResponse(t, resp, 1)
	values = append(values, tmp...)

	resp, err = http.Post("http://localhost:12351/uuid?count=3", "text/plain", nil)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	tmp = checkValidDebugResponse(t, resp, 3)
	values = append(values, tmp...)

	if values[0] != "0" || values[1] != "1" || values[2] != "2" || values[3] != "3" {
		t.Fatalf("invalid debug uuids returned: %v", values)
	}
}

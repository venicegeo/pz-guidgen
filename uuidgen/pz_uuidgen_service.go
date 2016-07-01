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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

type PzUuidGenService struct {
	url     string
	name    piazza.ServiceName
	address string
}

func NewPzUuidGenService(sys *piazza.SystemConfig) (*PzUuidGenService, error) {
	var _ IUuidGenService = new(PzUuidGenService)

	var err error

	err = sys.WaitForService(piazza.PzUuidgen)
	if err != nil {
		return nil, err
	}

	url, err := sys.GetURL(piazza.PzUuidgen)
	if err != nil {
		return nil, err
	}

	service := &PzUuidGenService{url: url}

	return service, nil
}

func (c *PzUuidGenService) PostToUuids(count int) (*UuidGenResponse, error) {

	url := fmt.Sprintf("%s/uuids?count=%d", c.url, count)

	resp, err := http.Post(url, piazza.ContentTypeJSON, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m UuidGenResponse
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (c *PzUuidGenService) PostToDebugUuids(count int, prefix string) (*UuidGenResponse, error) {

	url := fmt.Sprintf("%s/uuids?count=%d&debug=true&prefix=%s", c.url, count, prefix)

	resp, err := http.Post(url, piazza.ContentTypeJSON, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m UuidGenResponse
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (c *PzUuidGenService) GetFromAdminStats() (*UuidGenAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(UuidGenAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (pz *PzUuidGenService) GetUuid() (string, error) {

	resp, err := pz.PostToUuids(1)
	if err != nil {
		return "", err
	}

	return resp.Data[0], nil
}

func (pz *PzUuidGenService) GetDebugUuid(prefix string) (string, error) {

	resp, err := pz.PostToDebugUuids(1, prefix)
	if err != nil {
		return "", err
	}

	return resp.Data[0], nil
}

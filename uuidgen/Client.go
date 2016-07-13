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
	"errors"
	"fmt"
	"log"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

type Client struct {
	url string
}

//---------------------------------------------------------------------

func NewClient(sys *piazza.SystemConfig) (*Client, error) {
	var _ IClient = new(Client)

	var err error

	err = sys.WaitForService(piazza.PzUuidgen)
	if err != nil {
		return nil, err
	}

	url, err := sys.GetURL(piazza.PzUuidgen)
	if err != nil {
		return nil, err
	}

	service := &Client{url: url}
	return service, nil
}

//---------------------------------------------------------------------

func (c *Client) PostUuids(count int) (*[]string, error) {

	endpoint := fmt.Sprintf("/uuids?count=%d", count)

	resp := piazza.HttpPostJson(c.url+endpoint, nil)
	if resp.IsError() {
		return nil, resp.ToError()
	}

	// the only thing we return from a POST is a string-list
	if resp.Type != "string-list" {
		err := errors.New(fmt.Sprintf("Unsupported response data type: %s", resp.Type))
		return nil, err
	}

	out := make([]string, count)
	err := resp.ExtractData(out)
	return &out, err
}

func (c *Client) GetStats() (*UuidGenAdminStats, error) {
	resp := piazza.HttpGetJson(c.url + "/admin/stats")
	if resp.IsError() {
		return nil, resp.ToError()
	}
	out := &UuidGenAdminStats{}
	err := resp.ExtractData(out)
	return out, err
}

func (c *Client) GetUuid() (string, error) {

	data, err := c.PostUuids(1)
	if err != nil {
		return "", err
	}

	log.Printf("GetUuid returning %s / %#v", (*data)[0], data)
	return (*data)[0], nil
}

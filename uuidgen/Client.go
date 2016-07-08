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

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

type Client struct {
	url string
	//	name    piazza.ServiceName
	//	address string
}

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
	log.Printf("CLIENT URL: %s", url)
	fmt.Printf("CLIENT URL2: %s", url)
	return service, nil
}

//----------------------------------------
// TODO: move these to gocommon

func asObject(resp *piazza.JsonResponse, out interface{}) error {
	err := piazza.SuperConverter(resp.Data, out)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getObject(endpoint string, out interface{}) error {
	resp := piazza.HttpGetJson(c.url + endpoint)
	if resp.IsError() {
		return resp.ToError()
	}
	if resp.StatusCode != http.StatusOK {
		return resp.ToError()
	}

	return asObject(resp, out)
}

func (c *Client) postObject(obj interface{}, endpoint string, out interface{}) error {
	url := c.url + endpoint
	log.Printf("URL2: %s", url)
	resp := piazza.HttpPostJson(url, obj)
	if resp.IsError() {
		return resp.ToError()
	}
	if resp.StatusCode != http.StatusCreated {
		return resp.ToError()
	}

	return asObject(resp, out)
}

//---------------------------------------------------

func (c *Client) PostUuids(count int) (*[]string, error) {

	endpoint := fmt.Sprintf("/uuids?count=%d", count)
	log.Printf("URL1: %s", endpoint)
	out := make([]string, count)
	err := c.postObject(nil, endpoint, &out)
	return &out, err
}

func (c *Client) GetStats() (*UuidGenAdminStats, error) {
	out := &UuidGenAdminStats{}
	err := c.getObject("/admin/stats", out)
	return out, err
}

func (c *Client) GetUuid() (string, error) {

	log.Printf("Client:GetUuid")

	data, err := c.PostUuids(1)
	if err != nil {
		return "", err
	}

	return (*data)[0], nil
}

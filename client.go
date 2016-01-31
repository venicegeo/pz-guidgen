package main

import (
	"errors"
	piazza "github.com/venicegeo/pz-gocommon"
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
)

type PzUuidGenClient struct{
	url string
}

func NewPzUuidGenClient(address string) *PzUuidGenClient {
	c := new(PzUuidGenClient)
	c.url = fmt.Sprintf("http://%s/v1", address)

	return c
}

func (c *PzUuidGenClient) PostToUuids(count int) (*piazza.UuidGenResponse, error) {

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

	var m piazza.UuidGenResponse
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (c *PzUuidGenClient) GetFromAdminStats() (*piazza.UuidGenAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(piazza.UuidGenAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (c *PzUuidGenClient) GetFromAdminSettings() (*piazza.UuidGenAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(piazza.UuidGenAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzUuidGenClient) PostToAdminSettings(settings *piazza.UuidGenAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url + "/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

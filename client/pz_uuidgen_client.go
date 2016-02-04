package client

import (
	"errors"
	piazza "github.com/venicegeo/pz-gocommon"
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"log"
	"io/ioutil"
)

type PzUuidGenClient struct{
	Url string
	Name string
	Address string
}

func NewPzUuidGenClient(address string) *PzUuidGenClient {
	c := new(PzUuidGenClient)
	c.Url = fmt.Sprintf("http://%s/v1", address)
	c.Address = address
	c.Name = "pz-uuidgen"
	return c
}

func (c *PzUuidGenClient) PostToUuids(count int) (*UuidGenResponse, error) {

	url := fmt.Sprintf("%s/uuids?count=%d", c.Url, count)
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

func (c *PzUuidGenClient) GetFromAdminStats() (*UuidGenAdminStats, error) {

	resp, err := http.Get(c.Url + "/admin/stats")
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

func (c *PzUuidGenClient) GetFromAdminSettings() (*UuidGenAdminSettings, error) {

	resp, err := http.Get(c.Url + "/admin/settings")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(UuidGenAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzUuidGenClient) PostToAdminSettings(settings *UuidGenAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Url + "/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (pz *PzUuidGenClient) GetUuid() (string, error) {

	resp, err := http.Post(pz.Url + "/uuids", piazza.ContentTypeText, nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	var data map[string][]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		//pz.Error("PzService.GetUuid", err)
		log.Printf("PzService.GetUuid: %v", err)
	}

	uuids, ok := data["data"]
	if !ok {
		//pz.Error("PzService.GetUuid: returned data has invalid data type", nil)
		log.Printf("PzService.GetUuid: returned data has invalid data type")
	}

	if len(uuids) != 1 {
		//pz.Error("PzService.GetUuid: returned array wrong size", nil)
		log.Printf("PzService.GetUuid: returned array wrong size")
	}

	return uuids[0], nil
}

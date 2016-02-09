package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	piazza "github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
)

type PzUuidGenService struct {
	url     string
	name    piazza.ServiceName
	address string
}

func NewPzUuidGenService(sys *piazza.System, address string) (*PzUuidGenService, error) {
	var _ piazza.IService = new(PzUuidGenService)
	var _ IUuidGenService = new(PzUuidGenService)

	var err error

	service := &PzUuidGenService{
		url:     fmt.Sprintf("http://%s/v1", address),
		name:    piazza.PzUuidgen,
		address: address}

	err = sys.WaitForService(service)
	if err != nil {
		return nil, err
	}

	sys.Services[piazza.PzUuidgen] = service

	return service, nil
}

func (c PzUuidGenService) GetName() piazza.ServiceName {
	return c.name
}

func (c PzUuidGenService) GetAddress() string {
	return c.address
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

func (c *PzUuidGenService) GetFromAdminSettings() (*UuidGenAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
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

func (c *PzUuidGenService) PostToAdminSettings(settings *UuidGenAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url+"/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (pz *PzUuidGenService) GetUuid() (string, error) {

	resp, err := pz.PostToUuids(1)
	if err != nil {
		return "", err
	}

	return resp.Data[0], nil
}

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

type PzUuidGenService struct{
	Url string
	Name string
	Address string
}

func NewPzUuidGenService(sys *piazza.System) (*PzUuidGenService, error) {
	var _ piazza.IService = new(PzUuidGenService)
	var _ IUuidGenService = new(PzUuidGenService)

	service := new(PzUuidGenService)

	service.Url = fmt.Sprintf("http://%s/v1", sys.Config.ServerAddress)

	service.Name = "pz-uuidgen"

	data, err := sys.DiscoverService.GetData(service.Name)
	if err != nil {
		return nil, err
	}
	service.Address = data.Host

	err = sys.WaitForService(service.Name, service.Address)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (c *PzUuidGenService) GetName() string {
	return c.Name
}

func (c *PzUuidGenService) GetAddress() string {
	return c.Address
}

func (c *PzUuidGenService) PostToUuids(count int) (*UuidGenResponse, error) {

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

func (c *PzUuidGenService) GetFromAdminStats() (*UuidGenAdminStats, error) {

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

func (c *PzUuidGenService) GetFromAdminSettings() (*UuidGenAdminSettings, error) {

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

func (c *PzUuidGenService) PostToAdminSettings(settings *UuidGenAdminSettings) error {

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

func (pz *PzUuidGenService) GetUuid() (string, error) {

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

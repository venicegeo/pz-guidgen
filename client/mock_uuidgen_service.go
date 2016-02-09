package client

import (
	"github.com/venicegeo/pz-gocommon"
	"fmt"
)

type MockUuidGenService struct {
	name      string
	address   string
	currentId int
}

func NewMockUuidGenService(sys *piazza.System) (*MockUuidGenService, error) {
	var _ piazza.IService = new(MockUuidGenService)
	var _ IUuidGenService = new(MockUuidGenService)

	service := &MockUuidGenService{name: piazza.PzUuidgen, address: "0.0.0.0", currentId: 0}

	sys.Services[piazza.PzUuidgen] = service

	return service, nil
}

func (service MockUuidGenService) GetName() string {
	return service.name
}

func (service MockUuidGenService) GetAddress() string {
	return service.address
}

func (service *MockUuidGenService) PostToUuids(count int) (*UuidGenResponse, error) {

	data := make([]string, count)
	for i := 0; i< count; i++ {
		data[i] = fmt.Sprintf("%d", service.currentId)
		service.currentId++
	}
	m := &UuidGenResponse{Data: []string{"xxx"}}
	return m, nil
}

func (*MockUuidGenService) GetFromAdminStats() (*UuidGenAdminStats, error) {
	return &UuidGenAdminStats{}, nil
}

func (*MockUuidGenService) GetFromAdminSettings() (*UuidGenAdminSettings, error) {
	return &UuidGenAdminSettings{}, nil
}

func (*MockUuidGenService) PostToAdminSettings(*UuidGenAdminSettings) error {
	return nil
}
func (service *MockUuidGenService) GetUuid() (string, error) {
	resp, err := service.PostToUuids(1)
	if err != nil {
		return "", err
	}
	return resp.Data[0], nil
}

package client

import (
	"github.com/venicegeo/pz-gocommon"
)

type MockUuidGenService struct{
	name    string
	address string
}

func NewMockUuidGenService(sys *piazza.System) (*MockUuidGenService, error) {
	var _ piazza.IService = new(MockUuidGenService)
	var _ IUuidGenService = new(MockUuidGenService)

	service := &MockUuidGenService{name: piazza.PzUuidGen, address: "0.0.0.0"}

	sys.Services[piazza.PzUuidGen] = service

	return service, nil
}

func (c MockUuidGenService) GetName() string {
	return c.name
}

func (c MockUuidGenService) GetAddress() string {
	return c.address
}

func (*MockUuidGenService) PostToUuids(count int) (*UuidGenResponse, error) {
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
func (*MockUuidGenService) GetUuid() (string, error) {
	return "this-is-a-uuid", nil
}

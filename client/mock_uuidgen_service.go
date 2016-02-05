package client

import (
	"github.com/venicegeo/pz-gocommon"
)

type MockUuidGenService struct{
	Name string
	Address string
}

func NewMockUuidGenService(sys *piazza.System) (*MockUuidGenService, error) {
	var _ piazza.IService = new(MockUuidGenService)
	var _ IUuidGenService = new(MockUuidGenService)

	c := MockUuidGenService{Name: "pz-uuidgen", Address: "0.0.0.0"}

	return &c, nil
}

func (c *MockUuidGenService) GetName() string {
	return c.Name
}

func (c *MockUuidGenService) GetAddress() string {
	return c.Address
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

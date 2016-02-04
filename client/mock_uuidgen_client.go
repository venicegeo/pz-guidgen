package client

import (
//	"time"
	"github.com/venicegeo/pz-gocommon"
)

type MockUuidGenClient struct{}

func NewMockUuidGenClient(sys *piazza.System) (*MockUuidGenClient, error) {
	c := MockUuidGenClient{}
	return &c, nil
}

func (*MockUuidGenClient) PostToUuids(count int) (*UuidGenResponse, error) {
	m := &UuidGenResponse{Data: []string{"xxx"}}
	return m, nil
}

func (*MockUuidGenClient) GetFromAdminStats() (*UuidGenAdminStats, error) {
	return &UuidGenAdminStats{}, nil
}

func (*MockUuidGenClient) GetFromAdminSettings() (*UuidGenAdminSettings, error) {
	return &UuidGenAdminSettings{}, nil
}

func (*MockUuidGenClient) PostToAdminSettings(*UuidGenAdminSettings) error {
	return nil
}
func (*MockUuidGenClient) GetUuid() (string, error) {
	return "this-is-a-uuid", nil
}

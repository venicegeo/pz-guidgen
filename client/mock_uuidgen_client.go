package client

import (
//	"time"
)

type MockUuidGen struct{}

func (*MockUuidGen) PostToUuids(count int) (*UuidGenResponse, error) {
	m := &UuidGenResponse{Data: []string{"xxx"}}
	return m, nil
}

func (*MockUuidGen) GetFromAdminStats() (*UuidGenAdminStats, error) {
	return &UuidGenAdminStats{}, nil
}

func (*MockUuidGen) GetFromAdminSettings() (*UuidGenAdminSettings, error) {
	return &UuidGenAdminSettings{}, nil
}

func (*MockUuidGen) PostToAdminSettings(*UuidGenAdminSettings) error {
	return nil
}

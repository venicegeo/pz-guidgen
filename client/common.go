package client

import (
	"time"
)

type IUuidGenService interface {
	GetName() string
	GetAddress() string

	// high-level interfaces
	GetUuid() (string, error)

	// low-level interfaces
	PostToUuids(count int) (*UuidGenResponse, error)
	GetFromAdminStats() (*UuidGenAdminStats, error)
	GetFromAdminSettings() (*UuidGenAdminSettings, error)
	PostToAdminSettings(*UuidGenAdminSettings) error
}

type UuidGenResponse struct {
	Data []string
}

type UuidGenAdminStats struct {
	StartTime   time.Time `json:"starttime"`
	NumRequests int       `json:"num_requests"`
	NumUUIDs    int       `json:"num_uuids"`
}

type UuidGenAdminSettings struct {
	Debug bool `json:"debug"`
}

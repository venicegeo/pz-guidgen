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
	NumUUIDs    int       `json:"num_uuids"`
	DebugCount  int       `json:"debug_count"`
	NumRequests int       `json:"num_requests"`
	StartTime   time.Time `json:"starttime"`
}

type UuidGenAdminSettings struct {
	Debug bool `json:"debug"`
}

// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import "time"

type IUuidGenService interface {
	// high-level interfaces
	GetUuid() (string, error)
	GetDebugUuid(string) (string, error)

	// low-level interfaces
	PostToUuids(count int) (*UuidGenResponse, error)
	PostToDebugUuids(count int, prefix string) (*UuidGenResponse, error)
	GetFromAdminStats() (*UuidGenAdminStats, error)
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

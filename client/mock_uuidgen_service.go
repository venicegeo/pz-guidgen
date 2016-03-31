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

import (
	"fmt"

	"github.com/venicegeo/pz-gocommon"
)

type MockUuidGenService struct {
	name      piazza.ServiceName
	address   string
	currentId int
}

func NewMockUuidGenService(sys *piazza.SystemConfig) (*MockUuidGenService, error) {
	var _ IUuidGenService = new(MockUuidGenService)

	service := &MockUuidGenService{name: piazza.PzUuidgen, address: "0.0.0.0", currentId: 0}

	return service, nil
}

func (service *MockUuidGenService) PostToUuids(count int) (*UuidGenResponse, error) {

	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = fmt.Sprintf("%d", service.currentId)
		service.currentId++
	}
	m := &UuidGenResponse{Data: data}
	return m, nil
}

func (service *MockUuidGenService) PostToDebugUuids(count int, prefix string) (*UuidGenResponse, error) {

	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = fmt.Sprintf("%s%d", prefix, service.currentId)
		service.currentId++
	}
	m := &UuidGenResponse{Data: data}
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

func (service *MockUuidGenService) GetDebugUuid(prefix string) (string, error) {
	resp, err := service.PostToDebugUuids(1, prefix)
	if err != nil {
		return "", err
	}
	return resp.Data[0], nil
}

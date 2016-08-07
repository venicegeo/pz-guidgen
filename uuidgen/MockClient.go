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

package uuidgen

import (
	"errors"
	"time"

	"github.com/pborman/uuid"
)

type MockClient struct {
	stats UuidGenAdminStats
}

func NewMockClient() (*MockClient, error) {
	var _ IClient = new(MockClient)

	client := &MockClient{}

	client.stats.CreatedOn = time.Now()

	return client, nil
}

func (client *MockClient) PostUuids(count int) (*[]string, error) {

	if count < 0 || count > 255 {
		return nil, errors.New("invalid count value")
	}

	data := make([]string, count)
	for i := 0; i < count; i++ {
		data[i] = uuid.New()
	}

	client.stats.NumUUIDs += count
	client.stats.NumRequests++

	return &data, nil
}

func (client *MockClient) GetStats() (*UuidGenAdminStats, error) {
	return &client.stats, nil
}

func (client *MockClient) GetUuid() (string, error) {
	data, err := client.PostUuids(1)
	if err != nil {
		return "", err
	}
	return (*data)[0], nil
}

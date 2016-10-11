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

package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

const percolateTypeName = ".percolate"

type MockIndexType struct {
	// maps from id string to document body
	items map[string]*json.RawMessage

	mapping interface{}
}

type MockIndex struct {
	name     string
	types    map[string]*MockIndexType
	exists   bool
	open     bool
	settings interface{}
}

func NewMockIndex(indexName string) *MockIndex {
	var _ IIndex = new(MockIndex)

	esi := MockIndex{
		name:   indexName,
		types:  make(map[string]*MockIndexType),
		exists: false,
		open:   false,
	}
	return &esi
}

func (esi *MockIndex) GetVersion() string {
	return "2.2.0"
}

func (esi *MockIndex) IndexName() string {
	return esi.name
}

func (esi *MockIndex) IndexExists() (bool, error) {
	return esi.exists, nil
}

func (esi *MockIndex) TypeExists(typ string) (bool, error) {

	ok, err := esi.IndexExists()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	_, ok = esi.types[typ]
	return ok, nil
}

func (esi *MockIndex) ItemExists(typeName string, id string) (bool, error) {
	ok, err := esi.TypeExists(typeName)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	typ := esi.types[typeName]
	_, ok = (*typ).items[id]
	return ok, nil
}

// if index already exists, does nothing
func (esi *MockIndex) Create(settings string) error {
	if esi.exists {
		return fmt.Errorf("Index already exists")
	}

	esi.exists = true

	if settings == "" {
		esi.settings = nil
		return nil
	}

	obj := map[string]interface{}{}
	err := json.Unmarshal([]byte(settings), &obj)
	if err != nil {
		return err
	}

	esi.settings = obj

	for k, v := range obj["mappings"].(map[string]interface{}) {
		mapping, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = esi.addType(k, string(mapping))
		if err != nil {
			return err
		}
	}

	return nil
}

// if index doesn't already exist, does nothing
func (esi *MockIndex) Close() error {
	esi.open = false
	return nil
}

// if index doesn't already exist, does nothing
func (esi *MockIndex) Delete() error {
	esi.exists = false
	esi.open = false

	for tk, tv := range esi.types {
		for ik := range tv.items {
			delete(tv.items, ik)
		}
		delete(esi.types, tk)
	}

	return nil
}

func (esi *MockIndex) addType(typeName string, mapping string) error {

	if mapping == "" {
		return fmt.Errorf("addType: mapping may not be null")
	}

	obj := map[string]interface{}{}
	err := json.Unmarshal([]byte(mapping), &obj)
	if err != nil {
		return err
	}

	esi.types[typeName] = &MockIndexType{
		mapping: obj,
		items:   make(map[string]*json.RawMessage),
	}

	return nil
}

func (esi *MockIndex) SetMapping(typeName string, mapping piazza.JsonString) error {
	return esi.addType(typeName, string(mapping))
}

func (esi *MockIndex) PostData(typeName string, id string, obj interface{}) (*IndexResponse, error) {
	ok, err := esi.IndexExists()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Index does not exist")
	}
	ok, err = esi.TypeExists(typeName)
	if err != nil {
		return nil, err
	}

	var typ *MockIndexType
	if !ok {
		typ = &MockIndexType{}
		typ.items = make(map[string]*json.RawMessage)
		esi.types[typeName] = typ
	} else {
		typ = esi.types[typeName]
	}

	byts, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var raw json.RawMessage
	err = raw.UnmarshalJSON(byts)
	if err != nil {
		return nil, err
	}
	typ.items[id] = &raw

	r := &IndexResponse{Created: true, ID: id, Index: esi.name, Type: typeName}
	return r, nil
}

//TODO
func (esi *MockIndex) PutData(typeName string, id string, obj interface{}) (*IndexResponse, error) {
	return esi.PostData(typeName, id, obj)
}

func (esi *MockIndex) GetByID(typeName string, id string) (*GetResult, error) {
	ok, err := esi.TypeExists(typeName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("GetById: type does not exist: %s", typeName)
	}
	ok, err = esi.ItemExists(typeName, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("GetById: id does not exist: %s", id)
	}

	typ := esi.types[typeName]
	item := typ.items[id]
	r := &GetResult{ID: id, Source: item, Found: true}
	return r, nil
}

func (esi *MockIndex) DeleteByID(typeName string, id string) (*DeleteResponse, error) {
	ok, err := esi.TypeExists(typeName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("GetById: type does not exist: %s", typeName)
	}
	ok, err = esi.ItemExists(typeName, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &DeleteResponse{Found: false}, err
	}

	typ := esi.types[typeName]
	delete(typ.items, id)
	r := &DeleteResponse{Found: true}
	return r, nil

}

type srhByID []*SearchResultHit

func (a srhByID) Len() int {
	return len(a)
}
func (a srhByID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a srhByID) Less(i, j int) bool {
	return (*a[i]).ID < (*a[j]).ID
}
func srhSortMatches(matches []*SearchResultHit) []*SearchResultHit {
	sort.Sort(srhByID(matches))
	return matches
}

func (esi *MockIndex) FilterByMatchAll(typeName string, realFormat *piazza.JsonPagination) (*SearchResult, error) {
	format := NewQueryFormat(realFormat)

	objs := make(map[string]*json.RawMessage)

	if typeName == "" {
		for tk, tv := range esi.types {
			if tk == percolateTypeName {
				continue
			}
			for ik, iv := range tv.items {
				objs[ik] = iv
			}
		}
	} else {
		for ik, iv := range esi.types[typeName].items {
			objs[ik] = iv
		}
	}

	resp := &SearchResult{
		totalHits: int64(len(objs)),
		hits:      make([]*SearchResultHit, len(objs)),
	}

	i := 0
	for id, obj := range objs {
		tmp := &SearchResultHit{
			ID:     id,
			Source: obj,
		}
		resp.hits[i] = tmp
		i++
	}

	// TODO; sort key not supported
	// TODO: sort order not supported

	from := format.From
	size := format.Size

	resp.hits = srhSortMatches(resp.hits)

	if from >= len(resp.hits) {
		resp.hits = make([]*SearchResultHit, 0)
	}
	if from+size >= len(resp.hits) {
		size = len(resp.hits)
	}
	resp.hits = resp.hits[from : from+size]

	return resp, nil
}

func (esi *MockIndex) GetAllElements(typ string) (*SearchResult, error) {
	return nil, errors.New("GetAllElements not supported under mocking")
}

func (esi *MockIndex) FilterByMatchQuery(typ string, name string, value interface{}) (*SearchResult, error) {

	return nil, errors.New("FilterByMatchQuery not supported under mocking")
}

func (esi *MockIndex) FilterByTermQuery(typeName string, name string, value interface{}) (*SearchResult, error) {

	objs := make(map[string]*json.RawMessage)

	for ik, iv := range esi.types[typeName].items {
		objs[ik] = iv
	}

	resp := &SearchResult{
		totalHits: int64(len(objs)),
		hits:      make([]*SearchResultHit, 0),
	}

	i := 0
	for id, obj := range objs {
		var iface interface{}
		err := json.Unmarshal(*obj, &iface)
		if err != nil {
			return nil, err
		}
		actualValue := iface.(map[string]interface{})[name].(string)
		if actualValue != value.(string) {
			continue
		}
		tmp := &SearchResultHit{
			ID:     id,
			Source: obj,
		}
		resp.hits = append(resp.hits, tmp)
		i++
	}

	if len(resp.hits) > 0 {
		resp.Found = true
	}

	resp.hits = srhSortMatches(resp.hits)

	return resp, nil
}

func (esi *MockIndex) SearchByJSON(typ string, jsn string) (*SearchResult, error) {

	/*var obj interface{}
	err := json.Unmarshal([]byte(jsn), &obj)
	if err != nil {
		return nil, err
	}

	searchResult, err := esi.lib.Search().
		Index(esi.index).
		Type(typ).
		Source(obj).Do()

	return searchResult, err*/

	////resp := &SearchResult{}
	////return resp, nil

	return nil, errors.New("SearchByJSON not supported under mocking")
}

func (esi *MockIndex) GetTypes() ([]string, error) {
	var s []string

	for k := range esi.types {
		s = append(s, k)
	}

	return s, nil
}

func (esi *MockIndex) GetMapping(typ string) (interface{}, error) {
	return nil, errors.New("GetMapping not supported under mocking")
}

func (esi *MockIndex) AddPercolationQuery(id string, query piazza.JsonString) (*IndexResponse, error) {
	return esi.PostData(percolateTypeName, id, query)
}

func (esi *MockIndex) DeletePercolationQuery(id string) (*DeleteResponse, error) {
	return esi.DeleteByID(percolateTypeName, id)
}

var percid int

func (esi *MockIndex) AddPercolationDocument(typeName string, doc interface{}) (*PercolateResponse, error) {

	_, err := esi.PostData(percolateTypeName, strconv.Itoa(percid), doc)
	if err != nil {
		return nil, err
	}

	resp := &PercolateResponse{}
	return resp, nil
}

func (esi *MockIndex) DirectAccess(verb string, endpoint string, input interface{}, output interface{}) error {
	return fmt.Errorf("DirectAccess not supported")
}

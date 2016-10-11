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
	"fmt"
	"log"
	"sort"
	"testing"
	"time"

	"gopkg.in/olivere/elastic.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type EsTester struct {
	suite.Suite
	sys *piazza.SystemConfig
}

func (suite *EsTester) SetupSuite() {
	//t := suite.T()
}

func (suite *EsTester) TearDownSuite() {
}

func TestRunSuite(t *testing.T) {
	s1 := new(EsTester)
	suite.Run(t, s1)
}

type Obj struct {
	ID   string `json:"id" binding:"required"`
	Data string `json:"data" binding:"required"`
	Tags string `json:"tags" binding:"required"`
}

const objMapping = `{
	"Obj":{
		"properties":{
			"id": {
				"type":"string"
			},
			"data": {
				"type":"string"
			},
			"tags": {
				"type":"string"
			}
		}
	}
}`

var objs = []Obj{
	{ID: "id0", Data: "data0", Tags: "foo bar"},
	{ID: "id1", Data: "data1", Tags: "bar baz"},
	{ID: "id2", Data: "data2", Tags: "foo"},
}

const mapping = "Obj"

func (suite *EsTester) SetUpIndex() IIndex {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}

	sys, err := piazza.NewSystemConfig(piazza.PzGoCommon, required)
	if err != nil {
		log.Fatal(err)
	}

	suite.sys = sys

	esi, err := NewIndexInterface(sys, "estest$", "", true)
	assert.NoError(err)

	err = esi.Delete()
	//assert.NoError(err)

	ok, err := esi.IndexExists()
	assert.NoError(err)
	assert.False(ok)

	// make the index
	err = esi.Create("")
	assert.NoError(err)
	ok, err = esi.IndexExists()
	assert.NoError(err)
	assert.True(ok)
	err = esi.Create("")
	assert.Error(err)

	if mapping != "" {
		err = esi.SetMapping(mapping, objMapping)
		assert.NoError(err)
	}

	// populate the index
	for _, o := range objs {
		indexResult, err2 := esi.PostData(mapping, o.ID, o)
		assert.NoError(err2)
		assert.NotNil(indexResult)
	}

	// allow the database time to settle
	realFormat := &piazza.JsonPagination{
		PerPage: 10,
		Page:    0,
		Order:   piazza.SortOrderAscending,
		SortBy:  "",
	}
	pollingFn := GetData(func() (bool, error) {
		getResult, err2 := esi.FilterByMatchAll(mapping, realFormat)
		if err2 != nil {
			return false, err2
		}
		if getResult != nil && len(*getResult.GetHits()) == len(objs) {
			return true, nil
		}
		return false, nil
	})

	_, err = PollFunction(pollingFn)
	assert.NoError(err)

	return esi
}

func searchPoller(f func() (*SearchResult, error), expectedCount int) GetData {
	pollingFn := GetData(func() (bool, error) {
		getResult, err := f()
		if err != nil {
			return false, err
		}
		if getResult != nil && len(*getResult.GetHits()) == expectedCount {
			return true, nil
		}
		return false, nil
	})

	return pollingFn
}

func closerT(t *testing.T, esi IIndex) {
	assert := assert.New(t)
	err := esi.Close()
	assert.NoError(err)
	err = esi.Delete()
	assert.NoError(err)
}

//---------------------------------------------------------------------------

func (suite *EsTester) Test01Client() {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}

	sys, err := piazza.NewSystemConfig(piazza.PzGoCommon, required)
	assert.NoError(err)

	esi, err := NewIndexInterface(sys, "estest01$", "", true)
	assert.NoError(err)
	assert.EqualValues("estest01$", esi.IndexName())

	version := esi.GetVersion()
	assert.NoError(err)
	assert.Contains("2.2.0", version)

	{
		settings := `
		{
			"mappings": {
				"Frobnitz": {
					"properties": {
						"alertId": {
							"type": "string",
							"index": "not_analyzed"
						}
					}
				}
			}
		}`

		err = esi.Create(settings)
		assert.NoError(err)
	}
}

func (suite *EsTester) Test02SimplePost() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	esi := suite.SetUpIndex()
	assert.NotNil(esi)
	defer closerT(t, esi)

	err = esi.SetMapping(mapping, piazza.JsonString(objMapping))
	assert.NoError(err)

	type NotObj struct {
		ID   int    `json:"id" binding:"required"`
		Data string `json:"data" binding:"required"`
		Foo  bool   `json:"foo" binding:"required"`
	}
	o := NotObj{ID: 99, Data: "quick fox", Foo: true}

	indexResult, err := esi.PostData(mapping, "99", o)
	assert.NoError(err)
	assert.NotNil(indexResult)

	{
		// GET it
		getResult, err := esi.GetByID(mapping, "99")
		assert.NoError(err)
		assert.NotNil(getResult)
		src := getResult.Source
		assert.NotNil(src)
		var tmp1 NotObj
		err = json.Unmarshal(*src, &tmp1)
		assert.NoError(err)
		assert.EqualValues("quick fox", tmp1.Data)
		//DELETE it
		deleteResult, err := esi.DeleteByID(mapping, "99")
		assert.NoError(err)
		assert.NotNil(deleteResult)
		//PUT it
		indexResult, err := esi.PutData(mapping, "99", 0)
		assert.NoError(err)
		assert.NotNil(indexResult)
	}
}

func (suite *EsTester) Test03Operations() {
	t := suite.T()
	assert := assert.New(t)

	var tmp1 Obj
	var src *json.RawMessage

	esi := suite.SetUpIndex()
	assert.NotNil(esi)
	defer closerT(t, esi)

	{
		// GET a specific one
		getResult, err := esi.GetByID(mapping, "id1")
		assert.NoError(err)
		assert.NotNil(getResult)
		src = getResult.Source
		assert.NotNil(src)
		err = json.Unmarshal(*src, &tmp1)
		assert.NoError(err)
		assert.EqualValues("data1", tmp1.Data)
	}
	//Coverage
	_, err := esi.GetAllElements("")
	assert.Error(err)
	_, err = esi.FilterByMatchQuery("", "", "")
	assert.Error(err)
	_, err = esi.SearchByJSON("", "")
	assert.Error(err)
	_, err = esi.GetMapping("")
	assert.Error(err)
	_, err = esi.AddPercolationQuery("123", "{}")
	assert.NoError(err)
	_, err = esi.DeletePercolationQuery("123")
	assert.NoError(err)
	_, err = esi.AddPercolationDocument("123", "{}")
	assert.NoError(err)
}

func (suite *EsTester) Test07ConstructMapping() {
	t := suite.T()
	assert := assert.New(t)

	es := suite.SetUpIndex()
	assert.NotNil(es)
	defer closerT(t, es)

	items := make(map[string]MappingElementTypeName)

	items["integer1"] = MappingElementTypeInteger
	items["integer2"] = MappingElementTypeInteger
	items["double1"] = MappingElementTypeDouble
	items["bool1"] = MappingElementTypeBool
	items["date1"] = MappingElementTypeDate

	jsonstr, err := ConstructMappingSchema("MyTestObj", items)
	assert.NoError(err)
	assert.NotNil(jsonstr)
	assert.NotEmpty(jsonstr)

	var iface interface{}
	err = json.Unmarshal([]byte(jsonstr), &iface)
	assert.NoError(err)

	byts, err := json.Marshal(iface)
	assert.NoError(err)
	assert.NotNil(byts)

	actual := string(byts)

	expected :=
		`{"MyTestObj":{"properties":{"bool1":{"type":"boolean"},"date1":{"type":"date"},"double1":{"type":"double"},"integer1":{"type":"integer"},"integer2":{"type":"integer"}}}}`

	assert.Equal(expected, actual)

	err = es.SetMapping("MyTestObj", piazza.JsonString(actual))
	assert.NoError(err)
}

type ByID []*PercolateResponseMatch

func (a ByID) Len() int {
	return len(a)
}
func (a ByID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByID) Less(i, j int) bool {
	return a[i].Id < a[j].Id
}
func sortMatches(matches []*PercolateResponseMatch) []*PercolateResponseMatch {
	sort.Sort(ByID(matches))
	return matches
}

func (suite *EsTester) Test09FullPercolation() {
	t := suite.T()
	assert := assert.New(t)

	var esi IIndex

	var err error

	defer func() {
		closerT(t, esi)
	}()

	// create index
	esi, err = NewIndexInterface(suite.sys, "estest09$", "", true)
	assert.NoError(err)

	// make the index
	err = esi.Create("")
	assert.NoError(err)

	ok, err := esi.IndexExists()
	assert.NoError(err)
	assert.True(ok)
}

func (suite *EsTester) Test10GetAll() {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	required = []piazza.ServiceName{}

	sys, err := piazza.NewSystemConfig(piazza.PzGoCommon, required)
	if err != nil {
		log.Fatal(err)
	}

	esi, err := NewIndexInterface(sys, "getall$", "", true)
	assert.NoError(err)
	defer closerT(t, esi)

	// make the index
	err = esi.Create("")
	assert.NoError(err)

	type T1 struct {
		Data1  string `json:"data1" binding:"required"`
		Extra1 string `json:"extra1" binding:"required"`
	}

	type T2 struct {
		Data2  int    `json:"data2" binding:"required"`
		Extra2 string `json:"extra2" binding:"required"`
	}

	schema1 :=
		`{
			"schema1":{
				"properties":{
					"data1":{
						"type":"string",
						"store":true
					},
					"extra1":{
						"type":"string",
						"store":true
					}
				}
			}
		}`

	schema2 :=
		`{
			"schema2":{
				"properties":{
					"data2":{
						"type":"integer",
						"store":true
					},
					"extra2":{
						"type":"string",
						"store":true
					}
				}
			}
		}`

	err = esi.SetMapping("schema1", piazza.JsonString(schema1))
	assert.NoError(err)
	err = esi.SetMapping("schema2", piazza.JsonString(schema2))
	assert.NoError(err)

	obj1 := T1{Data1: "obj", Extra1: "extra1"}
	obj2 := T2{Data2: 123, Extra2: "extra2"}
	indexResult, err := esi.PostData("schema1", "id1", obj1)
	assert.NoError(err)
	assert.NotNil(indexResult)
	indexResult, err = esi.PostData("schema2", "id2", obj2)
	assert.NoError(err)
	assert.NotNil(indexResult)

	{
		getASpecificOne := func(input1 string, input2 string, tmp interface{}) {
			getResult, err := esi.GetByID(input1, input2)
			assert.NoError(err)
			assert.NotNil(getResult)
			src := getResult.Source
			assert.NotNil(src)
			err = json.Unmarshal(*src, tmp)
			assert.NoError(err)
		}

		tmp1 := T1{}
		getASpecificOne("schema1", "id1", &tmp1)
		assert.EqualValues("obj", tmp1.Data1)

		tmp2 := T2{}
		getASpecificOne("schema2", "id2", &tmp2)
		assert.EqualValues(123, tmp2.Data2)
	}

	{
		// GET the types
		strs, err := esi.GetTypes()
		assert.NoError(err)
		assert.Len(strs, 2)
		if strs[0] == "schema1" {
			assert.EqualValues("schema2", strs[1])
		} else if strs[0] == "schema2" {
			assert.EqualValues("schema1", strs[1])
		} else {
			assert.True(false)
		}
	}

	{
		realFormat := &piazza.JsonPagination{
			PerPage: 10,
			Page:    0,
			Order:   piazza.SortOrderAscending,
			SortBy:  "",
		}

		spf := func() (*SearchResult, error) { return esi.FilterByMatchAll("", realFormat) }

		_, err := PollFunction(searchPoller(spf, 2))
		getResult, err := esi.FilterByMatchAll("", realFormat)
		assert.NoError(err)
		assert.NotNil(getResult)
		assert.Len(*getResult.GetHits(), 2)
		src1 := getResult.GetHit(0).Source
		assert.NotNil(src1)
		src2 := getResult.GetHit(1).Source
		assert.NotNil(src2)

		var tmp1 T1
		var tmp2 T2
		err1 := json.Unmarshal(*src1, &tmp1)
		err2 := json.Unmarshal(*src2, &tmp2)
		assert.True((err1 == nil && err2 == nil) || (err1 != nil && err2 != nil))

		if err1 != nil {
			err = json.Unmarshal(*src1, &tmp1)
			assert.NoError(err)
			err = json.Unmarshal(*src2, &tmp2)
			assert.NoError(err)
		} else {
			err = json.Unmarshal(*src1, &tmp2)
			assert.NoError(err)
			err = json.Unmarshal(*src2, &tmp1)
			assert.NoError(err)
		}

		assert.Equal(tmp1.Data1, "obj")
		assert.Equal(tmp1.Extra1, "extra1")
		assert.Equal(tmp2.Data2, 123)
		assert.Equal(tmp2.Extra2, "extra2")
	}
}

func (suite *EsTester) Test11Pagination1() {
	t := suite.T()
	assert := assert.New(t)

	p := piazza.JsonPagination{
		PerPage: 10,
		Page:    32,
		Order:   piazza.SortOrderDescending,
		SortBy:  "id",
	}

	q := NewQueryFormat(&p)

	assert.Equal(10*32, q.From)
	assert.Equal(10, q.Size)
	assert.False(q.Order)
	assert.EqualValues("id", q.Key)
}

func (suite *EsTester) Test11Pagination2() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	esi := suite.SetUpIndex()
	assert.NotNil(esi)
	defer closerT(t, esi)

	type Obj3 struct {
		ID   string `json:"id3" binding:"required"`
		Data int    `json:"data3" binding:"required"`
	}
	obj3Mapping := `{
	"Obj3":{
		"properties":{
			"id3": {
				"type":"string",
				"store":true
			},
			"data3": {
				"type":"integer",
				"store": true
			}
		}
	}
}`

	err = esi.SetMapping("Obj3", piazza.JsonString(obj3Mapping))
	assert.NoError(err)

	p := fmt.Sprintf("%x", time.Now().Nanosecond()%0xffffffff)

	for i := 0; i <= 9; i++ {
		id := fmt.Sprintf("id%d_%s", i, p)
		obj := Obj3{ID: id, Data: i * i}
		indexResult, err := esi.PostData("Obj3", id, obj)
		assert.NoError(err)
		assert.NotNil(indexResult)
		assert.EqualValues(id, indexResult.ID)
	}

	{
		realFormat := &piazza.JsonPagination{
			PerPage: 4,
			Page:    0,
			Order:   piazza.SortOrderAscending,
			SortBy:  "id3",
		}

		spf := func() (*SearchResult, error) { return esi.FilterByMatchAll("Obj3", realFormat) }

		_, err := PollFunction(searchPoller(spf, 4))
		getResult, err := esi.FilterByMatchAll("Obj3", realFormat)
		assert.NoError(err)
		assert.Len(*getResult.GetHits(), 4)
		assert.Equal("id0_"+p, getResult.GetHit(0).ID)
		assert.Equal("id1_"+p, getResult.GetHit(1).ID)
		assert.Equal("id2_"+p, getResult.GetHit(2).ID)
		assert.Equal("id3_"+p, getResult.GetHit(3).ID)
	}

	{
		realFormat := &piazza.JsonPagination{
			PerPage: 3,
			Page:    1,
			Order:   piazza.SortOrderAscending,
			SortBy:  "",
		}
		getResult, err := esi.FilterByMatchAll("Obj3", realFormat)
		assert.NoError(err)
		assert.Len(*getResult.GetHits(), 3)
		assert.Equal("id3_"+p, getResult.GetHit(0).ID)
		assert.Equal("id4_"+p, getResult.GetHit(1).ID)
		assert.Equal("id5_"+p, getResult.GetHit(2).ID)
	}
}

func (suite *EsTester) Test12TermMatch() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	esi := suite.SetUpIndex()
	assert.NotNil(esi)
	defer closerT(t, esi)

	err = esi.SetMapping(mapping, piazza.JsonString(objMapping))
	assert.NoError(err)

	type NotObj struct {
		ID   int    `json:"id" binding:"required"`
		Data string `json:"data" binding:"required"`
		Foo  bool   `json:"foo" binding:"required"`
	}
	o1 := NotObj{ID: 99, Data: "quick fox", Foo: true}
	o2 := NotObj{ID: 17, Data: "lazy dog", Foo: false}

	indexResult, err := esi.PostData(mapping, "99", o1)
	assert.NoError(err)
	assert.NotNil(indexResult)

	indexResult, err = esi.PostData(mapping, "17", o2)
	assert.NoError(err)
	assert.NotNil(indexResult)

	searchResult, err := esi.FilterByTermQuery(mapping, "data", "lazy dog")
	assert.NoError(err)
	assert.NotNil(searchResult)
	assert.True(searchResult.Found)
	array := searchResult.GetHits()
	assert.Len(*array, 1)

	searchResult, err = esi.FilterByTermQuery(mapping, "data", "lazy sloth")
	assert.NoError(err)
	assert.NotNil(searchResult)
	assert.False(searchResult.Found)
}

func (suite *EsTester) Test13DirectAccess() {
	t := suite.T()
	assert := assert.New(t)

	var err error

	esi := suite.SetUpIndex()
	assert.NotNil(esi)
	defer closerT(t, esi)

	out := &map[string]interface{}{}
	err = esi.DirectAccess("GET", "", nil, out)
	assert.Error(err)
}

func (suite *EsTester) Test14Coverage() {
	t := suite.T()
	assert := assert.New(t)

	indexResponse := NewIndexResponse(&elastic.IndexResponse{
		Index:   "index",
		Type:    "type",
		Id:      "1",
		Version: 1,
		Created: true,
	},
	)
	assert.NotNil(indexResponse)

	percolateResponse := NewPercolateResponse(&elastic.PercolateResponse{
		TookInMillis: 0,
		Total:        1,
		Matches: []*elastic.PercolateMatch{
			&elastic.PercolateMatch{
				Index: "1",
				Id:    "2",
				Score: 3,
			},
		},
	},
	)
	assert.NotNil(percolateResponse)

	getResult := NewGetResult(&elastic.GetResult{Id: "", Source: &json.RawMessage{}, Found: false})
	assert.NotNil(getResult)

	deleteResponse := NewDeleteResponse(&elastic.DeleteResponse{Found: false, Id: ""})
	assert.NotNil(deleteResponse)

	searchResult := NewSearchResult(&elastic.SearchResult{Hits: &elastic.SearchHits{TotalHits: 1, Hits: []*elastic.SearchHit{&elastic.SearchHit{Id: "1", Source: &json.RawMessage{}}}}})
	assert.NotNil(searchResult)
	assert.True(searchResult.TotalHits() == int64(1))
	assert.True(searchResult.NumHits() == 1)

	assert.True(MappingElementTypeString.isValidMappingType())
	assert.True(MappingElementTypeString.isValidScalarMappingType())
	assert.True(MappingElementTypeStringA.isValidArrayMappingType())
	assert.False(MappingElementTypeString.isValidArrayMappingType())
	assert.False(MappingElementTypeStringA.isValidScalarMappingType())
	assert.True(IsValidMappingType("string"))
	assert.True(IsValidArrayTypeMapping("[string]"))
	assert.False(IsValidMappingType(5))
	assert.False(IsValidArrayTypeMapping(5))
}

func (suite *EsTester) Test15NewIndex2() {
	t := suite.T()
	assert := assert.New(t)

	required := []piazza.ServiceName{}
	sys, err := piazza.NewSystemConfig(piazza.PzGoCommon, required)
	assert.NoError(err)
	idx, err := NewIndex(sys, "", "")
	assert.Error(err)

	idx = &Index{
		version: "a",
		index:   "b",
	}
	assert.Equal("a", idx.GetVersion())
	assert.Equal("b", idx.IndexName())

	err = idx.DirectAccess("", "", nil, nil)
	assert.Error(err)

	_, err = NewIndex2("", "$", "")
	assert.Error(err)
	_, err = NewIndexInterface(sys, "", "", false)
	assert.Error(err)
}

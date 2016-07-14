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
	"net/http"

	"github.com/gin-gonic/gin"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	pzlogger "github.com/venicegeo/pz-logger/logger"
)

//--------------------------------------------------

type Server struct {
	logger  pzlogger.IClient
	Routes  []piazza.RouteData
	service *Service
}

//--------------------------------------------------

func (server *Server) Init(service *Service) {
	server.Routes = []piazza.RouteData{
		{"GET", "/", server.handleGetRoot},
		{"GET", "/admin/stats", server.handleGetStats},
		{"POST", "/uuids", server.handlePostUuids},
	}
	server.service = service
}

func (server *Server) handleGetRoot(c *gin.Context) {
	message := "Hi. I'm pz-uuidgen."
	resp := piazza.JsonResponse{StatusCode: http.StatusOK, Data: message}
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *Server) handleGetStats(c *gin.Context) {
	resp := server.service.GetStats()
	c.IndentedJSON(resp.StatusCode, resp)
}

// request body is ignored
// we allow a count of zero, for testing
func (server *Server) handlePostUuids(c *gin.Context) {
	params := piazza.NewQueryParams(c.Request)
	resp := server.service.PostUuids(params)
	c.IndentedJSON(resp.StatusCode, resp)
}
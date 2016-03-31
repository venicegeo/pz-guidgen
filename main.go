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

package main

import (
	//"github.com/pborman/uuid"
	"log"

	piazza "github.com/venicegeo/pz-gocommon"
	loggerPkg "github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-uuidgen/server"
)

func main() {

	required := []piazza.ServiceName{
		piazza.PzElasticSearch,
		piazza.PzLogger,
	}

	sys, err := piazza.NewSystemConfig(piazza.PzUuidgen, required, false)
	if err != nil {
		log.Fatal(err)
	}

	loggerService, err := loggerPkg.NewPzLoggerService(sys)
	if err != nil {
		log.Fatal(err)
	}

	done := sys.StartServer(server.CreateHandlers(sys, loggerService))

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}

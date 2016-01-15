package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	piazza "github.com/venicegeo/pz-gocommon"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var debugMode = false
var debugCounter = 0

var numRequests = 0
var numUUIDs = 0

var startTime = time.Now()

func handleAdminGet(w http.ResponseWriter, r *http.Request) {

	uuidgen := piazza.AdminResponse_UuidGen{NumRequests: numRequests, NumUUIDs: numUUIDs}
	m := piazza.AdminResponse{StartTime: startTime, UuidGen: &uuidgen}

	data, err := json.Marshal(m)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// request body is ignored
// we allow a count of zero, for testing
func handleUUIDService(w http.ResponseWriter, r *http.Request) {

	var count int
	var err error

	key := r.URL.Query().Get("count")
	if key == "" {
		count = 1
	} else {
		count, err = strconv.Atoi(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("query argument invalid: %s", key), http.StatusBadRequest)
			return
		}
	}

	if count < 0 || count > 255 {
		http.Error(w, fmt.Sprintf("query argument out of range: %d", count), http.StatusBadRequest)
		return
	}

	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		if debugMode {
			uuids[i] = fmt.Sprintf("%d", debugCounter)
			debugCounter++
		} else {
			uuids[i] = uuid.New()
		}
	}

	data := make(map[string]interface{})
	data["data"] = uuids

	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	numUUIDs += count
	numRequests++

	piazza.SendLogMessage("uuidgen", "0.0.0.0", piazza.SeverityInfo, "uuid generator")

	w.Write(bytes)
}

func runUUIDServer(host string, port string, debug bool) error {

	debugMode = debug

	r := mux.NewRouter()
	r.HandleFunc("/uuid/admin", handleAdminGet).
		Methods("GET")
	r.HandleFunc("/uuid", handleUUIDService).
		Methods("POST")

	server := &http.Server{Addr: host + ":" + port, Handler: piazza.HttpLogHandler(r)}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// not reached
	return nil
}

func app() int {
	var host = flag.String("host", "localhost", "host name")
	var port = flag.String("port", "12340", "port number")
	var debug = flag.Bool("debug", false, "use debug mode")

	flag.Parse()

	log.Printf("starting: host=%s, port=%s, debug=%t", *host, *port, *debug)

	err := runUUIDServer(*host, *port, *debug)
	if err != nil {
		fmt.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app()
}

func main() {
	os.Exit(app())
}

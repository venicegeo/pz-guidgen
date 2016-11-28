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

package syslog

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"strconv"

	"github.com/jeromer/syslogparser/rfc5424"
)

//---------------------------------------------------------------------

const privateEnterpriseNumber = "48851" // Flaxen's PEN
const DefaultFacility = 1
const DefaultVersion = 1

type Severity int

func (s Severity) Value() int { return int(s) }

const (
	Emergency     Severity = 0 // not used by Piazza
	Alert         Severity = 1 // not used by Piazza
	Fatal         Severity = 2 // called Critical in the spec
	Error         Severity = 3
	Warning       Severity = 4
	Notice        Severity = 5
	Informational Severity = 6
	Debug         Severity = 7
)

// Message represents all the fields of a native RFC5424 object, plus
// our own two SDEs.
type Message struct {
	Facility    int            `json:"facility"`
	Severity    Severity       `json:"severity"`
	Version     int            `json:"version"`
	TimeStamp   time.Time      `json:"timeStamp"`
	HostName    string         `json:"hostName"`
	Application string         `json:"application"`
	Process     string         `json:"process"`
	MessageID   string         `json:"messageId"`
	AuditData   *AuditElement  `json:"auditData"`
	MetricData  *MetricElement `json:"metricData"`
	SourceData  *SourceElement `json:"sourceData"`
	Message     string         `json:"message"`
}

// NewMessage returns a Message with the defaults filled in for you.
func NewMessage() *Message {
	var err error

	host, err := os.Hostname()
	if err != nil {
		host = "-"
	}
	host += " "

	m := &Message{
		Facility:    DefaultFacility,
		Severity:    Informational,
		Version:     DefaultVersion,
		TimeStamp:   time.Now().Round(time.Millisecond),
		HostName:    host,
		Application: "",
		Process:     strconv.Itoa(os.Getpid()),
		MessageID:   "",
		AuditData:   nil,
		MetricData:  nil,
		SourceData:  nil,
		Message:     "",
	}

	return m
}

// String builds and returns the RFC5424-style textual representation of a Message.
func (m *Message) String() string {
	pri := m.Facility*8 + m.Severity.Value()

	timestamp := m.TimeStamp.Format(time.RFC3339)

	host := m.HostName
	if host == "" {
		host = "-"
	}

	application := m.Application
	if application == "" {
		application = "-"
	}

	proc := m.Process
	if proc == "" {
		proc = "-"
	}

	messageID := m.MessageID
	if messageID == "" {
		messageID = "-"
	}

	header := fmt.Sprintf("<%d>%d %s %s %s %s %s",
		pri, m.Version, timestamp, host,
		application, proc, messageID)

	sdes := []string{}
	if m.AuditData != nil {
		sdes = append(sdes, m.AuditData.String())
	}
	if m.MetricData != nil {
		sdes = append(sdes, m.MetricData.String())
	}
	if m.SourceData != nil {
		sdes = append(sdes, m.SourceData.String())
	}
	sde := strings.Join(sdes, " ")
	if sde == "" {
		sde = "-"
	}

	mssg := m.Message

	s := fmt.Sprintf("%s %s %s", header, sde, mssg)
	return s
}

func ParseMessageString(s string) (*Message, error) {
	m := &Message{}

	buff := []byte(s)
	p := rfc5424.NewParser(buff)
	err := p.Parse()
	if err != nil {
		return nil, err
	}

	parts := p.Dump()
	m.Facility = parts["facility"].(int)
	m.Severity = Severity(parts["severity"].(int))
	m.Version = parts["version"].(int)
	m.TimeStamp = parts["timestamp"].(time.Time)
	m.HostName = parts["hostname"].(string)
	m.Application = parts["app_name"].(string)
	m.Process = parts["proc_id"].(string)
	m.MessageID = parts["msg_id"].(string)
	m.Message = parts["message"].(string)

	//sdes := parts["structured_data"].(string)
	//log.Printf("SDES: %s", sdes)

	return m, nil
}

// IsSecurityAudit returns true iff the audit action is something we need to formally
// record as an auidtable event.
func (m *Message) IsSecurityAudit() bool {
	if m.AuditData == nil {
		return false
	}

	for _, s := range securityAuditActions {
		if m.AuditData.Action == s {
			return true
		}
	}
	return false
}

func (m *Message) validate() error {
	if m.Facility != DefaultFacility {
		return fmt.Errorf("Invalid Message.Facility: %d", m.Facility)
	}
	if m.Severity < Emergency || m.Severity > Debug {
		return fmt.Errorf("Invalid Message.Severity: %d", m.Severity)
	}
	if m.Version != DefaultVersion {
		return fmt.Errorf("Invalid Message.Version: %d", m.Version)
	}
	// nothing to check for m.TimeStamp
	if m.HostName == "" {
		return fmt.Errorf("Message.HostnName not set")
	}

	if m.Application == "" {
		return fmt.Errorf("Message.Application not set")
	}

	if m.Process == "" {
		return fmt.Errorf("Message.Process not set")
	}

	return nil
}

// Validate checks to see if a Message is well-formed.
func (m *Message) Validate() error {
	var err error

	err = m.validate()
	if err != nil {
		return err
	}

	if m.AuditData != nil {
		err = m.AuditData.validate()
		if err != nil {
			return err
		}
	}

	if m.MetricData != nil {
		err = m.MetricData.validate()
		if err != nil {
			return err
		}
	}

	if m.SourceData != nil {
		err = m.SourceData.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

//---------------------------------------------------------------------

// AuditElement represents an SDE for auditing (security-specific of just general).
type AuditElement struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Actee  string `json:"actee"`
}

// TODO: fill these in
var securityAuditActions = []string{
	"create",
	"read",
	"update",
	"delete",
}

func (ae *AuditElement) validate() error {
	if ae.Actor == "" {
		return fmt.Errorf("AuditElement.Actor not set")
	}
	if ae.Action == "" {
		return fmt.Errorf("AuditElement.Action not set")
	}
	if ae.Actee == "" {
		return fmt.Errorf("AuditElement.Actee not set")
	}

	// TODO: check for valid UUIDs?

	return nil
}

// String builds and returns the RFC5424-style textual representation of an Audit SDE
func (ae *AuditElement) String() string {
	s := fmt.Sprintf("[pzaudit@%s Actor=\"%s\" Action=\"%s\" Actee=\"%s\"]",
		privateEnterpriseNumber, ae.Actor, ae.Action, ae.Actee)
	return s
}

//---------------------------------------------------------------------

// MetricElement represents an SDE for recoridng metrics.
type MetricElement struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Object string  `json:"object"`
}

func (me *MetricElement) validate() error {
	if me.Name == "" {
		return fmt.Errorf("MetricElement.Name not set")
	}
	if me.Object == "" {
		return fmt.Errorf("MetricElement.Object not set")
	}

	// TODO: check for valid UUIDs?

	return nil
}

// String builds and returns the RFC5424-style textual representation of an Metric SDE
func (me *MetricElement) String() string {
	s := fmt.Sprintf("[pzmetric@%s Name=\"%s\" Value=\"%f\" Object=\"%s\"]",
		privateEnterpriseNumber, me.Name, me.Value, me.Object)
	return s
}

//---------------------------------------------------------------------

// SourceElement represents an SDE for tracking the message back to the source code.
type SourceElement struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

func NewSourceElement(skip int) *SourceElement {
	function, file, line := stackFrame(skip)
	se := &SourceElement{
		File:     file,
		Function: function,
		Line:     line,
	}
	return se
}

// stackFrame returns info about the requested stack frame. If skip==0,
// info about the caller of stackFrame is returned. If skip==1, info
// about the caller of the caller of stackFrame is returned.
func stackFrame(skip int) (function string, file string, line int) {

	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0
	}

	fnc := runtime.FuncForPC(pc)
	function = fnc.Name()

	return path.Base(function), path.Base(file), line
}

func (se *SourceElement) validate() error {
	if se.Function == "" {
		return fmt.Errorf("SourceElement.Function not set")
	}
	if se.File == "" {
		return fmt.Errorf("SourceElement.File not set")
	}
	if se.Line < 0 || se.Line > 10000 {
		return fmt.Errorf("SourceElement.Line is invalid")
	}

	return nil
}

// String builds the text string of the SDE
func (me *SourceElement) String() string {
	s := fmt.Sprintf("[pzsource@%s File=\"%s\" Function=\"%s\" Line=\"%d\"]",
		privateEnterpriseNumber, me.File, me.Function, me.Line)
	return s
}

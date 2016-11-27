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
	"strings"
	"time"

	"strconv"

	"github.com/jeromer/syslogparser/rfc5424"
)

//---------------------------------------------------------------------

type Severity int

func (s Severity) Value() int { return int(s) }

const (
	Emergency     Severity = 0	// not used by Piazza
	Alert         Severity = 1	// not used by Piazza
	Fatal         Severity = 2 	// called Critical in the spec
	Error         Severity = 3
	Warning       Severity = 4
	Notice        Severity = 5
	Informational Severity = 6
	Debug         Severity = 7
)

const DefaultFacility = 1
const DefaultVersion = 1

// SyslogMessage represents all the fields of a native RFC5424 object, plus
// our own two SDEs.
type SyslogMessage struct {
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
	Message     string         `json:"message"`
}

// NewSyslogMessage returns a SyslogMessage with the defaults filled in for you.
func NewSyslogMessage() *SyslogMessage {
	var err error

	host, err := os.Hostname()
	if err != nil {
		host = "-"
	}
	host += " "

	m := &SyslogMessage{
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
		Message:     "",
	}

	return m
}

// String builds and returns the RFC5424-style textual representation of a SyslogMessage.
func (m *SyslogMessage) String() string {
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
	sde := strings.Join(sdes, " ")
	if sde == "" {
		sde = "-"
	}

	mssg := m.Message

	s := fmt.Sprintf("%s %s %s", header, sde, mssg)
	return s
}

func ParseSyslogMessage(s string) (*SyslogMessage, error) {
	m := &SyslogMessage{}

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
func (m *SyslogMessage) IsSecurityAudit() bool {
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

func (m *SyslogMessage) validate() error {
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

//---------------------------------------------------------------------

const privateEnterpriseNumber = "48851" // Flaxen's PEN

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

// MetricElement represents an SDE for recoridng metrics.
type MetricElement struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Object string  `json:"object"`
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

// Validate checks to see if a SyslogMessage is well-formed.
func (m *SyslogMessage) Validate() error {
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

	return nil
}

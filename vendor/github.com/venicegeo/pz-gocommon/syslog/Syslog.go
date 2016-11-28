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

//---------------------------------------------------------------------

// Logger is the "helper" class that can (should) be used by services to send messages.
// In most Piazza cases, the Writer field should be set to an ElkWriter.
type Logger struct {
	writer           WriterI
	MinimumSeverity  Severity // minimum severity level you want to record
	UseSourceElement bool
}

func NewLogger(writer WriterI) *Logger {
	logger := &Logger{
		writer:           writer,
		MinimumSeverity:  Informational,
		UseSourceElement: false,
	}
	return logger
}

func (logger *Logger) severityAllowed(desiredSeverity Severity) bool {
	return logger.MinimumSeverity.Value() >= desiredSeverity.Value()
}

// postMessage sends a log message
func (logger *Logger) postMessage(severity Severity, text string) {
	if !logger.severityAllowed(severity) {
		return
	}

	mssg := NewMessage()
	mssg.Message = text
	mssg.Severity = severity

	if logger.UseSourceElement {
		// -1: stackFrame
		// 0: NewSourceElement
		// 1: postMessage
		// 2: Fatal
		// 3: actual source
		const skip = 3
		mssg.SourceData = NewSourceElement(skip)
	}

	logger.writer.Write(mssg)
}

// Debug sends a log message with severity "Debug".
func (logger *Logger) Debug(text string) {
	logger.postMessage(Debug, text)
}

// Info sends a log message with severity "Informational".
func (logger *Logger) Info(text string) {
	logger.postMessage(Informational, text)
}

// Notice sends a log message with severity "Notice".
func (logger *Logger) Notice(text string) {
	logger.postMessage(Notice, text)
}

// Warning sends a log message with severity "Warning".
func (logger *Logger) Warning(text string) {
	logger.postMessage(Warning, text)
}

// Error sends a log message with severity "Error".
func (logger *Logger) Error(text string) {
	logger.postMessage(Error, text)
}

// Fatal sends a log message with severity "Fatal".
func (logger *Logger) Fatal(text string) {
	logger.postMessage(Fatal, text)
}

/*
Copyright © 2021 Upvest GmbH <support@upvest.co>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"net/http"
)

var HttpProxyNoLogging = http.CanonicalHeaderKey("X-HTTP-PROXY-NO-LOGGING")

type Logger interface {
	Log(message string)
	LogF(format string, a ...interface{})

	PrintF(format string, a ...interface{})
	Print(message string)
	PrintLn(message string)
}

var NoVerboseLogger = New(false)

func New(verboseMode bool) *ConsoleLogger {
	return &ConsoleLogger{verboseMode: verboseMode}
}

type ConsoleLogger struct {
	verboseMode bool
}

func (l *ConsoleLogger) PrintF(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func (l *ConsoleLogger) Print(message string) {
	fmt.Print(message)
}

func (l *ConsoleLogger) PrintLn(message string) {
	fmt.Println(message)
}

func (l *ConsoleLogger) LogF(format string, a ...interface{}) {
	if l.verboseMode {
		fmt.Printf(format+"\n", a...)
	}
}

func (l *ConsoleLogger) Log(message string) {
	if l.verboseMode {
		fmt.Print(message + "\n")
	}
}

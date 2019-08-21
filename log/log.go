// Package log provides leveled logging formatted for systemd
// daemon logs.
package log

import (
	"fmt"
	"strings"
)

// Level is the systemd log level, 0-7.
type Level uint8

// Log levels used in systemd, 0-7.
const (
	// level 0, "EMERG"
	Emergency Level = iota
	// level 1, "ALERT"
	Alert
	// level 2, "CRIT"
	Critical
	// level 3, "ERR"
	Error
	// level 4, "WARNING"
	Warning
	// level 5, "NOTICE"
	Notice
	// level 6, "INFO"
	Info
	// level 7, "DEBUG"
	Debug
)

var lvlStrings = [...]string{
	"EMERG",
	"ALERT",
	"CRIT",
	"ERR",
	"WARNING",
	"NOTICE",
	"INFO",
	"DEBUG"}

const sysdfmt = "<%d>%v"

var prefixWithLevel bool

// SetPrefixWithLevel will enable or disable prefixing the log message
// with the string representation of the log level. Default is disabled.
func SetPrefixWithLevel(enable bool) {
	prefixWithLevel = enable
}

// Printf logs v using the level and format given. The format string should
// not include newlines, as these will be replaced. (Systemd logs are a
// single line.)
func Printf(lvl Level, format string, v ...interface{}) {
	if prefixWithLevel {
		format = fmt.Sprintf("%s %s", lvlStrings[lvl], format)
	}

	msg := fmt.Sprintf(fmt.Sprintf(sysdfmt, lvl, format), v...)
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\n", "|") // remove newlines
	fmt.Println(msg)
}

// Print does the same as Printf, but uses the standard formatting
// for each value in v.
func Print(lvl Level, v ...interface{}) {
	fmt := "%v"
	for i := 1; i < len(v); i++ {
		fmt += " %v"
	}
	Printf(lvl, fmt, v...)
}
